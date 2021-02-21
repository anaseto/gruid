package paths

import "github.com/anaseto/gruid"

// DistanceManhattan computes the taxicab norm (1-norm). See:
// 	https://en.wikipedia.org/wiki/Taxicab_geometry
// It can often be used as A* distance heuristic when 4-way movement is used.
func DistanceManhattan(p, q gruid.Point) int {
	p = p.Sub(q)
	return abs(p.X) + abs(p.Y)
}

// DistanceChebyshev computes the maximum norm (infinity-norm). See:
// 	https://en.wikipedia.org/wiki/Chebyshev_distance
// It can often be used as A* distance heuristic when 8-way movement is used.
func DistanceChebyshev(p, q gruid.Point) int {
	p = p.Sub(q)
	return max(abs(p.X), abs(p.Y))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}
