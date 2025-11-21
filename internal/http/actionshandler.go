package httphandler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"power4/internal/game"
)

// Play handles a POST move, validates turn and column, updates room state, and redirects back to the board
func Play(w http.ResponseWriter, r *http.Request) {
	// rejects non‑POST methods
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// tries to resolve the room from the URL
	rm, code := roomFromPath(r.URL.Path)
	if rm == nil {
		NotFound(w, r)
		return
	}

	// rejects moves until both players are ready
	if !ready(rm) {
		http.Redirect(w, r, "/board/"+code, http.StatusSeeOther)
		return
	}

	// enforces turn ownership per player id
	pid := getOrSetPID(w, r)
	if (rm.Game.NextPlayer == game.Player1 && rm.Player1ID != pid) ||
		(rm.Game.NextPlayer == game.Player2 && rm.Player2ID != pid) {
		http.Redirect(w, r, "/board/"+code, http.StatusSeeOther)
		return
	}

	// parses target column and tries to play
	col, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("column")))
	if err := game.Play(rm.Game, col); err == nil {
		// updates revision and turn deadline
		roomsMu.Lock()
		rm.Rev++
		rm.TurnDeadline = time.Now().Add(2 * time.Minute)
		roomsMu.Unlock()

		// generates Elo update if game just ended and both users are known
		if rm.Game.Over && userStore != nil && rm.Player1User != "" && rm.Player2User != "" {
			scoreA := 0.5
			if rm.Game.Winner == game.Player1 {
				scoreA = 1
			} else if rm.Game.Winner == game.Player2 {
				scoreA = 0
			}
			_ = userStore.ApplyMatch(rm.Player1User, rm.Player2User, scoreA, 32)
		}
	}

	// notifies listeners and redirects back to the board with state hints
	notify(rm)
	http.Redirect(w, r, "/board/"+code+"?rev="+strconv.Itoa(rm.Rev)+"&immediate=1&m=1&hold=1", http.StatusSeeOther)
}
