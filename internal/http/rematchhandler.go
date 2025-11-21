package httphandler

import (
	"net/http"
	"strconv"
	"time"

	"power4/internal/game"
)

// Rematch records rematch consent and starts a new game when both sides agree or vs bot
func Rematch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// resolves room
	rm, code := roomFromPath(r.URL.Path)
	if rm == nil {
		NotFound(w, r)
		return
	}

	// only after the previous game is over
	if !rm.Game.Over {
		http.Redirect(w, r, "/board/"+code, http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)

	roomsMu.Lock()
	changed := false

	// records each player's consent once
	if pid == rm.Player1ID && !rm.RematchP1 {
		rm.RematchP1 = true
		rm.Rev++
		changed = true
	}
	if pid == rm.Player2ID && !rm.RematchP2 {
		rm.RematchP2 = true
		rm.Rev++
		changed = true
	}

	// starts immediately for bot games, otherwise after both consent
	start := false
	if rm.Bot {
		rm.RematchP1 = true
		rm.RematchP2 = true

		game.Reset(rm.Game)
		rm.Game.NextPlayer = rm.StartNext
		if rm.StartNext == game.Player1 {
			rm.StartNext = game.Player2
		} else {
			rm.StartNext = game.Player1
		}
		rm.Forfeit = ""
		rm.RematchP1 = false
		rm.RematchP2 = false
		rm.TurnDeadline = time.Now().Add(2 * time.Minute)
		rm.Rev++
		start = true
	} else if rm.RematchP1 && rm.RematchP2 {
		game.Reset(rm.Game)
		rm.Game.NextPlayer = rm.StartNext
		if rm.StartNext == game.Player1 {
			rm.StartNext = game.Player2
		} else {
			rm.StartNext = game.Player1
		}
		rm.Forfeit = ""
		rm.RematchP1 = false
		rm.RematchP2 = false
		rm.TurnDeadline = time.Now().Add(2 * time.Minute)
		rm.Rev++
		start = true
	}
	roomsMu.Unlock()

	// notifies subscribers and redirects back to board
	if changed || start {
		notify(rm)
	}
	http.Redirect(w, r, "/board/"+code+"?rev="+strconv.Itoa(rm.Rev)+"&immediate=1", http.StatusSeeOther)
}
