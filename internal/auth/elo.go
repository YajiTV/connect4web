package auth

import "math"

// expected returns a float representing the favourite player using the Elo formula. (0.5 = equal chances, >0.5 favourite, <0.5 outsider)
func expected(ra, rb int) float64 {
	return 1 / (1 + math.Pow(10, float64(rb-ra)/400))
}

// round returns f rounded to the nearest integer with 0.5 cases away from zero
func round(f float64) float64 {
	if f >= 0 {
		return math.Floor(f + 0.5)
	}
	return math.Ceil(f - 0.5)
}
