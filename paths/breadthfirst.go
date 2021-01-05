package paths

import (
	"math"

	"github.com/anaseto/gruid"
)

// CostAt returns the cost associated to a position in the last computed
// breadth first map. It returns the last maxCost + 1 if the position is out of
// range, the same as in-range unreachable positions.  CostAt uses a cached
// breadth first map, that will be invalidated in case a new one is computed
// using the same PathFinder.
func (pr *PathRange) CostAt(p gruid.Point) int {
	if !p.In(pr.Rg) || pr.Bfmap == nil {
		return pr.Bfunreachable
	}
	node := pr.Bfmap[pr.idx(p)]
	if node.Idx != pr.Bfidx {
		return pr.Bfunreachable
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
// unreachable ones. It can be viewed as a particular case of DijkstraMap built
// with a cost function that returns 1 for all neighbors, but it is more
// efficient.
func (pr *PathRange) BreadthFirstMap(nb Pather, sources []gruid.Point, maxCost int) {
	max := pr.Rg.Size()
	w, h := max.X, max.Y
	if pr.Bfmap == nil {
		pr.Bfmap = make([]bfNode, w*h)
		pr.Bfqueue = make([]int, w*h)
	}
	pr.Bfidx++
	var qstart, qend int
	pr.Bfunreachable = maxCost + 1
	for _, p := range sources {
		if !p.In(pr.Rg) {
			continue
		}
		idx := pr.idx(p)
		pr.Bfmap[idx].Cost = 0
		pr.Bfmap[idx].Idx = pr.Bfidx
		pr.Bfqueue[qend] = idx
		qend++
	}
	for qstart < qend {
		cidx := pr.Bfqueue[qstart]
		qstart++
		if pr.Bfmap[cidx].Cost >= maxCost {
			continue
		}
		cpos := idxToPos(cidx, w)
		for _, q := range nb.Neighbors(cpos) {
			if !q.In(pr.Rg) {
				continue
			}
			nidx := pr.idx(q)
			if pr.Bfmap[nidx].Idx != pr.Bfidx {
				pr.Bfqueue[qend] = nidx
				qend++
				pr.Bfmap[nidx].Cost = 1 + pr.Bfmap[cidx].Cost
				pr.Bfmap[nidx].Idx = pr.Bfidx
			}
		}
	}
	pr.checkBfIdx()
}

func (pr *PathRange) checkBfIdx() {
	if pr.Bfidx < math.MaxInt32 {
		return
	}
	for i, n := range pr.Bfmap {
		idx := 0
		if n.Idx == pr.Bfidx {
			idx = 1
		}
		pr.Bfmap[i] = bfNode{Cost: n.Cost, Idx: idx}
	}
	pr.Bfidx = 1
}
