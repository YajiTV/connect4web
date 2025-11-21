package httphandler

import (
	"sync"
	"time"

	"power4/internal/auth"
	"power4/internal/game"
)

type Room struct {
	Code         string                     // unique room code
	Game         *game.Game                 // game state
	Player1ID    string                     // pid for player 1
	Player2ID    string                     // pid for player 2
	Player1User  string                     // username for player 1
	Player2User  string                     // username for player 2
	CreatedAt    time.Time                  // room creation time
	Rev          int                        // revision counter for client sync
	subs         map[chan struct{}]struct{} // subscribers for long polling
	Random       bool                       // whether the room is from random matchmaking
	RematchP1    bool                       // player 1 rematch consent
	RematchP2    bool                       // player 2 rematch consent
	Forfeit      string                     // reason if ended by forfeit
	TurnDeadline time.Time                  // deadline for the current turn
	StartNext    game.Cell                  // who starts the next game on rematch
	Bot          bool                       // whether this is a bot match
	BotLevel     int                        // bot difficulty level
}

var (
	rooms     = make(map[string]*Room) // all active rooms by code
	roomsMu   sync.RWMutex             // guards rooms
	userStore *auth.Store              // global user store for lookups and ELO updates
)

// SetUserStore sets the global user store reference
func SetUserStore(s *auth.Store) { userStore = s }

type waiter struct {
	Ticket   string      // matchmaking ticket id
	PID      string      // player id cookie
	Username string      // username of the player
	Elo      int         // current elo used for range matching
	Ch       chan string // channel receiving room code when matched
	Created  time.Time   // when the player entered the queue
}

var (
	mmMu    sync.Mutex                 // guards matchmaking structures
	waiting []*waiter                  // queue of waiting players
	tickets = make(map[string]*waiter) // ticket -> waiter
)
