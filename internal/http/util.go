package httphandler

import (
	"crypto/rand"
	"encoding/base32"
	"path"
	"strings"

	"power4/internal/game"
)

func Iterate(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = i
	}
	return r
}

func NextEmptyRow(grid [game.Rows][game.Cols]game.Cell, col int) int {
	for r := game.Rows - 1; r >= 0; r-- {
		if grid[r][col] == game.Empty {
			return r
		}
	}
	return -1
}

func ready(rm *Room) bool { return rm != nil && rm.Player1ID != "" && rm.Player2ID != "" }

func genCode() string {
	for {
		b := make([]byte, 6)
		_, _ = rand.Read(b)
		code := strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
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

func roomFromPath(p string) (*Room, string) {
	code := path.Base(strings.TrimSuffix(p, "/"))
	roomsMu.RLock()
	rm := rooms[code]
	roomsMu.RUnlock()
	return rm, code
}
