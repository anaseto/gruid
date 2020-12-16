package paths

import "github.com/anaseto/gruid"

// Neighborer is the interface used by algorithmes that only need neighbor
// information.
type Neighborer interface {
	// Neighbors returns the available neighbor positions of a given
	// position. Implementations may use a cache to avoid allocations.
	Neighbors(gruid.Position) []gruid.Position
}

// Neighbors fetches adjacent positions. It returns a cached slice for
// efficiency, so results are invalidated by next method calls. It is suitable
// for use in satisfying the Dijkstra, Astar and Neighborer interfaces.
type Neighbors struct {
	nb []gruid.Position
}

// All returns 8 adjacent positions, including diagonal ones, filtered by keep
// function.
func (nf *Neighbors) All(pos gruid.Position, keep func(gruid.Position) bool) []gruid.Position {
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

// Cardinal returns 4 adjacent cardinal positions, excluding diagonal ones,
// filtered by keep function.
func (nf *Neighbors) Cardinal(pos gruid.Position, keep func(gruid.Position) bool) []gruid.Position {
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

// Diagonal returns 4 adjacent diagonal (inter-cardinal) positions, filtered by
// keep function.
func (nf *Neighbors) Diagonal(pos gruid.Position, keep func(gruid.Position) bool) []gruid.Position {
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
