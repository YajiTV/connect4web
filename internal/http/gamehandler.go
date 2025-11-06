package httphandler

import (
	"html/template"
	"log"
	"net/http"
)

func ShowGame(w http.ResponseWriter, r *http.Request) {
	rm, code := roomFromPath(r.URL.Path)
	if rm == nil || code == "" {
		NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "game.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reveal := r.URL.Query().Get("reveal") == "1"
	h := makeHeader(w, r)
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		Code             string
		Random           bool
		Reveal           bool
		LoggedIn         bool
		Username         string
		Initials         string
		CSRF             string
		HasFriendAlerts  bool
		FriendAlertCount int
	}{
		Code:             rm.Code,
		Random:           rm.Random,
		Reveal:           reveal,
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		CSRF:             h.CSRF,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
	})
}
