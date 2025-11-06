package httphandler

import (
	"html/template"
	"log"
	"net/http"
)

func ShowHome(w http.ResponseWriter, r *http.Request) {
	_ = getOrSetPID(w, r)
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "index.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func NotFound(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "404.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "404 - Page not found", http.StatusNotFound)
		return
	}
	h := makeHeader(w, r)
	w.WriteHeader(http.StatusNotFound)
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
