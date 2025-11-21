package util

import (
	"crypto/rand"
	"encoding/base32"
	"io"
	"strings"
)

// RandBase32 generates a random base32 string of n raw bytes, without padding
func RandBase32(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	return strings.TrimRight(s, "="), nil
}
