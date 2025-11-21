package httphandler

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"power4/internal/auth"
	"power4/internal/game"
)

// ShowTraining renders the training mode selector
func ShowTraining(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "training.tmpl")
	if err != nil {
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

// levelName returns a label for the bot difficulty level
func levelName(n int) string {
	switch n {
	case 1:
		return "Easy"
	case 2:
		return "Normal"
	case 3:
		return "Hard"
	case 4:
		return "Expert"
	default:
		return "Master"
	}
}

// StartTraining creates a bot match at the chosen difficulty
func StartTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/training", http.StatusSeeOther)
		return
	}
	if !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/training", http.StatusSeeOther)
		return
	}

	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)
	lv, _ := strconv.Atoi(r.FormValue("level"))
	if lv < 1 {
		lv = 1
	}
	if lv > 5 {
		lv = 5
	}

	now := time.Now()
	code := genCode()
	rm := &Room{
		Code:         code,
		Game:         game.NewGame(),
		Player1ID:    pid,
		Player2ID:    "BOT",
		Player1User:  u.Username,
		Player2User:  "",
		CreatedAt:    now,
		Rev:          1,
		subs:         make(map[chan struct{}]struct{}),
		Random:       false,
		TurnDeadline: now.Add(2 * time.Minute),
		StartNext:    game.Player2,
		Bot:          true,
		BotLevel:     lv,
	}
	rm.Game.Player1Name = u.Username
	rm.Game.Player2Name = "Bot " + levelName(lv)

	roomsMu.Lock()
	rooms[code] = rm
	roomsMu.Unlock()

	http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
}
