package auth

import "math"

func expected(ra, rb int) float64 {
	return 1 / (1 + math.Pow(10, float64(rb-ra)/400))
}

func round(f float64) float64 {
	if f >= 0 {
		return math.Floor(f + 0.5)
	}
	return math.Ceil(f - 0.5)
}
