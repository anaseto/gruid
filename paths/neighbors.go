package paths

import "github.com/anaseto/gruid"

// Pather is the interface used by algorithms that only need neighbor
// information. It's the minimal interface that allows to build paths.
type Pather interface {
	// Neighbors returns the available neighbor positions of a given
	// position. Implementations may use a cache to avoid allocations.
	Neighbors(gruid.Point) []gruid.Point
}

// Neighbors fetches adjacent positions. Its methods return a cached slice for
// efficiency, so results are invalidated by next method calls. It is suitable
// for use in satisfying the Dijkstra, Astar and Pather interfaces.
type Neighbors struct {
	ps []gruid.Point
}

// All returns 8 adjacent positions, including diagonal ones, filtered by keep
// function.
func (nb *Neighbors) All(p gruid.Point, keep func(gruid.Point) bool) []gruid.Point {
	nb.ps = nb.ps[:0]
	for y := -1; y <= 1; y++ {
		for x := -1; x <= 1; x++ {
			if x == 0 && y == 0 {
				continue
			}
			q := p.Shift(x, y)
			if keep(q) {
				nb.ps = append(nb.ps, q)
			}
		}
	}
	return nb.ps
}

// Cardinal returns 4 adjacent cardinal positions, excluding diagonal ones,
// filtered by keep function.
func (nb *Neighbors) Cardinal(p gruid.Point, keep func(gruid.Point) bool) []gruid.Point {
	nb.ps = nb.ps[:0]
	for i := -1; i <= 1; i += 2 {
		q := p.Shift(i, 0)
		if keep(q) {
			nb.ps = append(nb.ps, q)
		}
		q = p.Shift(0, i)
		if keep(q) {
			nb.ps = append(nb.ps, q)
		}
	}
	return nb.ps
}

// Diagonal returns 4 adjacent diagonal (inter-cardinal) positions, filtered by
// keep function.
func (nb *Neighbors) Diagonal(p gruid.Point, keep func(gruid.Point) bool) []gruid.Point {
	nb.ps = nb.ps[:0]
	for y := -1; y <= 1; y += 2 {
		for x := -1; x <= 1; x += 2 {
			q := p.Shift(x, y)
			if keep(q) {
				nb.ps = append(nb.ps, q)
			}
		}
	}
	return nb.ps
}
