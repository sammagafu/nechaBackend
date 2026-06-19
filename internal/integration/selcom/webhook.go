package selcom

import (
	"crypto/subtle"
	"strconv"
	"strings"
)

// VerifyWebhookSecret checks a shared secret header from Selcom callbacks.
func VerifyWebhookSecret(provided, expected string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(provided)), []byte(expected)) == 1
}

// ParseWebhookAmount normalises amount strings from webhook payloads.
func ParseWebhookAmount(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	// Selcom may send major units as decimal strings.
	if strings.Contains(raw, ".") {
		parts := strings.SplitN(raw, ".", 2)
		whole, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, false
		}
		return whole, true
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}
