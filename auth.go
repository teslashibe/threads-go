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

// Login performs the Bloks login flow against i.instagram.com to obtain a
// Bearer token suitable for use with NewWithAuth / NewFull.
//
// deviceID may be empty, in which case a random android-* identifier is
// generated. Provide your own to keep a stable device fingerprint across
// runs (recommended — Meta tracks device identity).
//
//	res, err := threads.Login(ctx, "user@example.com", "password", "")
//	if err != nil { ... }
//	c, _ := threads.NewWithAuth(threads.Auth{
//	    Token: res.Token, UserID: res.UserID, DeviceID: res.DeviceID,
//	})
//
// NOTE: This function is provided for completeness based on documented
// reverse-engineered behaviour of the Bloks endpoint. Meta actively detects
// non-app login traffic and may challenge with checkpoints; in that case
// the response will fail with status: fail and message containing
// "checkpoint_required". You may then need to perform the Bloks
// challenge handshake (out of scope here).
func Login(ctx context.Context, username, password, deviceID string, opts ...Option) (*LoginResult, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("%w: username and password must not be empty", ErrInvalidParams)
	}
	if deviceID == "" {
		deviceID = GenerateDeviceID()
	}

	c := newBaseClient(opts...)

	clientInput := map[string]interface{}{
		"password":      password,
		"contact_point": username,
		"device_id":     deviceID,
	}
	serverParams := map[string]interface{}{
		"credential_type": "password",
		"device_id":       deviceID,
	}
	params := map[string]interface{}{
		"client_input_params": clientInput,
		"server_params":       serverParams,
	}
	paramsJSON, _ := json.Marshal(params)

	form := url.Values{}
	form.Set("params", string(paramsJSON))
	form.Set("bk_client_context", `{"bloks_version":"`+blocksVersioningID+`","styles_id":"instagram"}`)
	form.Set("bloks_versioning_id", blocksVersioningID)

	endpoint := writeBaseURL + "/api/v1/bloks/apps/com.bloks.www.bloks.caa.login.async.send_login_request/"
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
