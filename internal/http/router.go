package httphandler

import (
	"io/fs"
	nethttp "net/http"
)

// NewRouter wires all routes and static handlers
func NewRouter(staticFS fs.FS) *nethttp.ServeMux {
	mux := nethttp.NewServeMux()

	// leaderboard
	mux.HandleFunc("/leaderboard", ShowLeaderboard)

	// Rules
	mux.HandleFunc("/rules", ShowRules)

	// signup
	mux.HandleFunc("/signup", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method == nethttp.MethodGet {
			ShowSignup(w, r)
			return
		}
		if r.Method == nethttp.MethodPost {
			DoSignup(w, r)
			return
		}
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
	})

	// login
	mux.HandleFunc("/login", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Method == nethttp.MethodGet {
			ShowLogin(w, r)
			return
		}
		if r.Method == nethttp.MethodPost {
			DoLogin(w, r)
			return
		}
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
	})

	// auth and profiles
	mux.HandleFunc("/logout", DoLogout)
	mux.HandleFunc("/u/", ShowProfile)

	// rooms and gameplay
	mux.HandleFunc("/rooms/create", CreateRoom)
	mux.HandleFunc("/rooms/join", JoinRoom)
	mux.HandleFunc("/game/", ShowGame)
	mux.HandleFunc("/board/", ShowBoard)
	mux.HandleFunc("/clock/", ShowClock)
	mux.HandleFunc("/play/", Play)
	mux.HandleFunc("/rematch/", Rematch)

	// random matchmaking
	mux.HandleFunc("/match/join", JoinRandom)
	mux.HandleFunc("/match/check/", CheckMatch)
	mux.HandleFunc("/match/leave/", LeaveMatch)
	mux.HandleFunc("/match/", ShowMatch)

	// friends and challenges
	mux.HandleFunc("/friends", ShowFriends)
	mux.HandleFunc("/friends/request", SendFriendRequest)
	mux.HandleFunc("/friends/request/accept", AcceptFriendRequest)
	mux.HandleFunc("/friends/request/decline", DeclineFriendRequest)
	mux.HandleFunc("/friends/challenge", SendChallenge)
	mux.HandleFunc("/friends/challenge/accept", AcceptChallenge)
	mux.HandleFunc("/friends/challenge/decline", DeclineChallenge)
	mux.HandleFunc("/friends/challenge/cancel", CancelChallenge)
	mux.HandleFunc("/friends/challenge/wait/", ShowChallengeWait)
	mux.HandleFunc("/friends/challenge/check/", CheckChallenge)
	mux.HandleFunc("/friends/section/requests", ShowFriendsRequestsSection)
	mux.HandleFunc("/friends/section/friends", ShowFriendsFriendsSection)

	// training mode
	mux.HandleFunc("/training", ShowTraining)
	mux.HandleFunc("/training/start", StartTraining)

	// static assets
	mux.Handle("/static/", nethttp.StripPrefix("/static/", NewStaticHandler(staticFS)))
	mux.HandleFunc("/static/css/assets/connect4.png", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		nethttp.Redirect(w, r, "/static/assets/connect4.png", nethttp.StatusMovedPermanently)
	})

	// home and 404
	mux.HandleFunc("/", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.URL.Path == "/" {
			ShowHome(w, r)
			return
		}
		NotFound(w, r)
	})

	return mux
}
