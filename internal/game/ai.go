package game

import (
	"math"
	"math/rand"
	"time"
)

func opponent(p Cell) Cell {
	if p == Player1 {
		return Player2
	}
	return Player1
}

func validMoves(b *Board) []int {
	var out []int
	for c := 0; c < Cols; c++ {
		if b.Grid[0][c] == Empty {
			out = append(out, c)
		}
	}
	return out
}

func apply(b Board, col int, p Cell) (Board, bool) {
	nb := b
	_, err := AddPeon(&nb, col, p)
	if err != nil {
		return b, false
	}
	return nb, true
}

func terminalScore(b *Board, me Cell, depth int) (bool, int) {
	if w, ok := IsGameWon(b); ok {
		if w == me {
			return true, 100000 - depth
		}
		return true, depth - 100000
	}
	if IsFull(b) {
		return true, 0
	}
	return false, 0
}

func countWindow(vals []Cell, me Cell) int {
	opp := opponent(me)
	meCount := 0
	oppCount := 0
	empty := 0
	for _, v := range vals {
		if v == me {
			meCount++
		} else if v == opp {
			oppCount++
		} else {
			empty++
		}
	}
	if meCount == 4 {
		return 10000
	}
	if meCount == 3 && empty == 1 {
		return 100
	}
	if meCount == 2 && empty == 2 {
		return 10
	}
	if oppCount == 3 && empty == 1 {
		return -120
	}
	if oppCount == 2 && empty == 2 {
		return -12
	}
	return 0
}

func eval(b *Board, me Cell) int {
	score := 0
	centerCol := Cols / 2
	centerCount := 0
	for r := 0; r < Rows; r++ {
		if b.Grid[r][centerCol] == me {
			centerCount++
		}
	}
	score += centerCount * 6
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += countWindow([]Cell{b.Grid[r][c], b.Grid[r][c+1], b.Grid[r][c+2], b.Grid[r][c+3]}, me)
		}
	}
	for c := 0; c < Cols; c++ {
		for r := 0; r <= Rows-4; r++ {
			score += countWindow([]Cell{b.Grid[r][c], b.Grid[r+1][c], b.Grid[r+2][c], b.Grid[r+3][c]}, me)
		}
	}
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += countWindow([]Cell{b.Grid[r][c], b.Grid[r+1][c+1], b.Grid[r+2][c+2], b.Grid[r+3][c+3]}, me)
		}
	}
	for r := 3; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += countWindow([]Cell{b.Grid[r][c], b.Grid[r-1][c+1], b.Grid[r-2][c+2], b.Grid[r-3][c+3]}, me)
		}
	}
	return score
}

func immediateWin(b *Board, p Cell) int {
	moves := validMoves(b)
	for _, c := range moves {
		nb, ok := apply(*b, c, p)
		if !ok {
			continue
		}
		if w, ok := IsGameWon(&nb); ok && w == p {
			return c
		}
	}
	return -1
}

func orderMovesCenterFirst(moves []int) []int {
	center := Cols / 2
	type pair struct{ c, w int }
	var ps []pair
	for _, m := range moves {
		ps = append(ps, pair{m, -absInt(m - center)})
	}
	for i := 0; i < len(ps)-1; i++ {
		for j := i + 1; j < len(ps); j++ {
			if ps[j].w > ps[i].w {
				ps[i], ps[j] = ps[j], ps[i]
			}
		}
	}
	out := make([]int, 0, len(moves))
	for _, p := range ps {
		out = append(out, p.c)
	}
	return out
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func minimax(b Board, depth int, alpha, beta int, maximizing bool, me Cell) int {
	if term, sc := terminalScore(&b, me, depth); term {
		return sc
	}
	if depth == 0 {
		return eval(&b, me)
	}
	moves := orderMovesCenterFirst(validMoves(&b))
	if maximizing {
		maxEval := math.MinInt32
		for _, c := range moves {
			nb, ok := apply(b, c, me)
			if !ok {
				continue
			}
			e := minimax(nb, depth-1, alpha, beta, false, me)
			if e > maxEval {
				maxEval = e
			}
			if maxEval > alpha {
				alpha = maxEval
			}
			if beta <= alpha {
				break
			}
		}
		return maxEval
	}
	minEval := math.MaxInt32
	opp := opponent(me)
	for _, c := range moves {
		nb, ok := apply(b, c, opp)
		if !ok {
			continue
		}
		e := minimax(nb, depth-1, alpha, beta, true, me)
		if e < minEval {
			minEval = e
		}
		if minEval < beta {
			beta = minEval
		}
		if beta <= alpha {
			break
		}
	}
	return minEval
}

func pickRandom(b *Board) int {
	ms := validMoves(b)
	if len(ms) == 0 {
		return -1
	}
	rand.Seed(time.Now().UnixNano())
	return ms[rand.Intn(len(ms))]
}

func pickGreedy(b *Board, me Cell) int {
	if c := immediateWin(b, me); c >= 0 {
		return c
	}
	opp := opponent(me)
	if c := immediateWin(b, opp); c >= 0 {
		return c
	}
	moves := validMoves(b)
	best := -1
	bestScore := math.MinInt32
	center := Cols / 2
	for _, c := range moves {
		nb, ok := apply(*b, c, me)
		if !ok {
			continue
		}
		sc := eval(&nb, me) - absInt(center-c)
		if sc > bestScore {
			bestScore = sc
			best = c
		}
	}
	return best
}

func pickMinimax(b *Board, me Cell, depth int) int {
	moves := orderMovesCenterFirst(validMoves(b))
	best := -1
	bestScore := math.MinInt32
	for _, c := range moves {
		nb, ok := apply(*b, c, me)
		if !ok {
			continue
		}
		score := minimax(nb, depth-1, math.MinInt32/2, math.MaxInt32/2, false, me)
		if score > bestScore {
			bestScore = score
			best = c
		}
	}
	return best
}

func ComputeBotMove(b *Board, who Cell, level int) int {
	if level <= 1 {
		return pickRandom(b)
	}
	if level == 2 {
		return pickGreedy(b, who)
	}
	if level == 3 {
		return pickMinimax(b, who, 3)
	}
	if level == 4 {
		return pickMinimax(b, who, 4)
	}
	return pickMinimax(b, who, 5)
}
