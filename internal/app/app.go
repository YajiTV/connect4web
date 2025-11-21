package app

import (
	"io/fs"
	"log"
	"net/http"
	"power4"
	"power4/internal/auth"
	httphandler "power4/internal/http"
)

// Boot wires up templates, sessions, stores, and HTTP routes, and returns the mux with a cleanup hook
func Boot(dataDir string) (*http.ServeMux, error) {
	// Loads templates from the embedded filesystem
	templatesFS, err := fs.Sub(power4.Content, "templates")
	if err != nil {
		return nil, err
	}
	// Makes templates available to HTTP handlers
	httphandler.SetTemplateFS(templatesFS)

	// Initializes session storage under dataDir
	if err := auth.InitSessions(dataDir); err != nil {
		return nil, err
	}

	// Opens the user store and exposes it to handlers
	store, err := auth.Open(dataDir)
	if err != nil {
		return nil, err
	}
	httphandler.SetUserStore(store)

	// Tries to load the friends list
	if err := httphandler.InitFriendsStore(dataDir + "/friends.json"); err != nil {
		log.Printf("friends load error: %v", err)
	}

	// Serves static assets from the embedded filesystem
	staticFS, err := fs.Sub(power4.Content, "static")
	if err != nil {
		return nil, err
	}

	// Builds the HTTP router
	mux := httphandler.NewRouter(staticFS)
	return mux, nil
}
