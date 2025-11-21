package httphandler

import (
	"html/template"
	"log"
	"net/http"
)

// ShowGame renders the game container page
func ShowGame(w http.ResponseWriter, r *http.Request) {
	// resolves room from the URL
	rm, code := roomFromPath(r.URL.Path)
	if rm == nil || code == "" {
		NotFound(w, r)
		return
	}

	// parses and renders the template
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "game.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// prepares header data and reveal flag
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
