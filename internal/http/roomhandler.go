package httphandler

import (
	"net/http"
	"strings"
	"time"

	"power4/internal/auth"
	"power4/internal/game"
)

// CreateRoom creates a private room and assigns the creator as player 1
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)
	now := time.Now()
	code := genCode()

	rm := &Room{
		Code:         code,
		Game:         game.NewGame(),
		Player1ID:    pid,
		CreatedAt:    now,
		Rev:          1,
		subs:         make(map[chan struct{}]struct{}),
		Random:       false,
		TurnDeadline: time.Time{},
		StartNext:    game.Player2,
	}
	rm.Player1User = u.Username
	rm.Game.Player1Name = u.Username

	roomsMu.Lock()
	rooms[code] = rm
	roomsMu.Unlock()

	http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
}

// JoinRoom joins an existing room, filling player slots and starting the clock if both are present
func JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)
	code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
	now := time.Now()

	roomsMu.Lock()
	rm, ok := rooms[code]
	if ok {
		// fills player 1 if empty or refreshes their profile info
		if rm.Player1ID == "" {
			if rm.Player1User == u.Username {
				roomsMu.Unlock()
				http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
				return
			}
			rm.Player1ID = pid
			rm.Player1User = u.Username
			rm.Game.Player1Name = u.Username
			rm.Rev++
		} else if rm.Player1ID == pid {
			if rm.Game.Player1Name == "" {
				rm.Player1User = u.Username
				rm.Game.Player1Name = u.Username
				rm.Rev++
			}
		} else if rm.Player2ID == "" || rm.Player2ID == pid {
			// fills or refreshes player 2
			if rm.Player1User == u.Username {
				roomsMu.Unlock()
				http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
				return
			}
			if rm.Player2ID == "" {
				rm.Rev++
			}
			rm.Player2ID = pid
			rm.Player2User = u.Username
			rm.Game.Player2Name = u.Username
		}

		// starts the turn deadline once both players are present
		if rm.Player1ID != "" && rm.Player2ID != "" && rm.TurnDeadline.IsZero() {
			rm.TurnDeadline = now.Add(2 * time.Minute)
			rm.Rev++
		}
	}
	roomsMu.Unlock()

	if !ok {
		NotFound(w, r)
		return
	}

	notify(rm)
	http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
}
