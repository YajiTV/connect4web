package game

import "errors"

var (
	// ErrColOutOfRange indicates the column index is out of range
	ErrColOutOfRange = errors.New("column out of range")

	// ErrColFull indicates the column is full
	ErrColFull = errors.New("column full")

	// ErrInvalidPlayer indicates an invalid player value was used
	ErrInvalidPlayer = errors.New("invalid player")

	// ErrGameOver indicates the game is already over
	ErrGameOver = errors.New("game over")
)

type Game struct {
	Board       Board  // current board state
	NextPlayer  Cell   // player who plays next
	Over        bool   // whether the game is over
	Winner      Cell   // winner when Over is true, or Empty for draw
	Player1Name string // display name for player 1
	Player2Name string // display name for player 2
	LastRow     int    // row of the last move, -1 if none
	LastCol     int    // column of the last move, -1 if none
}

// NewGame creates a new game with an empty board and default names
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

// ToggleTurn switches the turn to the other player
func ToggleTurn(g *Game) {
	if g.NextPlayer == Player1 {
		g.NextPlayer = Player2
	} else {
		g.NextPlayer = Player1
	}
}

// Play tries to apply a move in col, updates game state, and switches turn
func Play(g *Game, col int) error {
	// rejects moves after game is over
	if g.Over {
		return ErrGameOver
	}

	// drops a piece and records the last move
	row, err := AddPeon(&g.Board, col, g.NextPlayer)
	if err != nil {
		return err
	}
	g.LastRow = row
	g.LastCol = col

	// checks for a win
	if w, ok := IsGameWon(&g.Board); ok {
		g.Winner, g.Over = w, true
		return nil
	}

	// checks for a draw
	if IsFull(&g.Board) {
		g.Over = true
		return nil
	}

	// continues the game by switching turns
	ToggleTurn(g)
	return nil
}

// Reset resets the game while preserving player names
func Reset(g *Game) {
	p1, p2 := g.Player1Name, g.Player2Name
	*g = *NewGame()
	g.Player1Name, g.Player2Name = p1, p2
}
