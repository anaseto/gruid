// This file implements a JPS (jump-point-search) pathfinding algorithm. For
// more information: https://en.wikipedia.org/wiki/Jump_point_search

package paths

import (
	"github.com/anaseto/gruid"
	//"log"
)

// JPSPath returns a path from a position to another, including these
// positions, in the path order. It uses the given path slice to avoid
// allocations unless its capacity is not enough. The passable function
// controls which positions can be passed. If diags is false, only movements in
// straight cardinal directions are allowed.
//
// The function returns nil if no path was found.
//
// In most situations, JPSPath has significantly better performance than
// AstarPath. The algorithm's limitation is that it only handles uniform costs
// and natural neighbors in grid geometry.
func (pr *PathRange) JPSPath(path []gruid.Point, from, to gruid.Point, passable func(gruid.Point) bool, diags bool) []gruid.Point {
	if !from.In(pr.Rg) || !to.In(pr.Rg) {
		return nil
	}
	if from == to {
		return append(path, from)
	}
	pr.passable = passable
	pr.diags = diags
	path = path[:0]
	pr.initAstar()
	nm := pr.AstarNodes
	nm.Idx++
	defer checkNodesIdx(nm)
	pr.AstarQueue = pr.AstarQueue[:0]
	pqInit(&pr.AstarQueue)
	fromNode := nm.get(pr, from)
	fromNode.Closed = true
	fromNode.Open = false
	neighbors := make([]gruid.Point, 0, 8)
	pr.expandOrigin(from, to)
	//logrid.Map(func(p gruid.Point, c gruid.Cell) gruid.Cell {
	//if passable(p) {
	//return gruid.Cell{Rune: '.'}
	//}
	//return gruid.Cell{Rune: '#'}
	//})
	for {
		if (&pr.AstarQueue).Len() == 0 {
			// There's no path.
			return nil
		}
		n := pqPop(&pr.AstarQueue)
		//logrid.Set(n.P, gruid.Cell{Rune: 'X'})
		n.Open = false
		n.Closed = true

		if n.P == to {
			path = pr.path(path, from, n)
			//logPath(path)
			//log.Printf("\n%v\n", logrid)
			return path
		}

		neighbors = pr.neighbors(neighbors[:0], n, to)
		for _, p := range neighbors {
			dir := p.Sub(n.P)
			var q gruid.Point
			var i int
			if pr.diags {
				q, i = pr.jump(p, dir, to, n.Cost)
			} else {
				q, i = pr.jumpNoDiags(p, dir, to, n.Cost)
			}
			if i > 0 {
				pr.addSuccessor(q, n.P, to, n.Cost+i)
			}
		}
	}
}

func (pr *PathRange) expandOrigin(from, to gruid.Point) {
	for y := -1; y <= 1; y++ {
		for x := -1; x <= 1; x++ {
			if x == 0 && y == 0 {
				continue
			}
			dir := gruid.Point{x, y}
			q := from.Add(dir)
			if !pr.diags {
				if dir.X != 0 && dir.Y != 0 {
					if pr.pass(from.Add(gruid.Point{dir.X, 0})) || pr.pass(from.Add(gruid.Point{0, dir.Y})) {
						pr.addSuccessor(q, from, to, 2)
					}
					continue
				}
				pr.addSuccessor(q, from, to, 1)
				continue
			}
			pr.addSuccessor(q, from, to, 1)
		}
	}
}

func (pr *PathRange) pass(p gruid.Point) bool {
	return p.In(pr.Rg) && pr.passable(p)
}

func (pr *PathRange) obstacle(p gruid.Point) bool {
	return p.In(pr.Rg) && !pr.passable(p)
}

func right(p gruid.Point, dir gruid.Point) gruid.Point {
	return gruid.Point{p.X - dir.Y, p.Y + dir.X}
}

func left(p gruid.Point, dir gruid.Point) gruid.Point {
	return gruid.Point{p.X + dir.Y, p.Y - dir.X}
}

// forcedSucc controls whether a jump may have forced successors left, right,
// on both sides or none. This depends on the jump direction and whether it is
// done on the range edge.
type forcedSucc int

const (
	fsNone forcedSucc = iota
	fsLeft
	fsRight
	fsBoth
)

func (pr *PathRange) straightMax(p, dir gruid.Point) (int, forcedSucc) {
	fs := fsBoth
	max := 0
	switch {
	case dir.X > 0:
		max = pr.Rg.Max.X - p.X
		if p.Y == 0 {
			fs -= fsLeft
		}
		if p.Y == pr.Rg.Max.Y-1 {
			fs -= fsRight
		}
	case dir.X < 0:
		max = -pr.Rg.Min.X + p.X
		if p.Y == 0 {
			fs -= fsRight
		}
		if p.Y == pr.Rg.Max.Y-1 {
			fs -= fsLeft
		}
	case dir.Y > 0:
		max = pr.Rg.Max.Y - p.Y
		if p.X == 0 {
			fs -= fsLeft
		}
		if p.X == pr.Rg.Max.X-1 {
			fs -= fsRight
		}
	case dir.Y < 0:
		max = -pr.Rg.Min.Y + p.Y
		if p.X == 0 {
			fs -= fsRight
		}
		if p.X == pr.Rg.Max.X-1 {
			fs -= fsLeft
		}
	}
	return max, fs
}

func (pr *PathRange) jumpStraight(p, dir, to gruid.Point) (gruid.Point, int) {
	max, fs := pr.straightMax(p, dir)
	switch fs {
	case fsNone:
		return pr.jumpStraightNone(p, dir, to, max)
	case fsRight:
		return pr.jumpStraightRight(p, dir, to, max)
	case fsLeft:
		return pr.jumpStraightLeft(p, dir, to, max)
	default:
		for i := 1; i < max+1; i++ {
			if !pr.passable(p) {
				return p, 0
			}
			if p == to {
				return p, i
			}
			np := p.Add(dir)
			if q := left(p, dir); !pr.passable(q) && pr.pass(q.Add(dir)) {
				return p, i
			}
			if q := right(p, dir); !pr.passable(q) && pr.pass(q.Add(dir)) {
				return p, i
			}
			p = np
		}
		return p, 0
	}
}

func (pr *PathRange) jumpStraightNone(p, dir, to gruid.Point, max int) (gruid.Point, int) {
	for i := 1; i < max+1; i++ {
		if !pr.passable(p) {
			return p, 0
		}
		if p == to {
			return p, i
		}
		p = p.Add(dir)
	}
	return p, 0
}

func (pr *PathRange) jumpStraightLeft(p, dir, to gruid.Point, max int) (gruid.Point, int) {
	for i := 1; i < max+1; i++ {
		if !pr.passable(p) {
			return p, 0
		}
		if p == to {
			return p, i
		}
		np := p.Add(dir)
		if q := left(p, dir); !pr.passable(q) && pr.pass(q.Add(dir)) {
			return p, i
		}
		p = np
	}
	return p, 0
}

func (pr *PathRange) jumpStraightRight(p, dir, to gruid.Point, max int) (gruid.Point, int) {
	for i := 1; i < max+1; i++ {
		if !pr.passable(p) {
			return p, 0
		}
		if p == to {
			return p, i
		}
		np := p.Add(dir)
		if q := right(p, dir); !pr.passable(q) && pr.pass(q.Add(dir)) {
			return p, i
		}
		p = np
	}
	return p, 0
}

func (pr *PathRange) jumpStraightNoDiags(p, dir, to gruid.Point) (gruid.Point, int) {
	max, fs := pr.straightMax(p, dir)
	switch fs {
	case fsNone:
		return pr.jumpStraightNone(p, dir, to, max)
	case fsRight:
		return pr.jumpStraightRightNoDiags(p, dir, to, max)
	case fsLeft:
		return pr.jumpStraightLeftNoDiags(p, dir, to, max)
	default:
		for i := 1; i < max+1; i++ {
			if !pr.passable(p) {
				return p, 0
			}
			if p == to {
				return p, i
			}
			np := p.Add(dir)
			if q := left(p, dir); !pr.passable(q) {
				if pr.pass(q.Add(dir)) && pr.pass(np) {
					return p, i
				}
			}
			if q := right(p, dir); !pr.passable(q) {
				if pr.pass(q.Add(dir)) && pr.pass(np) {
					return p, i
				}
			}
			p = np
		}
		return p, 0
	}
}

func (pr *PathRange) jumpStraightLeftNoDiags(p, dir, to gruid.Point, max int) (gruid.Point, int) {
	for i := 1; i < max+1; i++ {
		if !pr.passable(p) {
			return p, 0
		}
		if p == to {
			return p, i
		}
		np := p.Add(dir)
		if q := left(p, dir); !pr.passable(q) {
			if pr.pass(q.Add(dir)) && pr.pass(np) {
				return p, i
			}
		}
		p = np
	}
	return p, 0
}

func (pr *PathRange) jumpStraightRightNoDiags(p, dir, to gruid.Point, max int) (gruid.Point, int) {
	for i := 1; i < max+1; i++ {
		if !pr.passable(p) {
			return p, 0
		}
		if p == to {
			return p, i
		}
		np := p.Add(dir)
		if q := right(p, dir); !pr.passable(q) {
			if pr.pass(q.Add(dir)) && pr.pass(np) {
				return p, i
			}
		}
		p = np
	}
	return p, 0
}

func (pr *PathRange) jumpDiagonal(p, dir, to gruid.Point, cost int) (gruid.Point, int) {
	i := 1
	from := p.Sub(dir)
	for {
		if !pr.pass(p) {
			return p, 0
		}
		if p == to {
			return p, i
		}
		if q := p.Shift(-dir.X, 0); pr.obstacle(q) {
			if pr.pass(p.Add(gruid.Point{-dir.X, dir.Y})) {
				return p, i
			}
		}
		if q := p.Shift(0, -dir.Y); pr.obstacle(q) {
			if pr.pass(p.Add(gruid.Point{dir.X, -dir.Y})) {
				return p, i
			}
		}
		q, j := pr.jumpStraight(p.Shift(dir.X, 0), gruid.Point{dir.X, 0}, to)
		if j > 0 {
			pr.addSuccessor(q, from, to, cost+i+j)
		}
		q, j = pr.jumpStraight(p.Shift(0, dir.Y), gruid.Point{0, dir.Y}, to)
		if j > 0 {
			pr.addSuccessor(q, from, to, cost+i+j)
		}
		p = p.Add(dir)
		i++
	}
}

func (pr *PathRange) jumpDiagonalNoDiags(p, dir, to gruid.Point, cost int) (gruid.Point, int) {
	i := 2 // diagonals cost 2 (two cardinal movements)
	from := p.Sub(dir)
	for {
		if !pr.pass(p) {
			return p, 0
		}
		px := p.Shift(-dir.X, 0)
		py := p.Shift(0, -dir.Y)
		pxpass := pr.pass(px)
		pypass := pr.pass(py)
		if !pxpass && !pypass {
			return p, 0
		}
		if p == to {
			return p, i
		}
		if !pxpass {
			if pr.pass(p.Add(gruid.Point{-dir.X, dir.Y})) && pr.pass(p.Add(gruid.Point{0, dir.Y})) {
				return p, i
			}
		}
		if !pypass {
			if pr.pass(p.Add(gruid.Point{dir.X, -dir.Y})) && pr.pass(p.Add(gruid.Point{dir.X, 0})) {
				return p, i
			}
		}
		q, j := pr.jumpStraightNoDiags(p.Shift(dir.X, 0), gruid.Point{dir.X, 0}, to)
		//_, j := pr.jumpStraightNoDiags(p.Shift(dir.X, 0), gruid.Point{dir.X, 0}, to)
		if j > 0 {
			//return p, i
			pr.addSuccessor(q, from, to, cost+i+j)
		}
		q, j = pr.jumpStraightNoDiags(p.Shift(0, dir.Y), gruid.Point{0, dir.Y}, to)
		//_, j = pr.jumpStraightNoDiags(p.Shift(0, dir.Y), gruid.Point{0, dir.Y}, to)
		if j > 0 {
			pr.addSuccessor(q, from, to, cost+i+j)
		}
		p = p.Add(dir)
		i += 2
	}
}

// jump makes a jump from a position in a given direction in order to find an
// appropriate successor, skipping nodes that do not require being added to the
// open list.
func (pr *PathRange) jump(p, dir, to gruid.Point, cost int) (gruid.Point, int) {
	switch {
	case dir.X == 0 || dir.Y == 0:
		return pr.jumpStraight(p, dir, to)
	default:
		return pr.jumpDiagonal(p, dir, to, cost)
	}
}

// jumpNoDiags is the same as jump, except that it uses the same concept but
// for paths that cannot be diagonal: in practice, this means that diagonal
// jumps and diagonal forced neighbors are only processed if they are doable
// using two cardinal movements.
func (pr *PathRange) jumpNoDiags(p, dir, to gruid.Point, cost int) (gruid.Point, int) {
	switch {
	case dir.X == 0 || dir.Y == 0:
		return pr.jumpStraightNoDiags(p, dir, to)
	default:
		return pr.jumpDiagonalNoDiags(p, dir, to, cost)
	}
}

// dirnorm returns a normalized direction between two points, so that
// directions that aren't cardinal nor diagonal are transformed into the
// cardinal part (this corresponds to pruned intermediate nodes in diagonal
// jump).
func dirnorm(p, q gruid.Point) gruid.Point {
	dir := q.Sub(p)
	dx := abs(dir.X)
	dy := abs(dir.Y)
	dir = gruid.Point{sign(dir.X), sign(dir.Y)}
	switch {
	case dx == dy:
	case dx > dy:
		dir.Y = 0
	default:
		dir.X = 0
	}
	return dir
}

// neighbors returns the natural neigbors of current node, and adds to the
// queue forced neighbors.
//
// NOTE: there's a bit of redundant work here, because forced neighbor
// positions are already computed during the jump. It should not matter in
// practice, because in most situations JPS adds very few nodes to the open
// list, so this function is not called very often, and so is not a bottleneck.
func (pr *PathRange) neighbors(neighbors []gruid.Point, n *node, to gruid.Point) []gruid.Point {
	dir := dirnorm(n.Parent, n.P)
	switch {
	case dir.X == 0 || dir.Y == 0:
		neighbors = append(neighbors, n.P.Add(dir))
		if q := left(n.P, dir); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(dir)) {
				p := q.Add(dir)
				cost := diagCost(pr.diags)
				pr.addSuccessor(p, n.P, to, n.Cost+cost)
			}
		}
		if q := right(n.P, dir); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(dir)) {
				p := q.Add(dir)
				cost := diagCost(pr.diags)
				pr.addSuccessor(p, n.P, to, n.Cost+cost)
			}
		}
	default:
		diag := false
		q0 := n.P.Shift(dir.X, 0)
		q1 := n.P.Shift(0, dir.Y)
		if !pr.diags {
			diag = pr.pass(q0) || pr.pass(q1)
		}
		neighbors = append(neighbors, q0)
		neighbors = append(neighbors, q1)
		if pr.diags || diag {
			neighbors = append(neighbors, n.P.Add(dir))
		}
		if q := n.P.Shift(-dir.X, 0); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(gruid.Point{0, dir.Y})) {
				cost := diagCost(pr.diags)
				pr.addSuccessor(q.Shift(0, dir.Y), n.P, to, n.Cost+cost)
			}
		}
		if q := n.P.Shift(0, -dir.Y); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(gruid.Point{dir.X, 0})) {
				cost := diagCost(pr.diags)
				pr.addSuccessor(q.Shift(dir.X, 0), n.P, to, n.Cost+cost)
			}
		}
	}
	return neighbors
}

func diagCost(diags bool) int {
	if diags {
		return 1
	}
	return 2
}

func (pr *PathRange) addSuccessor(p, parent, to gruid.Point, cost int) {
	if !pr.pass(p) {
		return
	}
	nbNode := pr.AstarNodes.get(pr, p)
	if cost < nbNode.Cost {
		if nbNode.Open {
			pqRemove(&pr.AstarQueue, nbNode.Idx)
		}
		nbNode.Open = false
		nbNode.Closed = false
	}
	if !nbNode.Open && !nbNode.Closed {
		nbNode.Cost = cost
		nbNode.Open = true
		delta := p.Sub(to)
		dx := abs(delta.X)
		dy := abs(delta.Y)
		nbNode.Estimation = dx + dy
		nbNode.Rank = cost + pr.estim(dx, dy)
		nbNode.Parent = parent
		pqPush(&pr.AstarQueue, nbNode)
	}
}

// jumpPath adds to the path the points from p (included) to q (excluded)
// corresponding to a jump (possibly diagonal+straight) from q to p.
func (pr *PathRange) jumpPath(path []gruid.Point, p, q gruid.Point) []gruid.Point {
	dir := q.Sub(p)
	dx := abs(dir.X)
	dy := abs(dir.Y)
	dir = gruid.Point{sign(dir.X), sign(dir.Y)}
	switch {
	case dx > dy:
		for i := 0; i < dx-dy; i++ {
			path = append(path, p)
			p = p.Add(gruid.Point{dir.X, 0})
		}
	case dx < dy:
		for i := 0; i < dy-dx; i++ {
			path = append(path, p)
			p = p.Add(gruid.Point{0, dir.Y})
		}
	}
	for ; p != q; p = p.Add(dir) {
		path = append(path, p)
		if !pr.diags {
			if dir.X != 0 && dir.Y != 0 {
				if px := p.Add(gruid.Point{dir.X, 0}); pr.pass(px) {
					path = append(path, px)
				} else if py := p.Add(gruid.Point{0, dir.Y}); pr.pass(py) {
					path = append(path, py)
				}
			}
		}
	}
	return path
}

func (pr *PathRange) path(path []gruid.Point, from gruid.Point, n *node) []gruid.Point {
	for {
		if n.P == from {
			path = append(path, n.P)
			break
		}
		path = pr.jumpPath(path, n.P, n.Parent)
		n = pr.AstarNodes.at(pr, n.Parent)
	}
	for i := range path[:len(path)/2] {
		path[i], path[len(path)-i-1] = path[len(path)-i-1], path[i]
	}
	return path
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

func (pr *PathRange) estim(x, y int) int {
	if pr.diags {
		return max(x, y)
	}
	return x + y
}

func sign(n int) int {
	var i int
	switch {
	case n > 0:
		i = 1
	case n < 0:
		i = -1
	}
	return i
}

//var logrid gruid.Grid

//func init() {
//logrid = gruid.NewGrid(80, 24)
//}

//func logPath(path []gruid.Point) {
//for _, p := range path {
//c := logrid.At(p)
//if c.Rune == '.' {
//logrid.Set(p, gruid.Cell{Rune: 'o'})
//}
//}
//}
