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
	LoggedIn         bool
	Username         string
	Initials         string
	CSRF             string
	HasFriendAlerts  bool
	FriendAlertCount int
	Friends          []string
	Requests         []string
	Invites          []struct {
		Ticket     string
		FromUser   string
		CreatedISO string
	}
	SearchQuery string
	SearchFound bool
	SearchUser  string
	AreFriends  bool
	OutPending  bool
	InPending   bool
}

func ShowFriends(w http.ResponseWriter, r *http.Request) {
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h := makeHeader(w, r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
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
	_ = sendFriendRequest(u.Username, to)
	if back != "" {
		http.Redirect(w, r, back, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/friends?q="+to, http.StatusSeeOther)
}

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

func ShowFriendsRequestsSection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
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

// Iframe: Your Friends (auto-refresh)
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
