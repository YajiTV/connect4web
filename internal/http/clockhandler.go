package httphandler

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"power4/internal/game"
)

func ShowClock(w http.ResponseWriter, r *http.Request) {
	code := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	roomsMu.RLock()
	rm := rooms[code]
	roomsMu.RUnlock()
	if rm == nil {
		NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store, max-age=0, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	now := time.Now()
	overNow := false
	scoreA := 0.5

	roomsMu.Lock()
	if ready(rm) && !rm.Game.Over {
		if rm.TurnDeadline.IsZero() {
			rm.TurnDeadline = now.Add(2 * time.Minute)
			rm.Rev++
		} else if now.After(rm.TurnDeadline) {
			if rm.Game.NextPlayer == game.Player1 {
				rm.Game.Over = true
				rm.Game.Winner = game.Player2
				scoreA = 0
			} else {
				rm.Game.Over = true
				rm.Game.Winner = game.Player1
				scoreA = 1
			}
			rm.Forfeit = "Time limit exceeded"
			rm.Rev++
			overNow = true
		}
	}
	remaining := time.Until(rm.TurnDeadline)
	if remaining < 0 {
		remaining = 0
	}
	roomsMu.Unlock()

	if overNow && userStore != nil && rm.Player1User != "" && rm.Player2User != "" {
		_ = userStore.ApplyMatch(rm.Player1User, rm.Player2User, scoreA, 32)
		notify(rm)
	}

	min2 := int(remaining.Minutes()) % 60
	sec := int(remaining.Seconds()) % 60
	tleft := ""
	if rm.Game.Over || !ready(rm) {
		tleft = "--:--"
	} else {
		tleft = fmt.Sprintf("%02d:%02d", min2, sec)
	}

	refreshURL := "/clock/" + rm.Code + "?ts=" + strconvI(int(time.Now().UnixNano()))

	tmpl, err := template.ParseFS(templateFS, "clock.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = tmpl.ExecuteTemplate(w, "clock", struct {
		Code       string
		TimeLeft   string
		RefreshURL string
	}{
		Code:       rm.Code,
		TimeLeft:   tleft,
		RefreshURL: refreshURL,
	})
}

func strconvI(n int) string {
	return strconv.FormatInt(int64(n), 10)
}
