// This file implements a JPS (jump-point-search) pathfinding algorithm. For
// more information: https://en.wikipedia.org/wiki/Jump_point_search

package paths

import (
	"github.com/anaseto/gruid"
	//"log"
)

// JPSPath returns a path from a position to another, including thoses
// positions, in the path order. It uses the given path slice to avoid
// allocations unless its capacity is not enough. The passable function
// controls which positions can be passed. If diags is false, only movements in
// straight cardinal directions are allowed.
//
// The function returns nil if no path was found.
//
// In most situations, JPSPath has significantly better performance than
// AstarPath. The algorithm limitation is that it only handles uniform costs
// and natural neighbors in grid geometry.
func (pr *PathRange) JPSPath(path []gruid.Point, from, to gruid.Point, passable func(gruid.Point) bool, diags bool) []gruid.Point {
	if !from.In(pr.Rg) || !to.In(pr.Rg) {
		return nil
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
	var neighbors []gruid.Point
	neighbors = pr.expandOrigin(neighbors, from, to)
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
			path = pr.path(path, n)
			return path
		}

		neighbors = pr.neighbors(neighbors[:0], n, to)
		for _, p := range neighbors {
			dir := p.Sub(n.P)
			var q gruid.Point
			var i int
			if pr.diags {
				q, i = pr.jump(p, dir, to)
			} else {
				q, i = pr.jumpNoDiags(p, dir, to)
			}
			if i > 0 {
				pr.addSuccessor(q, dir, to, n.Cost+i)
			}
		}
	}
}

func (pr *PathRange) expandOrigin(neighbors []gruid.Point, from, to gruid.Point) []gruid.Point {
	neighbors = append(neighbors,
		from.Add(gruid.Point{-1, 0}),
		from.Add(gruid.Point{1, 0}),
		from.Add(gruid.Point{0, 1}),
		from.Add(gruid.Point{0, -1}),
		from.Add(gruid.Point{-1, -1}),
		from.Add(gruid.Point{1, -1}),
		from.Add(gruid.Point{-1, 1}),
		from.Add(gruid.Point{1, 1}),
	)
	for _, q := range neighbors {
		dir := q.Sub(from)
		if !pr.diags {
			if dir.X != 0 && dir.Y != 0 {
				if pr.pass(from.Add(gruid.Point{dir.X, 0})) || pr.pass(from.Add(gruid.Point{0, dir.Y})) {
					pr.addSuccessor(q, dir, to, 2)
				}
				continue
			}
			pr.addSuccessor(q, dir, to, 1)
			continue
		}
		pr.addSuccessor(q, dir, to, 1)
	}
	return neighbors
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
			if p == to {
				return p, i
			}
			if !pr.passable(p) {
				return p, 0
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
		if p == to {
			return p, i
		}
		if !pr.passable(p) {
			return p, 0
		}
		p = p.Add(dir)
	}
	return p, 0
}

func (pr *PathRange) jumpStraightLeft(p, dir, to gruid.Point, max int) (gruid.Point, int) {
	for i := 1; i < max+1; i++ {
		if p == to {
			return p, i
		}
		if !pr.passable(p) {
			return p, 0
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
		if p == to {
			return p, i
		}
		if !pr.passable(p) {
			return p, 0
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
			if p == to {
				return p, i
			}
			if !pr.passable(p) {
				return p, 0
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
		if p == to {
			return p, i
		}
		if !pr.passable(p) {
			return p, 0
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
		if p == to {
			return p, i
		}
		if !pr.passable(p) {
			return p, 0
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

func (pr *PathRange) jumpDiagonal(p, dir, to gruid.Point) (gruid.Point, int) {
	i := 1
	for {
		if p == to {
			return p, i
		}
		if !pr.pass(p) {
			return p, 0
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
		_, j := pr.jumpStraight(p.Shift(dir.X, 0), gruid.Point{dir.X, 0}, to)
		if j > 0 {
			return p, i
		}
		_, j = pr.jumpStraight(p.Shift(0, dir.Y), gruid.Point{0, dir.Y}, to)
		if j > 0 {
			return p, i
		}
		p = p.Add(dir)
		i++
	}
}

func (pr *PathRange) jumpDiagonalNoDiags(p, dir, to gruid.Point) (gruid.Point, int) {
	i := 2 // diagonals cost 2 (two cordinal movements)
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
		_, j := pr.jumpStraightNoDiags(p.Shift(dir.X, 0), gruid.Point{dir.X, 0}, to)
		if j > 0 {
			return p, i
		}
		_, j = pr.jumpStraightNoDiags(p.Shift(0, dir.Y), gruid.Point{0, dir.Y}, to)
		if j > 0 {
			return p, i
		}
		p = p.Add(dir)
		i += 2
	}
}

// jump makes a jump from a position in a given direction in order to find an
// appropiate successor, skipping nodes that do not require being added to the
// open list.
func (pr *PathRange) jump(p, dir, to gruid.Point) (gruid.Point, int) {
	switch dir {
	case gruid.Point{-1, 0},
		gruid.Point{1, 0},
		gruid.Point{0, 1},
		gruid.Point{0, -1}:
		return pr.jumpStraight(p, dir, to)
	default:
		return pr.jumpDiagonal(p, dir, to)
	}
}

// jumpNoDiags is the same as jump, except that it uses the same concept but
// for paths that cannot be diagonal: in practice, this means that diagonal
// jumps and diagonal forced neighbors are only processed if they are doable
// using two cardinal movements.
func (pr *PathRange) jumpNoDiags(p, dir, to gruid.Point) (gruid.Point, int) {
	switch dir {
	case gruid.Point{-1, 0},
		gruid.Point{1, 0},
		gruid.Point{0, 1},
		gruid.Point{0, -1}:
		return pr.jumpStraightNoDiags(p, dir, to)
	default:
		return pr.jumpDiagonalNoDiags(p, dir, to)
	}
}

// neighbors returns the natural neigbors of current node, and adds to the
// queue forced neighbors.
//
// NOTE: there's a bit of redundant work here, because forced neighbor
// positions are already computed during the jump. It should not matter in
// practice, because in most situations JPS adds very few nodes to the open
// list, so this function is not called very often, and so is not a bottleneck.
func (pr *PathRange) neighbors(neighbors []gruid.Point, n *node, to gruid.Point) []gruid.Point {
	switch n.Dir {
	case gruid.Point{-1, 0},
		gruid.Point{1, 0},
		gruid.Point{0, 1},
		gruid.Point{0, -1}:
		neighbors = append(neighbors, n.P.Add(n.Dir))
		if q := left(n.P, n.Dir); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(n.Dir)) {
				p := q.Add(n.Dir)
				cost := diagCost(pr.diags)
				pr.addSuccessor(p, p.Sub(n.P), to, n.Cost+cost)
			}
		}
		if q := right(n.P, n.Dir); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(n.Dir)) {
				p := q.Add(n.Dir)
				cost := diagCost(pr.diags)
				pr.addSuccessor(p, p.Sub(n.P), to, n.Cost+cost)
			}
		}
	case gruid.Point{-1, -1},
		gruid.Point{1, -1},
		gruid.Point{-1, 1},
		gruid.Point{1, 1}:
		diag := false
		q0 := n.P.Shift(n.Dir.X, 0)
		q1 := n.P.Shift(0, n.Dir.Y)
		if !pr.diags {
			diag = pr.pass(q0) || pr.pass(q1)
		}
		neighbors = append(neighbors, q0)
		neighbors = append(neighbors, q1)
		if pr.diags || diag {
			neighbors = append(neighbors, n.P.Add(n.Dir))
		}
		if q := n.P.Shift(-n.Dir.X, 0); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(gruid.Point{0, n.Dir.Y})) {
				cost := diagCost(pr.diags)
				pr.addSuccessor(q.Shift(0, n.Dir.Y), gruid.Point{-n.Dir.X, n.Dir.Y}, to, n.Cost+cost)
			}
		}
		if q := n.P.Shift(0, -n.Dir.Y); !pr.pass(q) {
			if pr.diags || pr.pass(n.P.Add(gruid.Point{n.Dir.X, 0})) {
				cost := diagCost(pr.diags)
				pr.addSuccessor(q.Shift(n.Dir.X, 0), gruid.Point{n.Dir.X, -n.Dir.Y}, to, n.Cost+cost)
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

func (pr *PathRange) addSuccessor(p, dir, to gruid.Point, cost int) {
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
		nbNode.Dir = dir
		pqPush(&pr.AstarQueue, nbNode)
	}
}

func (pr *PathRange) path(path []gruid.Point, n *node) []gruid.Point {
	p := n.P
	dir := n.Dir
	count := 0
loop:
	for {
		count++
		if count > 1000 {
			break
		}
		path = append(path, p)
		switch dir {
		case gruid.Point{0, 0}:
			break loop
		}
		if !pr.diags {
			switch dir {
			case gruid.Point{-1, -1},
				gruid.Point{1, -1},
				gruid.Point{-1, 1},
				gruid.Point{1, 1}:
				if px := p.Sub(gruid.Point{dir.X, 0}); pr.pass(px) {
					path = append(path, px)
				} else if py := p.Sub(gruid.Point{0, dir.Y}); pr.pass(py) {
					path = append(path, py)
				}
			}
		}
		p = p.Sub(dir)
		n = pr.AstarNodes.at(pr, p)
		if n != nil {
			dir = n.Dir
		}
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
