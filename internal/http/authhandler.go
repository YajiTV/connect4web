package httphandler

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"power4/internal/auth"
)

type AuthPageData struct {
	CSRF             string // csrf token for form posts
	LoggedIn         bool   // whether the current request is authenticated
	Username         string // display name if logged in
	Initials         string // initials for avatar display
	Error            string // error message to show on the page
	HasFriendAlerts  bool   // whether there are pending friend alerts
	FriendAlertCount int    // number of pending friend alerts
}

// getSignupErrorMessage generates a user friendly signup error message
func getSignupErrorMessage(err error) string {
	e := strings.ToLower(err.Error())
	if strings.Contains(e, "username taken") || strings.Contains(e, "already") || strings.Contains(e, "duplicate") {
		return "Username is already taken"
	}
	if strings.Contains(e, "empty username") || strings.Contains(e, "username is required") {
		return "Username is required"
	}
	if strings.Contains(e, "weak password") || strings.Contains(e, "password must be at least 6") {
		return "Password must be at least 6 characters long"
	}
	return "Signup failed, please try again"
}

// getLoginErrorMessage generates a user friendly login error message
func getLoginErrorMessage(err error) string {
	e := strings.ToLower(err.Error())
	if strings.Contains(e, "invalid credentials") {
		return "Username or password is incorrect"
	}
	return "Username or password is incorrect"
}

// renderAuthPage renders an auth page using the shared base template and provided status
func renderAuthPage(w http.ResponseWriter, r *http.Request, templateName string, errorMsg string, status int) {
	csrf := auth.CSRFToken(w, r)
	h := makeHeader(w, r)

	tmpl, err := template.ParseFS(templateFS, "base.tmpl", templateName+".tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if status > 0 {
		w.WriteHeader(status)
	}

	_ = tmpl.ExecuteTemplate(w, "base", AuthPageData{
		CSRF:             csrf,
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		Error:            errorMsg,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
	})
}

// ShowSignup serves the signup page
func ShowSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}
	renderAuthPage(w, r, "signup", "", http.StatusOK)
}

// DoSignup handles signup form submission and creates an account
func DoSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	// validates input
	if username == "" {
		renderAuthPage(w, r, "signup", "Username is required", 422)
		return
	}
	if password == "" {
		renderAuthPage(w, r, "signup", "Password is required", 422)
		return
	}
	if len(username) > 20 {
		renderAuthPage(w, r, "signup", "Username must be at most 20 characters long", 422)
		return
	}
	if len(username) < 3 {
		renderAuthPage(w, r, "signup", "Username must be at least 3 characters long", 422)
		return
	}
	if len(password) < 6 {
		renderAuthPage(w, r, "signup", "Password must be at least 6 characters long", 422)
		return
	}

	// tries to create the user and starts a session
	u, err := userStore.Create(username, password)
	if err != nil {
		renderAuthPage(w, r, "signup", getSignupErrorMessage(err), 422)
		return
	}
	auth.StartSession(w, r, u.ID)
	http.Redirect(w, r, "/u/"+u.Username, http.StatusSeeOther)
}

// ShowLogin serves the login page
func ShowLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	renderAuthPage(w, r, "login", "", http.StatusOK)
}

// DoLogin handles login form submission and starts a session
func DoLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	// validates presence of credentials
	if username == "" || password == "" {
		renderAuthPage(w, r, "login", "All fields are required", 422)
		return
	}

	// tries to authenticate and start a session
	u, err := userStore.Authenticate(username, password)
	if err != nil {
		renderAuthPage(w, r, "login", getLoginErrorMessage(err), 422)
		return
	}
	auth.StartSession(w, r, u.ID)
	http.Redirect(w, r, "/u/"+u.Username, http.StatusSeeOther)
}

// DoLogout ends the current session
func DoLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	auth.Logout(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
