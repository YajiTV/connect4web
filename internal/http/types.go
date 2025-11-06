package httphandler

import (
	"sync"
	"time"

	"power4/internal/auth"
	"power4/internal/game"
)

type Room struct {
	Code         string
	Game         *game.Game
	Player1ID    string
	Player2ID    string
	Player1User  string
	Player2User  string
	CreatedAt    time.Time
	Rev          int
	subs         map[chan struct{}]struct{}
	Random       bool
	RematchP1    bool
	RematchP2    bool
	Forfeit      string
	TurnDeadline time.Time
	StartNext    game.Cell
	Bot          bool
	BotLevel     int
}

var (
	rooms     = make(map[string]*Room)
	roomsMu   sync.RWMutex
	userStore *auth.Store
)

func SetUserStore(s *auth.Store) { userStore = s }

type waiter struct {
	Ticket   string
	PID      string
	Username string
	Elo      int
	Ch       chan string
	Created  time.Time
}

var (
	mmMu    sync.Mutex
	waiting []*waiter
	tickets = make(map[string]*waiter)
)
