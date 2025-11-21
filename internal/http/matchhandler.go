package httphandler

import (
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"

	"power4/internal/auth"
	"power4/internal/game"
	"power4/internal/util"
)

// token generates a random base32 ticket id
func token() string {
	id, err := util.RandBase32(16)
	if err != nil {
		return ""
	}
	return id
}

// rangeFor returns the current matchmaking Elo range for a waiter and grows it over time
func rangeFor(w *waiter) int {
	el := time.Since(w.Created)
	base := 50
	step := 50
	incr := int(el.Seconds()) / 5
	r := base + incr*step
	if r > 800 {
		r = 800
	}
	return r
}

// abs returns the absolute value of x
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// removeTicket removes a ticket from both queues and indexes
func removeTicket(t string) {
	mmMu.Lock()
	for i, wq := range waiting {
		if wq.Ticket == t {
			delete(tickets, wq.Ticket)
			waiting = append(waiting[:i], waiting[i+1:]...)
			break
		}
	}
	delete(tickets, t)
	mmMu.Unlock()
}

// purgeUserFromQueue removes all queue entries for a username
func purgeUserFromQueue(username string) {
	mmMu.Lock()
	i := 0
	for i < len(waiting) {
		if waiting[i].Username == username {
			delete(tickets, waiting[i].Ticket)
			waiting = append(waiting[:i], waiting[i+1:]...)
			continue
		}
		i++
	}
	mmMu.Unlock()
}

// JoinRandom enqueues the user for matchmaking or pairs them immediately if a compatible opponent exists
func JoinRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u := auth.CurrentUser(userStore, r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	up := userStore.GetByUsername(u.Username)
	if up == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pid := getOrSetPID(w, r)

	// tries to find a compatible opponent
	mmMu.Lock()
	for _, op := range waiting {
		if op.PID == pid || op.Username == u.Username {
			continue
		}
		if abs(op.Elo-up.Elo) <= rangeFor(op) {
			code := genCode()
			now := time.Now()
			rm := &Room{
				Code:         code,
				Game:         game.NewGame(),
				Player1ID:    op.PID,
				Player2ID:    pid,
				Player1User:  op.Username,
				Player2User:  u.Username,
				CreatedAt:    now,
				Rev:          1,
				subs:         make(map[chan struct{}]struct{}),
				Random:       true,
				TurnDeadline: now.Add(2 * time.Minute),
				StartNext:    game.Player2,
			}
			rm.Game.Player1Name = op.Username
			rm.Game.Player2Name = u.Username

			roomsMu.Lock()
			rooms[code] = rm
			roomsMu.Unlock()

			// removes matched opponent from queue
			for i, wq := range waiting {
				if wq.Ticket == op.Ticket {
					waiting = append(waiting[:i], waiting[i+1:]...)
					break
				}
			}
			delete(tickets, op.Ticket)

			// notifies opponent and redirects
			select {
			case op.Ch <- code:
			default:
			}
			mmMu.Unlock()
			http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
			return
		}
	}

	// reuses existing ticket if already queued
	for _, wq := range waiting {
		if wq.Username == u.Username {
			t := wq.Ticket
			mmMu.Unlock()
			http.Redirect(w, r, "/match/"+t, http.StatusSeeOther)
			return
		}
	}

	// enqueues a new waiter
	t := token()
	wr := &waiter{
		Ticket:   t,
		PID:      pid,
		Username: u.Username,
		Elo:      up.Elo,
		Ch:       make(chan string, 1),
		Created:  time.Now(),
	}
	waiting = append(waiting, wr)
	tickets[t] = wr
	mmMu.Unlock()

	http.Redirect(w, r, "/match/"+t, http.StatusSeeOther)
}

// ShowMatch shows the current matchmaking status and search range
func ShowMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	t := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if t == "" {
		NotFound(w, r)
		return
	}

	mmMu.Lock()
	wr := tickets[t]
	mmMu.Unlock()
	if wr == nil {
		NotFound(w, r)
		return
	}

	rng := rangeFor(wr)
	min2 := wr.Elo - rng
	max2 := wr.Elo + rng

	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "match.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := makeHeader(w, r)
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		Ticket           string
		YourElo          int
		MinElo           int
		MaxElo           int
		LoggedIn         bool
		Username         string
		Initials         string
		CSRF             string
		HasFriendAlerts  bool
		FriendAlertCount int
	}{
		Ticket:           t,
		YourElo:          wr.Elo,
		MinElo:           min2,
		MaxElo:           max2,
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		CSRF:             h.CSRF,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
	})
}

// CheckMatch waits for a match, pairs if a compatible opponent appears, or refreshes
func CheckMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	t := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if t == "" {
		NotFound(w, r)
		return
	}

	mmMu.Lock()
	wr := tickets[t]
	if wr == nil {
		mmMu.Unlock()
		NotFound(w, r)
		return
	}

	// tries to pair immediately
	rng := rangeFor(wr)
	for _, op := range waiting {
		if op.Ticket == wr.Ticket {
			continue
		}
		if op.Username == wr.Username || op.PID == wr.PID {
			continue
		}
		if abs(op.Elo-wr.Elo) <= rng {
			code := genCode()
			now := time.Now()
			rm := &Room{
				Code:         code,
				Game:         game.NewGame(),
				Player1ID:    op.PID,
				Player2ID:    wr.PID,
				Player1User:  op.Username,
				Player2User:  wr.Username,
				CreatedAt:    now,
				Rev:          1,
				subs:         make(map[chan struct{}]struct{}),
				Random:       true,
				TurnDeadline: now.Add(2 * time.Minute),
				StartNext:    game.Player2,
			}
			rm.Game.Player1Name = op.Username
			rm.Game.Player2Name = wr.Username

			roomsMu.Lock()
			rooms[code] = rm
			roomsMu.Unlock()

			// remove both from queues
			for i, wq := range waiting {
				if wq.Ticket == op.Ticket {
					waiting = append(waiting[:i], waiting[i+1:]...)
					break
				}
			}
			for i, wq := range waiting {
				if wq.Ticket == wr.Ticket {
					waiting = append(waiting[:i], waiting[i+1:]...)
					break
				}
			}
			delete(tickets, op.Ticket)
			delete(tickets, wr.Ticket)

			select {
			case op.Ch <- code:
			default:
			}
			mmMu.Unlock()
			http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
			return
		}
	}

	// falls back to long‑poll
	ch := wr.Ch
	mmMu.Unlock()
	select {
	case code := <-ch:
		http.Redirect(w, r, "/game/"+code, http.StatusSeeOther)
		return
	case <-time.After(25 * time.Second):
		http.Redirect(w, r, "/match/"+t+"?ts="+strconvI(int(time.Now().Unix())), http.StatusSeeOther)
		return
	case <-r.Context().Done():
		removeTicket(t)
		return
	}
}

// LeaveMatch removes the ticket for this session and clears any of the user's pending entries
func LeaveMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !auth.CheckCSRF(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t := path.Base(strings.TrimSuffix(r.URL.Path, "/"))
	if t != "" {
		removeTicket(t)
	}

	u := auth.CurrentUser(userStore, r)
	if u != nil {
		purgeUserFromQueue(u.Username)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
