package game

const (
	Cols  = 7
	Rows  = 6
	toWin = 4
)

type Cell uint8

const (
	Empty Cell = iota
	Player1
	Player2
)

type Board struct {
	Grid  [Rows][Cols]Cell // board cells in row‑major order
	Moves int              // total number of pieces placed so far
}

// NewBoard creates an empty board with zero moves
func NewBoard() Board {
	return Board{}
}

// IsFull returns whether the board has no remaining moves
func IsFull(b *Board) bool {
	return b.Moves >= Rows*Cols
}

// IsGameWon checks horizontal, vertical, and diagonal lines and returns the winner if any
func IsGameWon(board *Board) (Cell, bool) {
	// checks horizontal sequences
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-toWin; c++ {
			p := board.Grid[r][c]
			if p != Empty &&
				board.Grid[r][c+1] == p &&
				board.Grid[r][c+2] == p &&
				board.Grid[r][c+3] == p {
				return p, true
			}
		}
	}

	// checks vertical sequences
	for c := 0; c < Cols; c++ {
		for r := 0; r <= Rows-toWin; r++ {
			p := board.Grid[r][c]
			if p != Empty &&
				board.Grid[r+1][c] == p &&
				board.Grid[r+2][c] == p &&
				board.Grid[r+3][c] == p {
				return p, true
			}
		}
	}

	// checks diagonal down‑right sequences
	for r := 0; r <= Rows-toWin; r++ {
		for c := 0; c <= Cols-toWin; c++ {
			p := board.Grid[r][c]
			if p != Empty &&
				board.Grid[r+1][c+1] == p &&
				board.Grid[r+2][c+2] == p &&
				board.Grid[r+3][c+3] == p {
				return p, true
			}
		}
	}

	// checks diagonal up‑right sequences
	for r := toWin - 1; r < Rows; r++ {
		for c := 0; c <= Cols-toWin; c++ {
			p := board.Grid[r][c]
			if p != Empty &&
				board.Grid[r-1][c+1] == p &&
				board.Grid[r-2][c+2] == p &&
				board.Grid[r-3][c+3] == p {
				return p, true
			}
		}
	}

	return Empty, false
}
