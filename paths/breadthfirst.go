package paths

import "github.com/anaseto/gruid"

// CostAt returns the cost associated to a position in the last computed
// breadth first map. It returns the last maxCost + 1 if the position is out of
// range, the same as in-range unreachable positions.  CostAt uses a cached
// breadth first map, that will be invalidated in case a new one is computed
// using the same PathFinder.
func (pr *PathRange) CostAt(pos gruid.Position) int {
	if !pos.In(pr.rg) || pr.bfmap == nil {
		return pr.bfunreachable
	}
	node := pr.bfmap[pr.idx(pos)]
	if node.idx != pr.bfidx {
		return pr.bfunreachable
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
func (pr *PathRange) BreadthFirstMap(nb Neighborer, sources []gruid.Position, maxCost int) {
	w, h := pr.rg.Size()
	if pr.bfmap == nil {
		pr.bfmap = make([]bfNode, w*h)
		pr.bfqueue = make([]int, w*h)
	}
	pr.bfidx++
	var qstart, qend int
	pr.bfunreachable = maxCost + 1
	for _, pos := range sources {
		if !pos.In(pr.rg) {
			continue
		}
		idx := pr.idx(pos)
		pr.bfmap[idx].cost = 0
		pr.bfmap[idx].idx = pr.bfidx
		pr.bfqueue[qend] = idx
		qend++
	}
	for qstart < qend {
		cidx := pr.bfqueue[qstart]
		qstart++
		if pr.bfmap[cidx].cost >= maxCost {
			continue
		}
		cpos := idxToPos(cidx, w)
		for _, npos := range nb.Neighbors(cpos) {
			if !npos.In(pr.rg) {
				continue
			}
			nidx := pr.idx(npos)
			if pr.bfmap[nidx].idx != pr.bfidx {
				pr.bfqueue[qend] = nidx
				qend++
				pr.bfmap[nidx].cost = 1 + pr.bfmap[cidx].cost
				pr.bfmap[nidx].idx = pr.bfidx
			}
		}
	}
}
