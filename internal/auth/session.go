package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	UserID  string
	CSRF    string
	Expires time.Time
}

var secret []byte

func InitSessions(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, "session.key")
	b, err := os.ReadFile(p)
	if err == nil && len(b) >= 32 {
		secret = b
		return nil
	}
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

func randToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), "=")
}

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

func parse(val string) (*Session, error) {
	parts := strings.Split(val, ".")
	if len(parts) != 2 {
		return nil, errors.New("bad token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	m := hmac.New(sha256.New, secret)
	m.Write(raw)
	if !hmac.Equal(m.Sum(nil), sig) {
		return nil, errors.New("bad mac")
	}
	fs := strings.Split(string(raw), "|")
	if len(fs) != 4 || fs[0] != "v1" {
		return nil, errors.New("bad payload")
	}
	ux, err := strconv.ParseInt(fs[3], 10, 64)
	if err != nil {
		return nil, err
	}
	s := &Session{
		UserID:  fs[1],
		CSRF:    fs[2],
		Expires: time.Unix(ux, 0),
	}
	if time.Now().After(s.Expires) {
		return nil, errors.New("expired")
	}
	return s, nil
}

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

func EnsureSession(w http.ResponseWriter, r *http.Request) *Session {
	if c, err := r.Cookie("sid"); err == nil && c.Value != "" {
		if s, err := parse(c.Value); err == nil {
			return s
		}
	}
	csrf := randToken(24)
	exp := time.Now().Add(24 * time.Hour)
	tok := build("", csrf, exp)
	setCookie(w, r, tok, exp)
	return &Session{CSRF: csrf, Expires: exp}
}

func StartSession(w http.ResponseWriter, r *http.Request, userID string) *Session {
	csrf := randToken(24)
	exp := time.Now().Add(7 * 24 * time.Hour)
	tok := build(userID, csrf, exp)
	setCookie(w, r, tok, exp)
	return &Session{UserID: userID, CSRF: csrf, Expires: exp}
}

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

func CSRFToken(w http.ResponseWriter, r *http.Request) string {
	s := EnsureSession(w, r)
	return s.CSRF
}

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
