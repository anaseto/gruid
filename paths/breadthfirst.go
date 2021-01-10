package paths

import (
	"math"

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
	node := pr.BfMap[pr.idx(p)]
	if node.Idx != pr.BfIdx {
		return pr.BfUnreachable
	}
	return node.Cost
}

type bfNode struct {
	Idx  int // map number (for caching)
	Cost int // path cost from source
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
		pr.BfMap = make([]bfNode, w*h)
		pr.BfQueue = make([]Node, w*h)
	}
	pr.BfIdx++
	var qstart, qend int
	pr.BfUnreachable = maxCost + 1
	for _, p := range sources {
		if !p.In(pr.Rg) {
			continue
		}
		pr.BfMap[pr.idx(p)] = bfNode{Idx: pr.BfIdx, Cost: 0}
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
		for _, q := range nb.Neighbors(n.P) {
			if !q.In(pr.Rg) {
				continue
			}
			nidx := pr.idx(q)
			if pr.BfMap[nidx].Idx != pr.BfIdx {
				c := 1 + pr.BfMap[cidx].Cost
				pr.BfMap[nidx] = bfNode{Idx: pr.BfIdx, Cost: c}
				pr.BfQueue[qend] = Node{P: q, Cost: c}
				qend++
			}
		}
	}
	pr.checkBfIdx()
	return pr.BfQueue[0:qend]
}

func (pr *PathRange) checkBfIdx() {
	if pr.BfIdx < math.MaxInt32 {
		return
	}
	for i, n := range pr.BfMap {
		idx := 0
		if n.Idx == pr.BfIdx {
			idx = 1
		}
		pr.BfMap[i] = bfNode{Cost: n.Cost, Idx: idx}
	}
	pr.BfIdx = 1
}
