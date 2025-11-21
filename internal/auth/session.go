package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"power4/internal/util"
)

// Session holds session payload carried in the signed token
type Session struct {
	UserID  string    // authenticated user id, empty for anonymous
	CSRF    string    // csrf token bound to this session
	Expires time.Time // absolute expiration time
}

// secret stores the HMAC key used to sign and verify session tokens and persists it on disk
var secret []byte

// InitSessions initializes the session key in dir: tries to load an existing 32‑byte key or generates and saves a new one
func InitSessions(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, "session.key")

	// tries to load existing key first
	b, err := os.ReadFile(p)
	if err == nil && len(b) >= 32 {
		secret = b
		return nil
	}

	// generates and saves a new key
	nb := make([]byte, 32)
	if _, err := rand.Read(nb); err != nil {
		return err
	}
	if err := os.WriteFile(p, nb, 0o600); err != nil {
		return err
	}
	secret = nb
	return nil
}

// randToken generates a random base32 token of length n
func randToken(n int) string {
	s, err := util.RandBase32(n)
	if err != nil {
		panic(err) // panics if CSPRNG is unavailable
	}
	return s
}

// build creates a signed session token v1|userID|csrf|unix using HMAC‑SHA256 and URL‑safe base64
func build(userID, csrf string, exp time.Time) string {
	payload := strings.Join([]string{
		"v1",
		userID,
		csrf,
		strconv.FormatInt(exp.Unix(), 10),
	}, "|")

	m := hmac.New(sha256.New, secret)
	m.Write([]byte(payload))
	sig := m.Sum(nil)

	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// parse validates HMAC, checks version and expiry, and returns the decoded session
func parse(val string) (*Session, error) {
	parts := strings.Split(val, ".")
	if len(parts) != 2 {
		return nil, errors.New("bad token")
	}

	// decodes payload and signature
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// verifies signature
	m := hmac.New(sha256.New, secret)
	m.Write(raw)
	if !hmac.Equal(m.Sum(nil), sig) {
		return nil, errors.New("bad mac")
	}

	// validates format and version
	fs := strings.Split(string(raw), "|")
	if len(fs) != 4 || fs[0] != "v1" {
		return nil, errors.New("bad payload")
	}

	// parses expiry and builds the session
	ux, err := strconv.ParseInt(fs[3], 10, 64)
	if err != nil {
		return nil, err
	}
	s := &Session{
		UserID:  fs[1],
		CSRF:    fs[2],
		Expires: time.Unix(ux, 0),
	}

	// rejects expired session
	if time.Now().After(s.Expires) {
		return nil, errors.New("expired")
	}
	return s, nil
}

// setCookie sets the signed session cookie with HttpOnly, SameSite Lax, and a TLS‑aware Secure flag
func setCookie(w http.ResponseWriter, r *http.Request, tok string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Expires:  exp,
	})
}

// EnsureSession tries to use an existing valid session or generates a new anonymous one with CSRF
func EnsureSession(w http.ResponseWriter, r *http.Request) *Session {
	if c, err := r.Cookie("sid"); err == nil && c.Value != "" {
		if s, err := parse(c.Value); err == nil {
			return s
		}
	}

	// issues a fresh anonymous session
	csrf := randToken(24)
	exp := time.Now().Add(24 * time.Hour)
	tok := build("", csrf, exp)
	setCookie(w, r, tok, exp)
	return &Session{CSRF: csrf, Expires: exp}
}

// StartSession creates an authenticated session for userID and sets the cookie
func StartSession(w http.ResponseWriter, r *http.Request, userID string) *Session {
	csrf := randToken(24)
	exp := time.Now().Add(7 * 24 * time.Hour)
	tok := build(userID, csrf, exp)
	setCookie(w, r, tok, exp)
	return &Session{UserID: userID, CSRF: csrf, Expires: exp}
}

// CurrentUser looks up and returns the user from the store for the current session or nil if unauthenticated
func CurrentUser(store *Store, r *http.Request) *User {
	c, err := r.Cookie("sid")
	if err != nil || c.Value == "" {
		return nil
	}
	s, err := parse(c.Value)
	if err != nil || s.UserID == "" {
		return nil
	}
	return store.GetByID(s.UserID)
}

// Logout clears the session cookie
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}

// CSRFToken gets the CSRF token for the current session, creating a session if needed
func CSRFToken(w http.ResponseWriter, r *http.Request) string {
	s := EnsureSession(w, r)
	return s.CSRF
}

// CheckCSRF verifies that POST requests carry a CSRF token matching the session
func CheckCSRF(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return true
	}
	c, err := r.Cookie("sid")
	if err != nil || c.Value == "" {
		return false
	}
	s, err := parse(c.Value)
	if err != nil {
		return false
	}
	return r.PostFormValue("csrf") == s.CSRF
}
