package httphandler

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"power4/internal/auth"
)

// ShowProfile renders a user's public profile with friendship context
func ShowProfile(w http.ResponseWriter, r *http.Request) {
	// extracts the profile username from the URL
	username := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if username == "" || userStore == nil {
		NotFound(w, r)
		return
	}

	// looks up the user
	u := userStore.GetByUsername(username)
	if u == nil {
		NotFound(w, r)
		return
	}

	h := makeHeader(w, r)
	csrf := auth.CSRFToken(w, r)

	// computes friendship state relative to the viewer
	areFriends, outPending, inPending := false, false, false
	if h.LoggedIn && !strings.EqualFold(h.Username, u.Username) {
		areFriends, outPending, inPending = friendsState(h.Username, u.Username)
	}

	// parses and renders the template
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "profile.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		Username         string
		Initials         string
		LoggedIn         bool
		HasFriendAlerts  bool
		FriendAlertCount int
		CSRF             string

		ProfileUsername string
		Elo             int
		Games           int
		Wins            int
		Losses          int
		Self            string

		AreFriends bool
		OutPending bool
		InPending  bool
	}{
		Username:         h.Username,
		Initials:         h.Initials,
		LoggedIn:         h.LoggedIn,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
		CSRF:             csrf,

		ProfileUsername: u.Username,
		Elo:             u.Elo,
		Games:           u.Games,
		Wins:            u.Wins,
		Losses:          u.Losses,
		Self:            h.Username,

		AreFriends: areFriends,
		OutPending: outPending,
		InPending:  inPending,
	})
}
