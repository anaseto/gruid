package paths

import "github.com/anaseto/gruid"

type ccNode struct {
	idx int // map number (for caching)
	id  int // component identifier
}

// ComputeCCAll computes a map of the connected components. It makes the
// assumption that the paths are bidirectional, allowing for efficient
// computation. It uses the same caching structures as ComputeCC.
func (pr *PathRange) ComputeCCAll(nb Neighborer) {
	w, h := pr.rg.Size()
	if pr.cc == nil {
		pr.cc = make([]ccNode, w*h)
	}
	pr.ccidx++
	pr.neighbors = nb
	pr.ccstack = pr.ccstack[:0]
	ccid := 0
	for i := 0; i < len(pr.cc); i++ {
		if pr.cc[i].idx == pr.ccidx {
			continue
		}
		pr.cc[i].id = ccid
		pr.cc[i].idx = pr.ccidx
		pr.ccstack = append(pr.ccstack, i)
		for len(pr.ccstack) > 0 {
			idx := pr.ccstack[len(pr.ccstack)-1]
			pr.ccstack = pr.ccstack[:len(pr.ccstack)-1]
			pos := idxToPos(idx, w)
			for _, npos := range nb.Neighbors(pos) {
				if !npos.In(pr.rg) {
					continue
				}
				nidx := pr.idx(npos)
				if pr.cc[nidx].idx == pr.ccidx {
					continue
				}
				pr.cc[nidx].id = ccid
				pr.cc[nidx].idx = pr.ccidx
				pr.ccstack = append(pr.ccstack, nidx)
			}
		}
		ccid++
	}
}

// ComputeCC computes the connected component which contains a given position.
// It makes the assumption that the paths are bidirectional, allowing for
// efficient computation.
func (pr *PathRange) ComputeCC(nb Neighborer, pos gruid.Position) {
	w, h := pr.rg.Size()
	if pr.cc == nil {
		pr.cc = make([]ccNode, w*h)
	}
	pr.ccidx++
	if pr.ccIterCache == nil {
		pr.ccIterCache = make([]int, w*h)
	}
	pr.ccIterCache = pr.ccIterCache[:0]
	pr.neighbors = nb
	pr.ccstack = pr.ccstack[:0]
	ccid := 0
	idx := pr.idx(pos)
	pr.cc[idx].id = ccid
	pr.cc[idx].idx = pr.ccidx
	pr.ccstack = append(pr.ccstack, idx)
	for len(pr.ccstack) > 0 {
		idx = pr.ccstack[len(pr.ccstack)-1]
		pr.ccstack = pr.ccstack[:len(pr.ccstack)-1]
		pr.ccIterCache = append(pr.ccIterCache, idx)
		pos := idxToPos(idx, w)
		for _, npos := range nb.Neighbors(pos) {
			if !npos.In(pr.rg) {
				continue
			}
			nidx := pr.idx(npos)
			if pr.cc[nidx].idx == pr.ccidx {
				continue
			}
			pr.cc[nidx].id = ccid
			pr.cc[nidx].idx = pr.ccidx
			pr.ccstack = append(pr.ccstack, nidx)
		}
	}
}

// CCAt returns a positive number identifying the position's connected
// component as computed by either the last ComputeCC or ComputeCCAll call. It
// returns -1 on out of range positions.
func (pr *PathRange) CCAt(pos gruid.Position) int {
	if !pos.In(pr.rg) || pr.cc == nil {
		return -1
	}
	node := pr.cc[pr.idx(pos)]
	if node.idx != pr.ccidx {
		return -1
	}
	return node.id
}

// CCIter iterates a function on all the positions belonging to the connected
// component computed by the last ComputeCC call.  Caching is used for
// efficiency, so the iteration function should avoid calling CCIter and
// ComputeCC.
func (pr *PathRange) CCIter(fn func(gruid.Position)) {
	if pr.ccIterCache == nil {
		return
	}
	w, _ := pr.rg.Size()
	for _, idx := range pr.ccIterCache {
		pos := idxToPos(idx, w)
		fn(pos)
	}
}
