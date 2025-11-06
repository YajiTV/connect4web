package httphandler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"power4/internal/game"
)

func ShowBoard(w http.ResponseWriter, r *http.Request) {
	rm, code := roomFromPath(r.URL.Path)
	if rm == nil || code == "" {
		NotFound(w, r)
		return
	}

	pid := getOrSetPID(w, r)

	q := r.URL.Query()
	qrev := strings.TrimSpace(q.Get("rev"))
	immediate := q.Get("immediate") == "1"
	forceNew := q.Get("m") == "1"
	hold := r.URL.Query().Get("hold") == "1"

	cur := 0
	if qrev != "" {
		if n, err := strconv.Atoi(qrev); err == nil {
			cur = n
		}
	}

	if rm.Bot && !rm.Game.Over && rm.Player2ID == "BOT" && rm.Game.NextPlayer == game.Player2 && !hold {
		col := game.ComputeBotMove(&rm.Game.Board, game.Player2, rm.BotLevel)
		if col >= 0 {
			_ = game.Play(rm.Game, col)
			roomsMu.Lock()
			rm.Rev++
			rm.TurnDeadline = time.Now().Add(2 * time.Minute)
			roomsMu.Unlock()
			notify(rm)
			http.Redirect(w, r, "/board/"+code+"?rev="+strconv.Itoa(rm.Rev)+"&immediate=1&m=1", http.StatusSeeOther)
			return
		}
	}

	isNewMove := false

	if !immediate && cur >= rm.Rev {
		ch, unsub := subscribe(rm)
		defer unsub()
		select {
		case <-ch:
			isNewMove = true
		case <-time.After(25 * time.Second):
		case <-r.Context().Done():
			return
		}
	}

	w.Header().Set("Cache-Control", "no-store")

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
	roomsMu.Unlock()
	if overNow && userStore != nil && rm.Player1User != "" && rm.Player2User != "" {
		_ = userStore.ApplyMatch(rm.Player1User, rm.Player2User, scoreA, 32)
		notify(rm)
	}

	canPlay := false
	if ready(rm) && !rm.Game.Over {
		if rm.Game.NextPlayer == game.Player1 && rm.Player1ID == pid {
			canPlay = true
		}
		if rm.Game.NextPlayer == game.Player2 && rm.Player2ID == pid {
			canPlay = true
		}
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"Iterate":      Iterate,
		"NextEmptyRow": NextEmptyRow,
		"DropDuration": func(row int) int { return 280 + row*120 },
	}).ParseFS(templateFS, "board.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := makeHeader(w, r)
	self := 0
	if pid == rm.Player1ID {
		self = 1
	} else if pid == rm.Player2ID {
		self = 2
	}

	lastPlayer := game.Cell(0)
	if rm.Game.NextPlayer == game.Player1 {
		lastPlayer = game.Player2
	} else if rm.Game.NextPlayer == game.Player2 {
		lastPlayer = game.Player1
	}

	if qrev != "" && cur < rm.Rev {
		isNewMove = true
	}
	if immediate && self != 0 && lastPlayer == game.Cell(self) {
		isNewMove = true
	}
	if forceNew {
		isNewMove = true
	}

	ref := r.Referer()
	fromPlay := strings.Contains(ref, "/play/")
	fromRematch := strings.Contains(ref, "/rematch/")

	if fromRematch {
		isNewMove = false
	} else if fromPlay && self != 0 && lastPlayer == game.Cell(self) {
		isNewMove = true
	}

	validLast := rm.Game.LastRow >= 0 && rm.Game.LastCol >= 0

	data := struct {
		Code       string
		Rev        int
		Grid       [game.Rows][game.Cols]game.Cell
		NextPlayer game.Cell
		Over       bool
		Winner     game.Cell
		P1Name     string
		P2Name     string
		Ready      bool
		CanPlay    bool
		LoggedIn   bool
		Username   string
		Initials   string
		CSRF       string
		Forfeit    string
		RematchP1  bool
		RematchP2  bool
		Self       int
		LastRow    int
		LastCol    int
		LastPlayer game.Cell
		IsNewMove  bool
		HasLast    bool
	}{
		Code:       rm.Code,
		Rev:        rm.Rev,
		Grid:       rm.Game.Board.Grid,
		NextPlayer: rm.Game.NextPlayer,
		Over:       rm.Game.Over,
		Winner:     rm.Game.Winner,
		P1Name:     rm.Game.Player1Name,
		P2Name:     rm.Game.Player2Name,
		Ready:      ready(rm),
		CanPlay:    canPlay,
		LoggedIn:   h.LoggedIn,
		Username:   h.Username,
		Initials:   h.Initials,
		CSRF:       h.CSRF,
		Forfeit:    rm.Forfeit,
		RematchP1:  rm.RematchP1,
		RematchP2:  rm.RematchP2,
		Self:       self,
		LastRow:    rm.Game.LastRow,
		LastCol:    rm.Game.LastCol,
		LastPlayer: lastPlayer,
		IsNewMove:  isNewMove && validLast,
		HasLast:    validLast,
	}
	_ = tmpl.ExecuteTemplate(w, "board", data)
}
