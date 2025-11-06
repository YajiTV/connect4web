package httphandler

import (
	"net/http"
	"strings"
	"unicode"

	"power4/internal/auth"
)

type headerData struct {
	LoggedIn         bool
	Username         string
	Initials         string
	CSRF             string
	HasFriendAlerts  bool
	FriendAlertCount int
}

func makeHeader(w http.ResponseWriter, r *http.Request) headerData {
	u := auth.CurrentUser(userStore, r)
	csrf := auth.CSRFToken(w, r)
	if u == nil {
		return headerData{CSRF: csrf}
	}
	rc := friendsIncomingCount(u.Username) + invitesIncomingCount(u.Username)
	return headerData{
		LoggedIn:         true,
		Username:         u.Username,
		Initials:         initials(u.Username),
		CSRF:             csrf,
		HasFriendAlerts:  rc > 0,
		FriendAlertCount: rc,
	}
}

func initials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "??"
	}
	parts := strings.Fields(name)
	var out []rune
	for _, p := range parts {
		for _, r := range p {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				out = append(out, unicode.ToUpper(r))
				break
			}
		}
		if len(out) == 2 {
			break
		}
	}
	if len(out) == 1 {
		for _, r := range parts[0] {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				if r == out[0] {
					continue
				}
				out = append(out, unicode.ToUpper(r))
				break
			}
		}
	}
	if len(out) == 0 {
		return "??"
	}
	if len(out) == 1 {
		return string(out[0])
	}
	return string(out[:2])
}
