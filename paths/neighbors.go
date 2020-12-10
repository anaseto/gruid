package paths

import "github.com/anaseto/gruid"

// NeighborFinder returns neighboring positions. It returns a cached slice for
// efficiency, so results are invalidated by next method calls. It is suitable
// for use in the provided Dijkstra and Astar algorithms.
type NeighborFinder struct {
	nb []gruid.Position
}

// Neighbors returns 8 adjacent positions, including diagonal ones, filtered by
// keep function.
func (nf NeighborFinder) Neighbors(pos gruid.Position, keep func(gruid.Position) bool) []gruid.Position {
	nf.nb = nf.nb[:0]
	for y := -1; y <= 1; y++ {
		for x := -1; x <= 1; x++ {
			if x == 0 && y == 0 {
				continue
			}
			npos := pos.Shift(x, y)
			if keep(npos) {
				nf.nb = append(nf.nb, npos)
			}
		}
	}
	return nf.nb
}

// Neighbors returns 4 adjacent cardinal positions, excluding diagonal ones,
// filtered by keep function.
func (nf NeighborFinder) CardinalNeighbors(pos gruid.Position, keep func(gruid.Position) bool) []gruid.Position {
	nf.nb = nf.nb[:0]
	for i := -1; i <= 1; i += 2 {
		npos := pos.Shift(i, 0)
		if keep(npos) {
			nf.nb = append(nf.nb, npos)
		}
		npos = pos.Shift(0, i)
		if keep(npos) {
			nf.nb = append(nf.nb, npos)
		}
	}
	return nf.nb
}

// Neighbors returns 4 adjacent diagonal positions, filtered by keep function.
func (nf NeighborFinder) DiagonalNeighbors(pos gruid.Position, keep func(gruid.Position) bool) []gruid.Position {
	nf.nb = nf.nb[:0]
	for y := -1; y <= 1; y += 2 {
		for x := -1; x <= 1; x += 2 {
			npos := pos.Shift(x, y)
			if keep(npos) {
				nf.nb = append(nf.nb, npos)
			}
		}
	}
	return nf.nb
}
