package httphandler

import (
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"strings"
)

func pidCookieName() string { return "pid" }

func getOrSetPID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie(pidCookieName()); err == nil && c.Value != "" {
		return c.Value
	}
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	id := strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), "=")
	http.SetCookie(w, &http.Cookie{
		Name:     pidCookieName(),
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}
