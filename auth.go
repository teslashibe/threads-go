package threads

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoginResult is the output of Login. Token is the IGT:2:... bearer string;
// UserID is the numeric ID of the authenticated user. The DeviceID echoes
// the value passed to Login (or generated if blank).
type LoginResult struct {
	Token    string `json:"token"`
	UserID   string `json:"userId"`
	Username string `json:"username,omitempty"`
	DeviceID string `json:"deviceId"`
}

// LoginParams configures a Bloks login. Username and Password are required.
// DeviceID should be a stable android-{13chars} value (generated if blank)
// reused across logins so Meta sees a consistent device fingerprint.
// TOTPSecret, when set, is the base32 authenticator-app secret used to
// auto-solve a two-factor challenge if the account has app-based 2FA enabled.
type LoginParams struct {
	Username   string
	Password   string
	DeviceID   string
	TOTPSecret string
}

// Login performs the Bloks login flow against i.instagram.com to obtain a
// Bearer token suitable for use with NewWithAuth / NewFull. It is a thin
// wrapper over LoginWith for the common no-2FA case.
//
//	res, err := threads.Login(ctx, "user@example.com", "password", "")
func Login(ctx context.Context, username, password, deviceID string, opts ...Option) (*LoginResult, error) {
	return LoginWith(ctx, LoginParams{Username: username, Password: password, DeviceID: deviceID}, opts...)
}

// LoginWith performs the Bloks login flow with full control over parameters,
// including TOTP-based two-factor auth. When the first login response is a
// two_factor_required challenge and a TOTPSecret is supplied, it generates
// the current code and completes the second factor automatically.
//
// Meta actively detects non-app login traffic and may challenge with a
// checkpoint; in that case the response fails with a message containing
// "checkpoint_required" and the caller must complete verification in-app
// (or paste a Bearer token manually). SMS/email 2FA is not supported — only
// authenticator-app TOTP.
func LoginWith(ctx context.Context, p LoginParams, opts ...Option) (*LoginResult, error) {
	return newBaseClient(opts...).performLogin(ctx, p)
}

// performLogin runs the Bloks login (+ optional TOTP second factor) using the
// receiver's HTTP transport and User-Agents. Reused both by the package-level
// LoginWith and by a live client's auto re-login on write-auth failure.
func (c *Client) performLogin(ctx context.Context, p LoginParams) (*LoginResult, error) {
	if p.Username == "" || p.Password == "" {
		return nil, fmt.Errorf("%w: username and password must not be empty", ErrInvalidParams)
	}
	deviceID := p.DeviceID
	if deviceID == "" {
		deviceID = GenerateDeviceID()
	}

	clientInput := map[string]interface{}{
		"password":      p.Password,
		"contact_point": p.Username,
		"device_id":     deviceID,
	}
	serverParams := map[string]interface{}{
		"credential_type": "password",
		"device_id":       deviceID,
	}
	body, err := c.bloksLoginRequest(ctx, "com.bloks.www.bloks.caa.login.async.send_login_request", clientInput, serverParams)
	if err != nil {
		return nil, err
	}

	// Detect a two-factor challenge and solve it with TOTP when possible.
	if id := twoFactorIdentifier(body); id != "" {
		if p.TOTPSecret == "" {
			return nil, fmt.Errorf("%w: account requires two-factor auth — provide a TOTPSecret (authenticator app) or paste a Bearer token manually", ErrUnauthorized)
		}
		code, terr := totpCode(p.TOTPSecret, time.Now())
		if terr != nil {
			return nil, terr
		}
		tfInput := map[string]interface{}{
			"two_factor_identifier": id,
			"verification_code":     code,
			"verification_method":   "3", // 3 = authenticator app (TOTP)
			"device_id":             deviceID,
			"username":              p.Username,
		}
		tfServer := map[string]interface{}{
			"device_id":             deviceID,
			"two_factor_identifier": id,
		}
		body, err = c.bloksLoginRequest(ctx, "com.bloks.www.bloks.caa.login.async.send_two_factor_login_request", tfInput, tfServer)
		if err != nil {
			return nil, err
		}
	}

	token := extractBearerToken(body)
	if token == "" {
		return nil, fmt.Errorf("%w: no Bearer token in login response (snippet: %s)", ErrRequestFailed, truncate(string(body), 300))
	}

	res := &LoginResult{
		Token:    token,
		DeviceID: deviceID,
		UserID:   extractField(body, "user_id"),
		Username: extractField(body, "user_name"),
	}
	if res.UserID == "" {
		res.UserID = extractField(body, "ig_user_id")
	}
	return res, nil
}

// bloksLoginRequest POSTs a Bloks CAA login app request and returns the raw
// response body. appName is the bloks app id (without the trailing slash),
// e.g. "com.bloks.www.bloks.caa.login.async.send_login_request".
func (c *Client) bloksLoginRequest(ctx context.Context, appName string, clientInput, serverParams map[string]interface{}) ([]byte, error) {
	params := map[string]interface{}{
		"client_input_params": clientInput,
		"server_params":       serverParams,
	}
	paramsJSON, _ := json.Marshal(params)

	form := url.Values{}
	form.Set("params", string(paramsJSON))
	form.Set("bk_client_context", `{"bloks_version":"`+blocksVersioningID+`","styles_id":"instagram"}`)
	form.Set("bloks_versioning_id", blocksVersioningID)

	endpoint := writeBaseURL + "/api/v1/bloks/apps/" + appName + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: building login request: %v", ErrRequestFailed, err)
	}
	req.Header.Set("User-Agent", c.writeUA)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-IG-App-ID", appID)
	req.Header.Set("X-IG-Capabilities", "3brTvx0=")
	req.Header.Set("X-IG-Connection-Type", "WIFI")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: login: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("%w: reading login body: %v", ErrRequestFailed, err)
	}
	if err := mapStatus(resp.StatusCode, body); err != nil {
		return nil, err
	}
	return body, nil
}

// twoFactorIdentifier returns the two_factor_identifier embedded in a Bloks
// login response when the account is challenged for a second factor, or ""
// when no challenge is present.
func twoFactorIdentifier(body []byte) string {
	if !strings.Contains(string(body), "two_factor") {
		return ""
	}
	return extractField(body, "two_factor_identifier")
}

// extractBearerToken finds a "Bearer IGT:2:..." substring inside a Bloks
// login response and returns the IGT:2:... token (without the Bearer prefix).
func extractBearerToken(body []byte) string {
	const marker = "Bearer IGT:2:"
	s := string(body)
	idx := strings.Index(s, marker)
	if idx < 0 {
		return ""
	}
	rest := s[idx+len("Bearer "):]
	end := strings.IndexAny(rest, "\\\"\\\\\\, \t\n\r")
	if end < 0 {
		end = len(rest)
	}
	return rest[:end]
}

// extractField scans the Bloks login response for a quoted "key": "value"
// pair and returns value, if present. This is a best-effort regex-free
// extraction sufficient for the few well-known fields that survive Bloks'
// nested escaping.
func extractField(body []byte, key string) string {
	s := string(body)
	target := `\"` + key + `\":\"`
	idx := strings.Index(s, target)
	if idx < 0 {
		target = `"` + key + `":"`
		idx = strings.Index(s, target)
		if idx < 0 {
			return ""
		}
	}
	rest := s[idx+len(target):]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// GenerateDeviceID returns a random android-{13chars} identifier suitable
// for the Bloks login flow. Persist this value to keep a stable device
// fingerprint across runs.
func GenerateDeviceID() string {
	b := make([]byte, 7)
	_, _ = rand.Read(b)
	hexStr := hex.EncodeToString(b) // 14 chars; trim to 13 for the canonical shape
	return "android-" + hexStr[:13]
}
