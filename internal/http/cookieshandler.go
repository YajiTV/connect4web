package httphandler

import (
	"net/http"

	"power4/internal/util"
)

// pidCookieName returns the cookie name used to identify players
func pidCookieName() string { return "pid" }

// getOrSetPID tries to read the player id cookie or generates a new one and sets it
func getOrSetPID(w http.ResponseWriter, r *http.Request) string {
	// tries existing cookie first
	if c, err := r.Cookie(pidCookieName()); err == nil && c.Value != "" {
		return c.Value
	}

	// generates a new id and sets a secure cookie
	id, err := util.RandBase32(12)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     pidCookieName(),
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
	return id
}
