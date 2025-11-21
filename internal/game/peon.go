package game

// AddPeon adds a piece to the given column for the specified player and returns the row index
func AddPeon(board *Board, col int, cell Cell) (int, error) {
	// rejects columns outside bounds
	if col < 0 || col >= Cols {
		return -1, ErrColOutOfRange
	}
	// rejects non playable cell values
	if cell != Player1 && cell != Player2 {
		return -1, ErrInvalidPlayer
	}
	// scans from bottom to top and tries to place the piece
	for r := Rows - 1; r >= 0; r-- {
		if board.Grid[r][col] == Empty {
			board.Grid[r][col] = cell
			board.Moves++ // increments move counter
			return r, nil
		}
	}
	// reports a full column when no slots remain
	return -1, ErrColFull
}
