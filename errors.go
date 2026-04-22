package threads

import "errors"

var (
	// ErrInvalidAuth is returned when required credentials are missing
	// from a Cookies or Auth struct passed to New / NewWithAuth.
	ErrInvalidAuth = errors.New("threads: missing required credentials")

	// ErrUnauthorized is returned when the server rejects the session,
	// typically with HTTP 401 or "login_required". The session may have
	// expired or been invalidated.
	ErrUnauthorized = errors.New("threads: authentication failed — session may be expired")

	// ErrSessionSuspended is returned when the server returns 403 with
	// logout_reason 8 (device fingerprint anomaly). Sessions usually
	// recover within ~15-30 minutes of no further requests.
	ErrSessionSuspended = errors.New("threads: session temporarily suspended (logout_reason 8) — wait and retry")

	// ErrForbidden is returned when access to a resource is denied
	// (private account the viewer doesn't follow, blocked viewer, etc.).
	ErrForbidden = errors.New("threads: access denied")

	// ErrNotFound is returned when a resource (user, post, hashtag) does
	// not exist or has been deleted.
	ErrNotFound = errors.New("threads: resource not found")

	// ErrRateLimited is returned when the server returns HTTP 429.
	ErrRateLimited = errors.New("threads: rate limited")

	// ErrInvalidParams is returned when the caller passes invalid or
	// missing required arguments.
	ErrInvalidParams = errors.New("threads: invalid or missing required parameters")

	// ErrRequestFailed wraps unexpected HTTP / transport / decoding errors.
	ErrRequestFailed = errors.New("threads: HTTP request failed")

	// ErrWriteAuthRequired is returned when a write method is called on
	// a client created via New() (cookie-only). Write operations require
	// a Bearer token client created via NewWithAuth().
	ErrWriteAuthRequired = errors.New("threads: write operations require Bearer token auth (use NewWithAuth)")

	// ErrUserAgentMismatch indicates the server rejected the request's
	// User-Agent. Use the default UA or one of the supported variants.
	ErrUserAgentMismatch = errors.New("threads: useragent mismatch (set via WithUserAgent)")
)

// FailStatusError carries the body of a "status: fail" response from
// the Threads / Instagram backend, including the optional message.
// It wraps ErrRequestFailed for errors.Is compatibility.
type FailStatusError struct {
	Message      string
	ErrorTitle   string
	ErrorBody    string
	LogoutReason int
	HTTPStatus   int
}

func (e *FailStatusError) Error() string {
	if e.Message == "" {
		return ErrRequestFailed.Error()
	}
	return "threads: " + e.Message
}

func (e *FailStatusError) Unwrap() error {
	return ErrRequestFailed
}
