package httphandler

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

// ShowLeaderboard renders the Elo leaderboard with optional username filtering
func ShowLeaderboard(w http.ResponseWriter, r *http.Request) {
	// reads query and fetches users
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	list := userStore.UsersByElo(q)

	// parses template
	tmpl, err := template.ParseFS(templateFS, "base.tmpl", "leaderboard.tmpl")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := makeHeader(w, r)

	// prepares rows with computed win rate
	type row struct {
		Rank     int
		Username string
		Elo      int
		Games    int
		Wins     int
		Losses   int
		WinRate  int // percentage 0..100
	}
	rows := make([]row, 0, len(list))
	for i, u := range list {
		wr := 0
		if u.Games > 0 {
			wr = int((float64(u.Wins) / float64(u.Games)) * 100)
		}
		rows = append(rows, row{
			Rank:     i + 1,
			Username: u.Username,
			Elo:      u.Elo,
			Games:    u.Games,
			Wins:     u.Wins,
			Losses:   u.Losses,
			WinRate:  wr,
		})
	}

	// renders page
	_ = tmpl.ExecuteTemplate(w, "base", struct {
		Query            string
		Rows             []row
		LoggedIn         bool
		Username         string
		Initials         string
		CSRF             string
		HasFriendAlerts  bool
		FriendAlertCount int
	}{
		Query:            q,
		Rows:             rows,
		LoggedIn:         h.LoggedIn,
		Username:         h.Username,
		Initials:         h.Initials,
		CSRF:             h.CSRF,
		HasFriendAlerts:  h.HasFriendAlerts,
		FriendAlertCount: h.FriendAlertCount,
	})
}
