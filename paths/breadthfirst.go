package paths

import "github.com/anaseto/gruid"

// CostAt returns the cost associated to a position in the last computed
// breadth first map. It returns the last maxCost + 1 if the position is out of
// range, the same as in-range unreachable positions.  CostAt uses a cached
// breadth first map, that will be invalidated in case a new one is computed
// using the same PathFinder.
func (pf *PathRange) CostAt(pos gruid.Position) int {
	if !pos.In(pf.rg) || pf.bfmap == nil {
		return pf.bfunreachable
	}
	node := pf.bfmap[pf.idx(pos)]
	if node.idx != pf.bfidx {
		return pf.bfunreachable
	}
	return node.cost
}

type bfNode struct {
	idx  int // map number (for caching)
	cost int // path cost from source
}

// BreadthFirstMap efficiently computes a map of minimal distance costs from
// source positions to all the positions in the PathFinder range up to a
// maximal cost. Other positions will have the value maxCost+1, including
// unreachable ones. It can be viewed as a particular case of DijkstraMap built
// with a cost function that returns 1 for all neighbors, but it is more
// efficient.
func (pf *PathRange) BreadthFirstMap(nb Neighborer, sources []gruid.Position, maxCost int) {
	w, h := pf.rg.Size()
	if pf.bfmap == nil {
		pf.bfmap = make([]bfNode, w*h)
		pf.bfqueue = make([]int, w*h)
	} else {
		pf.bfidx++
	}
	var qstart, qend int
	pf.bfunreachable = maxCost + 1
	for _, pos := range sources {
		if !pos.In(pf.rg) {
			continue
		}
		idx := pf.idx(pos)
		pf.bfmap[idx].cost = 0
		pf.bfmap[idx].idx = pf.bfidx
		pf.bfqueue[qend] = idx
		qend++
	}
	for qstart < qend {
		cidx := pf.bfqueue[qstart]
		qstart++
		if pf.bfmap[cidx].cost >= maxCost {
			continue
		}
		cpos := idxToPos(cidx, w)
		for _, npos := range nb.Neighbors(cpos) {
			nidx := pf.idx(npos)
			if pf.bfmap[nidx].idx != pf.bfidx {
				pf.bfqueue[qend] = nidx
				qend++
				pf.bfmap[nidx].cost = 1 + pf.bfmap[cidx].cost
				pf.bfmap[nidx].idx = pf.bfidx
			}
		}
	}
}
