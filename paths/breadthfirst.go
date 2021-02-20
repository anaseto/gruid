package paths

import (
	"github.com/anaseto/gruid"
)

// BreadthFirstMapAt returns the cost associated to a position in the last
// computed breadth first map. It returns the last maxCost + 1 if the position
// is out of range, the same as in-range unreachable positions.
// BreadthFirstMapAt uses a cached breadth first map, that will be invalidated
// in case a new one is computed using the same PathFinder.
func (pr *PathRange) BreadthFirstMapAt(p gruid.Point) int {
	if !p.In(pr.Rg) || pr.BfMap == nil {
		return pr.BfUnreachable
	}
	cost := pr.BfMap[pr.idx(p)]
	if cost == 0 {
		return pr.BfUnreachable
	}
	return cost - 1
}

// BreadthFirstMap efficiently computes a map of minimal distance costs from
// source positions to all the positions in the PathFinder range up to a
// maximal cost. Other positions will have the value maxCost+1, including
// unreachable ones. It returns a cached slice of map nodes in increasing cost
// order.
//
// It can be viewed as a particular case of DijkstraMap built with a cost
// function that returns 1 for all neighbors, but it is more efficient.
func (pr *PathRange) BreadthFirstMap(nb Pather, sources []gruid.Point, maxCost int) []Node {
	max := pr.Rg.Size()
	w, h := max.X, max.Y
	if pr.BfMap == nil {
		pr.BfMap = make([]int, w*h)
		pr.BfQueue = make([]Node, w*h)
	} else {
		for i := range pr.BfMap {
			pr.BfMap[i] = 0
		}
	}
	var qstart, qend int
	pr.BfUnreachable = maxCost + 1
	for _, p := range sources {
		if !p.In(pr.Rg) {
			continue
		}
		pr.BfMap[pr.idx(p)] = 1
		pr.BfQueue[qend] = Node{P: p, Cost: 0}
		qend++
	}
	for qstart < qend {
		n := pr.BfQueue[qstart]
		qstart++
		if n.Cost >= maxCost {
			continue
		}
		cidx := pr.idx(n.P)
		cost := pr.BfMap[cidx]
		for _, q := range nb.Neighbors(n.P) {
			if !q.In(pr.Rg) {
				continue
			}
			nidx := pr.idx(q)
			if pr.BfMap[nidx] == 0 {
				pr.BfMap[nidx] = cost + 1
				pr.BfQueue[qend] = Node{P: q, Cost: cost}
				qend++
			}
		}
	}
	return pr.BfQueue[0:qend]
}
