package main

import (
	"io/fs"
	"log"
	"net/http"
	"power4"
	"power4/internal/auth"
	httphandler "power4/internal/http"
)

func main() {
	templatesFS, _ := fs.Sub(power4.Content, "templates")
	httphandler.SetTemplateFS(templatesFS)

	if err := auth.InitSessions("data"); err != nil {
		log.Fatal(err)
	}

	store, err := auth.Open("data")
	if err != nil {
		log.Fatal(err)
	}
	httphandler.SetUserStore(store)

	if err := httphandler.InitFriendsStore("data/friends.json"); err != nil {
		log.Printf("friends load error: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/leaderboard", httphandler.ShowLeaderboard)

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			httphandler.ShowSignup(w, r)
			return
		}
		if r.Method == http.MethodPost {
			httphandler.DoSignup(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			httphandler.ShowLogin(w, r)
			return
		}
		if r.Method == http.MethodPost {
			httphandler.DoLogin(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/logout", httphandler.DoLogout)
	mux.HandleFunc("/u/", httphandler.ShowProfile)
	mux.HandleFunc("/rooms/create", httphandler.CreateRoom)
	mux.HandleFunc("/rooms/join", httphandler.JoinRoom)
	mux.HandleFunc("/game/", httphandler.ShowGame)
	mux.HandleFunc("/board/", httphandler.ShowBoard)
	mux.HandleFunc("/clock/", httphandler.ShowClock)
	mux.HandleFunc("/play/", httphandler.Play)
	mux.HandleFunc("/rematch/", httphandler.Rematch)
	mux.HandleFunc("/match/join", httphandler.JoinRandom)
	mux.HandleFunc("/match/check/", httphandler.CheckMatch)
	mux.HandleFunc("/match/leave/", httphandler.LeaveMatch)
	mux.HandleFunc("/match/", httphandler.ShowMatch)
	mux.HandleFunc("/friends", httphandler.ShowFriends)
	mux.HandleFunc("/friends/request", httphandler.SendFriendRequest)
	mux.HandleFunc("/friends/request/accept", httphandler.AcceptFriendRequest)
	mux.HandleFunc("/friends/request/decline", httphandler.DeclineFriendRequest)
	mux.HandleFunc("/friends/challenge", httphandler.SendChallenge)
	mux.HandleFunc("/friends/challenge/accept", httphandler.AcceptChallenge)
	mux.HandleFunc("/friends/challenge/decline", httphandler.DeclineChallenge)
	mux.HandleFunc("/friends/challenge/cancel", httphandler.CancelChallenge)
	mux.HandleFunc("/friends/challenge/wait/", httphandler.ShowChallengeWait)
	mux.HandleFunc("/friends/challenge/check/", httphandler.CheckChallenge)
	mux.HandleFunc("/friends/section/requests", httphandler.ShowFriendsRequestsSection)
	mux.HandleFunc("/friends/section/friends", httphandler.ShowFriendsFriendsSection)

	mux.HandleFunc("/training", httphandler.ShowTraining)
	mux.HandleFunc("/training/start", httphandler.StartTraining)

	staticFS, _ := fs.Sub(power4.Content, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			httphandler.ShowHome(w, r)
			return
		}
		httphandler.NotFound(w, r)
	})

	log.Println("Server started on :8090")
	if err := http.ListenAndServe(":8090", mux); err != nil {
		log.Fatal(err)
	}
}
