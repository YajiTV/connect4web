package httphandler

import (
	"net/http"
	"strings"
	"unicode"

	"power4/internal/auth"
)

type headerData struct {
	LoggedIn         bool   // whether the user is authenticated
	Username         string // username for display
	Initials         string // up to two initials derived from the username
	CSRF             string // csrf token for forms
	HasFriendAlerts  bool   // whether to show friend alert badge
	FriendAlertCount int    // number of pending friend alerts
}

// makeHeader builds the header data for templates and generates a csrf token
func makeHeader(w http.ResponseWriter, r *http.Request) headerData {
	u := auth.CurrentUser(userStore, r)
	csrf := auth.CSRFToken(w, r)
	if u == nil {
		return headerData{CSRF: csrf}
	}

	// aggregates pending requests and invites
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

// initials generates up to two alphanumeric initials from a name
func initials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "??"
	}

	parts := strings.Fields(name)
	var out []rune

	// picks the first alphanumeric rune from each part
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

	// if only one found, tries to get a second from the first part
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

	// fallbacks
	if len(out) == 0 {
		return "??"
	}
	if len(out) == 1 {
		return string(out[0])
	}
	return string(out[:2])
}
