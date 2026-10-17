package share

import (
	"crypto/rand"
	"errors"
	"strings"
)

const codeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateCode returns a cryptographically-random base62 short code of the
// given length. Rejection sampling avoids modulo bias.
func GenerateCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("code length must be positive")
	}

	// 62 * 4 == 248 is the largest multiple of the alphabet size below 256,
	// which lets us map a single random byte without modulo bias.
	const cutoff = 248

	out := make([]byte, length)
	rb := make([]byte, 1)
	for i := range out {
		for {
			if _, err := rand.Read(rb); err != nil {
				return "", err
			}
			if v := int(rb[0]); v < cutoff {
				out[i] = codeAlphabet[v%62]
				break
			}
		}
	}
	return string(out), nil
}

// ValidCodeFormat reports whether code is a well-formed short code (bounded
// length, base62 alphabet only).
func ValidCodeFormat(code string) bool {
	if code == "" || len(code) > 64 {
		return false
	}
	for _, r := range code {
		if !strings.ContainsRune(codeAlphabet, r) {
			return false
		}
	}
	return true
}
