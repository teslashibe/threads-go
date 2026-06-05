package threads

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// totpCode derives the current RFC 6238 TOTP code for a base32-encoded
// shared secret (the value Instagram shows when you enable an authenticator
// app, with spaces/padding tolerated). It uses the standard 30-second step,
// 6 digits, and HMAC-SHA1 — the parameters Instagram/Meta use.
//
// Returns ErrInvalidParams if the secret is empty or not valid base32.
func totpCode(secret string, t time.Time) (string, error) {
	key, err := decodeBase32Secret(secret)
	if err != nil {
		return "", err
	}
	return hotpCode(key, uint64(t.Unix())/30), nil
}

// decodeBase32Secret normalises an authenticator secret (uppercases, strips
// spaces, adds padding) and decodes it as standard base32.
func decodeBase32Secret(secret string) ([]byte, error) {
	s := strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	if s == "" {
		return nil, fmt.Errorf("%w: empty TOTP secret", ErrInvalidParams)
	}
	if pad := len(s) % 8; pad != 0 {
		s += strings.Repeat("=", 8-pad)
	}
	key, err := base32.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid base32 TOTP secret: %v", ErrInvalidParams, err)
	}
	return key, nil
}

// hotpCode computes the 6-digit HOTP value (RFC 4226) for a counter, the
// primitive TOTP is built on.
func hotpCode(key []byte, counter uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])
	return fmt.Sprintf("%06d", value%1000000)
}
