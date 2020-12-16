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
	return pf.bfmap[pf.idx(pos)]
}

// BreadthFirstMap efficiently computes a map of minimal distance costs from
// source positions to all the positions in the PathFinder range up to a
// maximal cost. Other positions will have the value maxCost+1, including
// unreachable ones. It can be viewed as a particular case of DijkstraMap built
// with a cost function that returns 1 for all neighbors, but it is more
// efficient.
func (pf *PathRange) BreadthFirstMap(nb Neighborer, sources []gruid.Position, maxCost int) {
	if pf.bfvisited == nil {
		w, h := pf.rg.Size()
		pf.bfvisited = make([]bool, w*h)
		pf.bfqueue = make([]int, w*h)
		pf.bfmap = make([]int, w*h)
	}
	bfmap := pf.bfmap[:]
	var qstart, qend int
	w, h := pf.rg.Size()
	for i := 0; i < w*h; i++ {
		bfmap[i] = maxCost + 1
	}
	pf.bfunreachable = maxCost + 1
	for _, pos := range sources {
		if !pos.In(pf.rg) {
			continue
		}
		s := pf.idx(pos)
		bfmap[s] = 0
		pf.bfqueue[qend] = s
		qend++
		pf.bfvisited[s] = true
	}
	for qstart < qend {
		cidx := pf.bfqueue[qstart]
		qstart++
		if bfmap[cidx] == maxCost {
			continue
		}
		cpos := idxToPos(cidx, w)
		for _, npos := range nb.Neighbors(cpos) {
			nidx := pf.idx(npos)
			if !pf.bfvisited[nidx] {
				pf.bfqueue[qend] = nidx
				qend++
				pf.bfvisited[nidx] = true
				bfmap[nidx] = 1 + bfmap[cidx]
			}
		}
	}
}
