package game

import "errors"

var (
	ErrColOutOfRange = errors.New("column out of range")
	ErrColFull       = errors.New("column full")
	ErrInvalidPlayer = errors.New("invalid player")
	ErrGameOver      = errors.New("game over")
)

type Game struct {
	Board       Board
	NextPlayer  Cell
	Over        bool
	Winner      Cell
	Player1Name string
	Player2Name string
	LastRow     int
	LastCol     int
}

func NewGame() *Game {
	return &Game{
		Board:       NewBoard(),
		NextPlayer:  Player1,
		Winner:      Empty,
		Over:        false,
		Player1Name: "Player 1",
		Player2Name: "Player 2",
		LastRow:     -1,
		LastCol:     -1,
	}
}

func ToggleTurn(g *Game) {
	if g.NextPlayer == Player1 {
		g.NextPlayer = Player2
	} else {
		g.NextPlayer = Player1
	}
}

func Play(g *Game, col int) error {
	if g.Over {
		return ErrGameOver
	}
	row, err := AddPeon(&g.Board, col, g.NextPlayer)
	if err != nil {
		return err
	}
	g.LastRow = row
	g.LastCol = col
	if w, ok := IsGameWon(&g.Board); ok {
		g.Winner, g.Over = w, true
		return nil
	}
	if IsFull(&g.Board) {
		g.Over = true
		return nil
	}
	ToggleTurn(g)
	return nil
}

func Reset(g *Game) {
	p1, p2 := g.Player1Name, g.Player2Name
	*g = *NewGame()
	g.Player1Name, g.Player2Name = p1, p2
}
