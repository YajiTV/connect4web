package httphandler

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"

	"power4/internal/auth"
)

func ShowProfile(w http.ResponseWriter, r *http.Request) {
	username := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if username == "" || userStore == nil {
		NotFound(w, r)
		return
	}
	u := userStore.GetByUsername(username)
	if u == nil {
		NotFound(w, r)
		return
	}

	h := makeHeader(w, r)
	csrf := auth.CSRFToken(w, r)

	areFriends, outPending, inPending := false, false, false
	if h.LoggedIn && !strings.EqualFold(h.Username, u.Username) {
		areFriends, outPending, inPending = friendsState(h.Username, u.Username)
	}

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
