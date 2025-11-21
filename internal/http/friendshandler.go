package httphandler

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"power4/internal/auth"
)

type friendsPageData struct {
	LoggedIn         bool     // whether the viewer is authenticated
	Username         string   // display name
	Initials         string   // initials for avatar
	CSRF             string   // csrf token for actions
	HasFriendAlerts  bool     // whether the header should show alerts
	FriendAlertCount int      // number of pending alerts
	Friends          []string // accepted friends
	Requests         []string // incoming friend requests
	Invites          []struct {
		Ticket     string // challenge ticket id
		FromUser   string // challenger username
		CreatedISO string // creation time in RFC3339
	}
	SearchQuery string // query typed by the user
	SearchFound bool   // whether a user was found
	SearchUser  string // found username
	AreFriends  bool   // whether the viewer and found user are friends
	OutPending  bool   // whether an outgoing request exists
	InPending   bool   // whether an incoming request exists
}

// ShowFriends renders the friends page with search context
func ShowFriends(w http.ResponseWriter, r *http.Request) {
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h := makeHeader(w, r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	// resolves search hit and relationship state
	var found bool
	var su string
	var areFriends, outP, inP bool
	if q != "" {
		up := userStore.GetByUsername(q)
		if up != nil && !strings.EqualFold(up.Username, u.Username) {
			found = true
			su = up.Username
			areFriends, outP, inP = friendsState(u.Username, up.Username)
		}
	}

	// renders page
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "friends.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "base", friendsPageData{
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		CSRF:             h.CSRF,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
		Friends:          nil,
		Requests:         nil,
		Invites:          nil,
		SearchQuery:      q,
		SearchFound:      found,
		SearchUser:       su,
		AreFriends:       areFriends,
		OutPending:       outP,
		InPending:        inP,
	})
}

// SendFriendRequest creates or reuses a friend request and redirects back
func SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	to := strings.TrimSpace(r.FormValue("to"))
	back := strings.TrimSpace(r.FormValue("back"))

	// validates target
	if to == "" {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	if userStore.GetByUsername(to) == nil {
		if back != "" {
			http.Redirect(w, r, back, http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/friends?q="+to, http.StatusSeeOther)
		return
	}

	// sends the request and redirects
	_ = sendFriendRequest(u.Username, to)
	if back != "" {
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/friends?q="+to, http.StatusSeeOther)
}

// AcceptFriendRequest accepts an incoming friend request
func AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	from := strings.TrimSpace(r.FormValue("from"))
	back := strings.TrimSpace(r.FormValue("back"))
	if from != "" {
		_ = acceptFriendRequest(u.Username, from)
	}
	if back != "" {
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}

// DeclineFriendRequest declines an incoming friend request
func DeclineFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	from := strings.TrimSpace(r.FormValue("from"))
	back := strings.TrimSpace(r.FormValue("back"))
	if from != "" {
		_ = declineFriendRequest(u.Username, from)
	}
	if back != "" {
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}

// SendChallenge sends a challenge to a friend and redirects to the wait page
func SendChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	friend := strings.TrimSpace(r.FormValue("friend"))
	if friend == "" {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	if !isFriends(u.Username, friend) {
		http.Redirect(w, r, "/friends?q="+friend, http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)
	t := token()
	_, _ = sendChallenge(u.Username, pid, friend, t)
	http.Redirect(w, r, "/friends/challenge/wait/"+t, http.StatusSeeOther)
}

// ShowChallengeWait displays the waiting page for an outgoing challenge
func ShowChallengeWait(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	t := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if t == "" || getInvite(t) == nil {
		NotFound(w, r)
		return
	}

	h := makeHeader(w, r)
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "challenge_wait.tmpl")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		Ticket           string
		LoggedIn         bool
		Username         string
		Initials         string
		CSRF             string
		HasFriendAlerts  bool
		FriendAlertCount int
	}{
		Ticket:           t,
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		CSRF:             h.CSRF,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
	})
}

// CheckChallenge polls for acceptance and redirects to the game when ready
func CheckChallenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	t := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	inv := getInvite(t)
	if inv == nil {
		NotFound(w, r)
		return
	}

	select {
	case code := <-inv.Ch:
		http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
		return
	case <-time.After(25 * time.Second):
		http.Redirect(w, r, "/friends/challenge/wait/"+t+"?ts="+strconvI(int(time.Now().Unix())), http.StatusSeeOther)
		return
	case <-r.Context().Done():
		return
	}
}

// AcceptChallenge accepts a received challenge and starts a game
func AcceptChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	t := strings.TrimSpace(r.FormValue("ticket"))
	if t == "" {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)
	code, err := acceptChallenge(t, u.Username, pid)
	if err != nil {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
}

// DeclineChallenge declines a received challenge
func DeclineChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	t := strings.TrimSpace(r.FormValue("ticket"))
	back := strings.TrimSpace(r.FormValue("back"))
	if t != "" {
		_ = declineChallenge(t, u.Username)
	}
	if back != "" {
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}

// CancelChallenge cancels an outgoing challenge
func CancelChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/friends", http.StatusSeeOther)
		return
	}
	t := strings.TrimSpace(r.FormValue("ticket"))
	if t != "" {
		cancelChallenge(t, "")
	}
	http.Redirect(w, r, "/friends", http.StatusSeeOther)
}

// ShowFriendsRequestsSection renders the iframe with incoming requests and invites
func ShowFriendsRequestsSection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// gathers requests and incoming invites
	reqs := listIncomingRequests(u.Username)
	invRaw := listIncomingInvites(u.Username)
	inv := make([]struct {
		Ticket     string
		FromUser   string
		CreatedISO string
	}, 0, len(invRaw))
	for _, iv := range invRaw {
		inv = append(inv, struct {
			Ticket     string
			FromUser   string
			CreatedISO string
		}{
			Ticket:     iv.Ticket,
			FromUser:   iv.Challenger,
			CreatedISO: iv.Created.UTC().Format(time.RFC3339),
		})
	}

	// renders iframe
	tmpl, err := template.ParseFS(templateFS, "friends_requests_iframe.tmpl")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	h := makeHeader(w, r)
	_ = tmpl.Execute(w, struct {
		CSRF     string
		Requests []string
		Invites  []struct {
			Ticket     string
			FromUser   string
			CreatedISO string
		}
	}{
		CSRF:     h.CSRF,
		Requests: reqs,
		Invites:  inv,
	})
}

// ShowFriendsFriendsSection renders the iframe with the friends list
func ShowFriendsFriendsSection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	fl := listFriends(u.Username)

	tmpl, err := template.ParseFS(templateFS, "friends_friends_iframe.tmpl")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	h := makeHeader(w, r)
	_ = tmpl.Execute(w, struct {
		CSRF    string
		Friends []string
	}{
		CSRF:    h.CSRF,
		Friends: fl,
	})
}
