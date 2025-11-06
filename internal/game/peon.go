package game

func AddPeon(board *Board, col int, cell Cell) (int, error) {
	if col < 0 || col >= Cols {
		return -1, ErrColOutOfRange
	}
	if cell != Player1 && cell != Player2 {
		return -1, ErrInvalidPlayer
	}
	for r := Rows - 1; r >= 0; r-- {
		if board.Grid[r][col] == Empty {
			board.Grid[r][col] = cell
			board.Moves++
			return r, nil
		}
	}
	return -1, ErrColFull
}
