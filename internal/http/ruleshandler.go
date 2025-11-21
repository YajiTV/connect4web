package httphandler

import (
	"html/template"
	"log"
	"net/http"
)

func ShowRules(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "rules.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	h := makeHeader(w, r)
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		LoggedIn         bool
		Username         string
		Initials         string
		CSRF             string
		HasFriendAlerts  bool
		FriendAlertCount int
	}{
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		CSRF:             h.CSRF,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
	})
}
