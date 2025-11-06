package httphandler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"power4/internal/game"
)

func Play(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	rm, code := roomFromPath(r.URL.Path)
	if rm == nil {
		NotFound(w, r)
		return
	}
	if !ready(rm) {
		http.Redirect(w, r, "/board/"+code, http.StatusSeeOther)
		return
	}

	pid := getOrSetPID(w, r)
	if (rm.Game.NextPlayer == game.Player1 && rm.Player1ID != pid) ||
		(rm.Game.NextPlayer == game.Player2 && rm.Player2ID != pid) {
		http.Redirect(w, r, "/board/"+code, http.StatusSeeOther)
		return
	}

	col, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("column")))
	if err := game.Play(rm.Game, col); err == nil {
		roomsMu.Lock()
		rm.Rev++
		rm.TurnDeadline = time.Now().Add(2 * time.Minute)
		roomsMu.Unlock()
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
	notify(rm)
	http.Redirect(w, r, "/board/"+code+"?rev="+strconv.Itoa(rm.Rev)+"&immediate=1&m=1&hold=1", http.StatusSeeOther)
}
