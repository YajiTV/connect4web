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
	Grid  [Rows][Cols]Cell
	Moves int
}

func NewBoard() Board {
	return Board{}
}

func IsFull(b *Board) bool {
	return b.Moves >= Rows*Cols
}

func IsGameWon(board *Board) (Cell, bool) {
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
