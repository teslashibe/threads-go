package threads

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// readBaseURL is the host used for cookie-authenticated read endpoints.
	readBaseURL = "https://www.threads.com"

	// writeBaseURL is the host used for Bearer-authenticated write endpoints.
	// Threads writes share Instagram's REST backend.
	writeBaseURL = "https://i.instagram.com"

	// appID is the X-IG-App-ID header value used by both the Threads web
	// app and the Threads mobile app. This identifies the calling app.
	appID = "238260118697367"

	// blocksVersioningID is the current Bloks versioning hash for the Threads
	// app. Used during the Bloks login flow. May change with app updates.
	blocksVersioningID = "00ba6fa565c3c707243ad976fa30a071a625f2a3d158d9412091176fe35027d8"

	// defaultReadUserAgent is the User-Agent for cookie read requests.
	// The Threads backend rejects desktop browser UAs with "useragent mismatch".
	defaultReadUserAgent = "Barcelona 289.0.0.14.109 Android"

	// defaultWriteUserAgent is the User-Agent for Bearer write requests.
	defaultWriteUserAgent = "Barcelona 289.0.0.77.109 Android"

	defaultMinGap     = 1500 * time.Millisecond
	defaultMaxRetries = 3
	defaultRetryBase  = 750 * time.Millisecond
	maxResponseBody   = 32 << 20 // 32 MB

	// defaultReadCooldown / defaultWriteCooldown are the durations the
	// circuit-breaker holds off new requests after observing a rate-limit
	// signal. Threads/Instagram's soft-block typically clears within 5–10 min
	// for reads and 15+ min for writes.
	defaultReadCooldown  = 5 * time.Minute
	defaultWriteCooldown = 15 * time.Minute
)

// Client is a Threads API client. It is safe for concurrent use.
//
// A Client may be created with cookies (read access only), Bearer auth
// (write access only), or both (full access). The constructors are:
//
//   - New(Cookies, ...Option)            — cookie auth, reads only
//   - NewWithAuth(Auth, ...Option)       — Bearer auth, writes only
//   - NewFull(Cookies, Auth, ...Option)  — both, full read/write
//
// Methods that require Bearer auth will return ErrWriteAuthRequired if
// called on a cookie-only client.
type Client struct {
	cookies    Cookies
	auth       Auth
	hasCookies bool
	hasBearer  bool

	httpClient *http.Client

	readUA  string
	writeUA string

	maxRetries int
	retryBase  time.Duration

	minGap    time.Duration
	gapMu     sync.Mutex
	lastReqAt time.Time

	readCooldown  time.Duration
	writeCooldown time.Duration
	rlMu          sync.Mutex
	rlState       RateLimitState
	cooldownRead  time.Time
	cooldownWrt   time.Time

	viewer   *User
	viewerMu sync.RWMutex

	// auth (the Bearer token set) is mutable: it can be re-minted on a
	// write-auth failure when credentials are configured. authMu guards it.
	authMu sync.RWMutex

	// creds, when set via WithCredentials, lets the client auto re-login
	// (re-mint the Bearer token) when a write returns an auth failure.
	creds    Credentials
	hasCreds bool
	reauthMu sync.Mutex // serialises re-login so only one runs at a time

	// onTokenRefresh, when set, is invoked with the freshly minted Auth
	// after a successful auto re-login so the caller can persist it.
	onTokenRefresh func(Auth)
}

// Credentials hold the Instagram login used to mint (and auto-refresh) the
// Bearer token via the Bloks login flow. TOTPSecret is the optional base32
// authenticator-app secret used to auto-solve app-based two-factor auth.
type Credentials struct {
	Username   string
	Password   string
	DeviceID   string
	TOTPSecret string
}

// WithCredentials configures the client to auto re-login (re-mint its Bearer
// token) when a write operation fails with an auth error. Pair with
// WithTokenRefreshCallback to persist the refreshed token.
func WithCredentials(creds Credentials) Option {
	return func(c *Client) {
		c.creds = creds
		c.hasCreds = creds.Username != "" && creds.Password != ""
	}
}

// WithTokenRefreshCallback registers a callback invoked with the new Auth
// each time the client successfully re-mints its Bearer token, so the caller
// can persist the updated token / device id.
func WithTokenRefreshCallback(fn func(Auth)) Option {
	return func(c *Client) { c.onTokenRefresh = fn }
}

// getAuth returns a copy of the current Bearer auth under lock.
func (c *Client) getAuth() Auth {
	c.authMu.RLock()
	defer c.authMu.RUnlock()
	return c.auth
}

// reauth re-mints the Bearer token from stored credentials and updates the
// client in place. Serialised so concurrent writers trigger at most one
// re-login. Returns ErrUnauthorized when no credentials are configured.
func (c *Client) reauth(ctx context.Context) error {
	if !c.hasCreds {
		return ErrUnauthorized
	}
	c.reauthMu.Lock()
	defer c.reauthMu.Unlock()

	res, err := c.performLogin(ctx, LoginParams{
		Username:   c.creds.Username,
		Password:   c.creds.Password,
		DeviceID:   firstNonEmptyStr(c.creds.DeviceID, c.getAuth().DeviceID),
		TOTPSecret: c.creds.TOTPSecret,
	})
	if err != nil {
		return err
	}
	newAuth := Auth{Token: res.Token, UserID: res.UserID, DeviceID: res.DeviceID}
	c.authMu.Lock()
	c.auth = newAuth
	c.hasBearer = true
	c.authMu.Unlock()
	if c.onTokenRefresh != nil {
		c.onTokenRefresh(newAuth)
	}
	return nil
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Option configures a Client.
type Option func(*Client)

// WithUserAgent overrides the default read and write User-Agents to the
// same value. Use WithReadUserAgent / WithWriteUserAgent to set them
// independently.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.readUA = ua
		c.writeUA = ua
	}
}

// WithReadUserAgent overrides the User-Agent used for cookie read calls.
// Threads rejects desktop browser UAs; stick to "Barcelona ..." or
// "Instagram ... Android" forms.
func WithReadUserAgent(ua string) Option {
	return func(c *Client) { c.readUA = ua }
}

// WithWriteUserAgent overrides the User-Agent used for Bearer write calls.
func WithWriteUserAgent(ua string) Option {
	return func(c *Client) { c.writeUA = ua }
}

// WithRetry configures retry behaviour for transient errors.
// Default: 3 attempts, 750ms exponential base.
func WithRetry(maxAttempts int, base time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxAttempts
		c.retryBase = base
	}
}

// WithHTTPClient replaces the default http.Client. Useful for plugging
// in a custom transport or proxy. Nil is ignored.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithProxy routes all HTTP traffic through the given proxy URL.
func WithProxy(proxyURL string) Option {
	return func(c *Client) {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = http.ProxyURL(parsed)
		c.httpClient = &http.Client{
			Timeout:   c.httpClient.Timeout,
			Transport: transport,
		}
	}
}

// WithMinRequestGap sets the minimum time between consecutive HTTP
// requests. Default: 1.5s. Lower values risk triggering Threads' aggressive
// behavioral rate limiter, which suspends sessions for ~15-30 minutes.
func WithMinRequestGap(d time.Duration) Option {
	return func(c *Client) { c.minGap = d }
}

// New creates a Client authenticated for read operations only (cookie auth).
// Returns ErrInvalidAuth if SessionID or CSRFToken is empty. The session is
// validated by calling /accounts/current_user/ which warms the Me() cache.
func New(cookies Cookies, opts ...Option) (*Client, error) {
	if cookies.SessionID == "" || cookies.CSRFToken == "" {
		return nil, fmt.Errorf("%w: SessionID and CSRFToken must both be non-empty", ErrInvalidAuth)
	}

	c := newBaseClient(opts...)
	c.cookies = cookies
	c.hasCookies = true

	if err := c.validateSession(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

// NewWithAuth creates a Client authenticated for write operations only
// (Bearer token from a Bloks login). Use this when you only need to post,
// like, follow, etc., and don't have browser cookies. Read operations on
// such a client will return ErrInvalidAuth.
func NewWithAuth(auth Auth, opts ...Option) (*Client, error) {
	if auth.Token == "" || auth.UserID == "" {
		return nil, fmt.Errorf("%w: Token and UserID must both be non-empty", ErrInvalidAuth)
	}
	c := newBaseClient(opts...)
	c.auth = auth
	c.hasBearer = true
	return c, nil
}

// NewFull creates a Client authenticated for both read and write operations.
// Equivalent to combining New + NewWithAuth on a single client.
func NewFull(cookies Cookies, auth Auth, opts ...Option) (*Client, error) {
	if cookies.SessionID == "" || cookies.CSRFToken == "" {
		return nil, fmt.Errorf("%w: SessionID and CSRFToken must both be non-empty", ErrInvalidAuth)
	}
	if auth.Token == "" || auth.UserID == "" {
		return nil, fmt.Errorf("%w: Token and UserID must both be non-empty", ErrInvalidAuth)
	}
	c := newBaseClient(opts...)
	c.cookies = cookies
	c.auth = auth
	c.hasCookies = true
	c.hasBearer = true

	if err := c.validateSession(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

func newBaseClient(opts ...Option) *Client {
	c := &Client{
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		readUA:        defaultReadUserAgent,
		writeUA:       defaultWriteUserAgent,
		maxRetries:    defaultMaxRetries,
		retryBase:     defaultRetryBase,
		minGap:        defaultMinGap,
		readCooldown:  defaultReadCooldown,
		writeCooldown: defaultWriteCooldown,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Me returns the authenticated user's profile, populated from the session
// validation call made at construction time.
func (c *Client) Me(ctx context.Context) (*User, error) {
	c.viewerMu.RLock()
	v := c.viewer
	c.viewerMu.RUnlock()
	if v != nil {
		return v, nil
	}
	if !c.hasCookies {
		return nil, ErrUnauthorized
	}
	return c.refreshMe(ctx)
}

// refreshMe re-fetches /accounts/current_user/ and caches the result.
func (c *Client) refreshMe(ctx context.Context) (*User, error) {
	body, err := c.readGET(ctx, "/api/v1/accounts/current_user/", url.Values{"edit": []string{"true"}})
	if err != nil {
		return nil, err
	}
	var resp struct {
		User userPayload `json:"user"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: decoding current_user: %v", ErrRequestFailed, err)
	}
	u := toUser(resp.User)
	c.viewerMu.Lock()
	c.viewer = &u
	c.viewerMu.Unlock()
	return &u, nil
}

// validateSession is called during construction to ensure the cookies are
// live before returning a Client.
func (c *Client) validateSession(ctx context.Context) error {
	_, err := c.refreshMe(ctx)
	return err
}

// ----------------------------------------------------------------------------
// Read transport (cookie auth, www.threads.com)
// ----------------------------------------------------------------------------

func (c *Client) readGET(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	if !c.hasCookies {
		return nil, fmt.Errorf("%w: read requires cookie auth (use New)", ErrUnauthorized)
	}
	endpoint := readBaseURL + path
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doWithRetry(ctx, false, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		c.setReadHeaders(req)
		return req, nil
	})
}

func (c *Client) readPOSTForm(ctx context.Context, path string, form url.Values) (json.RawMessage, error) {
	if !c.hasCookies {
		return nil, fmt.Errorf("%w: read requires cookie auth", ErrUnauthorized)
	}
	endpoint := readBaseURL + path
	body := form.Encode()
	return c.doWithRetry(ctx, false, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		c.setReadHeaders(req)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
}

func (c *Client) setReadHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.readUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-IG-App-ID", appID)
	req.Header.Set("X-CSRFToken", c.cookies.CSRFToken)
	req.Header.Set("X-ASBD-ID", "129477")
	req.Header.Set("X-IG-WWW-Claim", "0")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Referer", readBaseURL+"/")
	req.Header.Set("Cookie", c.readCookieHeader())
}

func (c *Client) readCookieHeader() string {
	var b strings.Builder
	add := func(name, val string) {
		if val == "" {
			return
		}
		if b.Len() > 0 {
			b.WriteString("; ")
		}
		b.WriteString(name)
		b.WriteByte('=')
		b.WriteString(val)
	}
	add("ig_did", c.cookies.IgDid)
	add("mid", c.cookies.Mid)
	add("csrftoken", c.cookies.CSRFToken)
	add("ds_user_id", c.cookies.DSUserID)
	add("sessionid", c.cookies.SessionID)
	return b.String()
}

// ----------------------------------------------------------------------------
// Write transport (Bearer auth, i.instagram.com)
// ----------------------------------------------------------------------------

func (c *Client) writePOSTForm(ctx context.Context, path string, form url.Values) (json.RawMessage, error) {
	if !c.hasBearer {
		return nil, ErrWriteAuthRequired
	}
	endpoint := writeBaseURL + path
	body := form.Encode()
	return c.doWithRetry(ctx, true, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		c.setWriteHeaders(req)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		return req, nil
	})
}

// writeSignedPOST performs a write with a signed_body envelope. The payload
// is JSON-encoded, URL-escaped, and prefixed with the literal "SIGNATURE."
// (no real cryptographic signing is required for this flow).
func (c *Client) writeSignedPOST(ctx context.Context, path string, payload map[string]interface{}) (json.RawMessage, error) {
	if !c.hasBearer {
		return nil, ErrWriteAuthRequired
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	auth := c.getAuth()
	if _, ok := payload["_uid"]; !ok {
		payload["_uid"] = auth.UserID
	}
	if _, ok := payload["device_id"]; !ok && auth.DeviceID != "" {
		payload["device_id"] = auth.DeviceID
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: marshalling signed_body: %v", ErrRequestFailed, err)
	}
	form := url.Values{}
	form.Set("signed_body", "SIGNATURE."+url.QueryEscape(string(jsonBytes)))
	return c.writePOSTForm(ctx, path, form)
}

// writeGET performs a Bearer-authenticated GET (used for some read endpoints
// that only work via i.instagram.com — feed_timeline, news/inbox, etc.).
func (c *Client) writeGET(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	if !c.hasBearer {
		return nil, ErrWriteAuthRequired
	}
	endpoint := writeBaseURL + path
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	return c.doWithRetry(ctx, true, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		c.setWriteHeaders(req)
		return req, nil
	})
}

func (c *Client) setWriteHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.writeUA)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-IG-App-ID", appID)
	req.Header.Set("X-IG-Capabilities", "3brTvx0=")
	req.Header.Set("X-IG-Connection-Type", "WIFI")
	if tok := c.getAuth().Token; tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
}

// ----------------------------------------------------------------------------
// Retry / pacing / response handling
// ----------------------------------------------------------------------------

// doWithRetry executes the request returned by build, applying minimum
// request gap pacing and exponential backoff on transient errors.
// isWrite must be true for Bearer write calls, false for cookie read calls.
func (c *Client) doWithRetry(ctx context.Context, isWrite bool, build func() (*http.Request, error)) (json.RawMessage, error) {
	attempts := c.maxRetries
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	reauthed := false
	for i := 0; i < attempts; i++ {
		if i > 0 {
			wait := c.retryBase * time.Duration(math.Pow(2, float64(i-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		// Honour any active cooldown before pacing — this is the circuit-breaker.
		if err := c.waitForCooldownInternal(ctx, isWrite); err != nil {
			return nil, err
		}
		c.waitForGap(ctx)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := build()
		if err != nil {
			return nil, fmt.Errorf("%w: building request: %v", ErrRequestFailed, err)
		}
		body, err := c.do(req, isWrite)
		if err == nil {
			return body, nil
		}
		// On a write auth failure, re-mint the Bearer token once from stored
		// credentials and retry. Subsequent build() picks up the new token via
		// setWriteHeaders/getAuth.
		if isWrite && c.hasCreds && !reauthed &&
			(errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrSessionSuspended)) {
			reauthed = true
			if rerr := c.reauth(ctx); rerr == nil {
				i-- // don't consume a retry slot for the re-login round-trip
				continue
			}
			// Re-login failed: surface the original auth error.
			return nil, err
		}
		if isNonRetriable(err) {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (c *Client) do(req *http.Request, isWrite bool) (json.RawMessage, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if readErr != nil {
		return nil, fmt.Errorf("%w: reading body: %v", ErrRequestFailed, readErr)
	}

	// Capture soft capacity / timing signals regardless of status.
	c.recordSignals(resp, isWrite)

	if err := mapStatus(resp.StatusCode, body); err != nil {
		// 429 → trip cooldown so the retry loop (or next caller) waits.
		if errors.Is(err, ErrRateLimited) {
			c.tripCooldown(isWrite, c.cooldownFor(isWrite), fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
		return nil, err
	}

	if err := mapFailStatus(resp.StatusCode, body); err != nil {
		// "Please wait a few minutes" → also a rate-limit signal.
		if fe, ok := err.(*FailStatusError); ok {
			msg := strings.ToLower(fe.Message)
			if strings.Contains(msg, "wait") || strings.Contains(msg, "few minutes") {
				cd := c.cooldownFor(isWrite)
				c.tripCooldown(isWrite, cd, "wait-a-few-minutes: "+fe.Message)
				return nil, fmt.Errorf("%w: cooldown for %s (%s)", ErrRateLimited, cd, fe.Message)
			}
		}
		return nil, err
	}

	return body, nil
}

// mapStatus translates HTTP status codes to sentinel errors. Body is
// inspected when a 403 may be a temporary suspension vs. a real forbid.
func mapStatus(code int, body []byte) error {
	if code >= 200 && code < 300 {
		return nil
	}
	switch {
	case code == http.StatusUnauthorized:
		return ErrUnauthorized
	case code == http.StatusForbidden:
		// Try to detect logout_reason 8 (temporary suspension)
		var f struct {
			Message      string `json:"message"`
			LogoutReason int    `json:"logout_reason"`
		}
		_ = json.Unmarshal(body, &f)
		if f.LogoutReason == 8 || f.Message == "login_required" {
			return ErrSessionSuspended
		}
		return ErrForbidden
	case code == http.StatusNotFound:
		return ErrNotFound
	case code == http.StatusTooManyRequests:
		return ErrRateLimited
	case code == 400:
		// Try to surface "useragent mismatch" as a typed error.
		var f struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(body, &f)
		if strings.Contains(strings.ToLower(f.Message), "useragent mismatch") {
			return ErrUserAgentMismatch
		}
		return fmt.Errorf("%w: HTTP 400: %s", ErrRequestFailed, truncate(string(body), 200))
	case code >= 500:
		return fmt.Errorf("%w: HTTP %d", ErrRequestFailed, code)
	default:
		return fmt.Errorf("%w: unexpected HTTP %d: %s", ErrRequestFailed, code, truncate(string(body), 200))
	}
}

// mapFailStatus inspects 2xx responses for application-level "status: fail"
// envelopes returned by some Threads / Instagram endpoints.
func mapFailStatus(code int, body []byte) error {
	if code < 200 || code >= 300 {
		return nil
	}
	if !bytes.Contains(body, []byte(`"status":"fail"`)) {
		return nil
	}
	var f struct {
		Status       string `json:"status"`
		Message      string `json:"message"`
		ErrorTitle   string `json:"error_title"`
		ErrorBody    string `json:"error_body"`
		LogoutReason int    `json:"logout_reason"`
	}
	if err := json.Unmarshal(body, &f); err != nil || f.Status != "fail" {
		return nil
	}
	if f.LogoutReason == 8 || f.Message == "login_required" {
		return ErrSessionSuspended
	}
	return &FailStatusError{
		Message:      f.Message,
		ErrorTitle:   f.ErrorTitle,
		ErrorBody:    f.ErrorBody,
		LogoutReason: f.LogoutReason,
		HTTPStatus:   code,
	}
}

// ----------------------------------------------------------------------------
// Rate-limit circuit-breaker
// ----------------------------------------------------------------------------

// cooldownFor returns the cooldown duration for the given request kind.
func (c *Client) cooldownFor(isWrite bool) time.Duration {
	if isWrite {
		return c.writeCooldown
	}
	return c.readCooldown
}

// tripCooldown records a rate-limit event and arms the cooldown timer for
// the appropriate request kind. Concurrent calls are safe.
func (c *Client) tripCooldown(isWrite bool, d time.Duration, reason string) {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()
	until := time.Now().Add(d)
	c.rlState.LastBlockedAt = time.Now()
	c.rlState.BlockedReason = reason
	c.rlState.WriteBlocked = isWrite
	if isWrite {
		if until.After(c.cooldownWrt) {
			c.cooldownWrt = until
		}
	} else {
		if until.After(c.cooldownRead) {
			c.cooldownRead = until
		}
	}
}

// waitForCooldownInternal blocks until the per-kind cooldown has elapsed,
// or the context is cancelled. It is called inside doWithRetry before
// applying the normal request-gap pacing.
func (c *Client) waitForCooldownInternal(ctx context.Context, isWrite bool) error {
	c.rlMu.Lock()
	var until time.Time
	if isWrite {
		until = c.cooldownWrt
	} else {
		until = c.cooldownRead
	}
	c.rlMu.Unlock()

	if d := time.Until(until); d > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}
	}
	return nil
}

// recordSignals captures the Instagram/Meta soft-signal response headers
// that indicate server capacity and request timing. Called after every
// successful HTTP round-trip regardless of status code.
func (c *Client) recordSignals(resp *http.Response, isWrite bool) {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()
	if isWrite {
		c.rlState.LastWriteAt = time.Now()
	} else {
		c.rlState.LastReadAt = time.Now()
	}
	if v := resp.Header.Get("x-ig-capacity-level"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.rlState.CapacityLevel = n
		}
	} else {
		c.rlState.CapacityLevel = -1
	}
	c.rlState.PeakTime = resp.Header.Get("x-ig-peak-time") == "1"
	c.rlState.PeakV2 = resp.Header.Get("x-ig-peak-v2") == "1"
	if q := resp.Header.Get("x-fb-connection-quality"); q != "" {
		c.rlState.ConnectionQuality = q
	}
	if r := resp.Header.Get("x-ig-origin-region"); r != "" {
		c.rlState.OriginRegion = r
	}
	if r := resp.Header.Get("x-ig-server-region"); r != "" {
		c.rlState.ServerRegion = r
	}
	if ms := resp.Header.Get("x-ig-request-elapsed-time-ms"); ms != "" {
		if n, err := strconv.Atoi(ms); err == nil {
			c.rlState.LastServerElapsedMs = n
		}
	}
}

// RateLimit returns a snapshot of the most recent rate-limit observations.
// Use CooldownReadUntil / CooldownWriteUntil to determine when the client
// will resume normal operation.
func (c *Client) RateLimit() RateLimitState {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()
	rs := c.rlState
	rs.CooldownReadUntil = c.cooldownRead
	rs.CooldownWriteUntil = c.cooldownWrt
	return rs
}

// WaitForCooldown blocks until both the read and write cooldowns have
// elapsed, or the context is cancelled. Returns ctx.Err() on cancellation.
// Call this before resuming a batch of requests after a rate-limit event.
func (c *Client) WaitForCooldown(ctx context.Context) error {
	c.rlMu.Lock()
	readUntil := c.cooldownRead
	writeUntil := c.cooldownWrt
	c.rlMu.Unlock()

	until := readUntil
	if writeUntil.After(until) {
		until = writeUntil
	}
	if d := time.Until(until); d > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}
	}
	return nil
}

// waitForGap enforces the minimum request gap to avoid tripping Threads'
// behavioural rate limiter.
func (c *Client) waitForGap(ctx context.Context) {
	c.gapMu.Lock()
	now := time.Now()
	nextSlot := c.lastReqAt.Add(c.minGap)
	if now.After(nextSlot) {
		nextSlot = now
	}
	c.lastReqAt = nextSlot
	c.gapMu.Unlock()

	if wait := time.Until(nextSlot); wait > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
	}
}

// isNonRetriable reports whether err should not be retried.
func isNonRetriable(err error) bool {
	return errors.Is(err, ErrInvalidAuth) ||
		errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrSessionSuspended) ||
		errors.Is(err, ErrInvalidParams) ||
		errors.Is(err, ErrUserAgentMismatch) ||
		errors.Is(err, ErrWriteAuthRequired) ||
		errors.Is(err, ErrRateLimited) // cooldown already tripped; don't hammer
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
