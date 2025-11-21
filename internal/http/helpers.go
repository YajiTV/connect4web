package httphandler

import (
	"path"
	"strings"

	"power4/internal/game"
	"power4/internal/util"
)

// Iterate generates a sequence [0..n-1]
func Iterate(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = i
	}
	return r
}

// NextEmptyRow returns the lowest empty row index in the given column or -1 if full
func NextEmptyRow(grid [game.Rows][game.Cols]game.Cell, col int) int {
	for r := game.Rows - 1; r >= 0; r-- {
		if grid[r][col] == game.Empty {
			return r
		}
	}
	return -1
}

// ready checks whether both player ids are set
func ready(rm *Room) bool { return rm != nil && rm.Player1ID != "" && rm.Player2ID != "" }

// genCode generates a unique 6‑char uppercase room code
func genCode() string {
	for {
		id, err := util.RandBase32(6)
		if err != nil {
			continue
		}
		code := strings.ToUpper(id)
		if len(code) > 6 {
			code = code[:6]
		}

		roomsMu.RLock()
		_, exists := rooms[code]
		roomsMu.RUnlock()
		if !exists {
			return code
		}
	}
}

// roomFromPath resolves the room and code from a URL path
func roomFromPath(p string) (*Room, string) {
	code := path.Base(strings.TrimSuffix(p, "/"))

	roomsMu.RLock()
	rm := rooms[code]
	roomsMu.RUnlock()

	return rm, code
}
