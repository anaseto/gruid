package paths

import "github.com/anaseto/gruid"

type ccNode struct {
	idx int // map number (for caching)
	id  int // component identifier
}

// ComputeCCAll computes a map of the connected components. It makes the
// assumption that the paths are bidirectional, allowing for efficient
// computation. It uses the same caching structures as ComputeCC.
func (pf *PathRange) ComputeCCAll(nb Neighborer) {
	w, h := pf.rg.Size()
	if pf.cc == nil {
		pf.cc = make([]ccNode, w*h)
	}
	pf.ccidx++
	pf.neighbors = nb
	pf.ccstack = pf.ccstack[:0]
	ccid := 0
	for i := 0; i < len(pf.cc); i++ {
		if pf.cc[i].idx == pf.ccidx {
			continue
		}
		pf.cc[i].id = ccid
		pf.cc[i].idx = pf.ccidx
		pf.ccstack = append(pf.ccstack, i)
		for len(pf.ccstack) > 0 {
			idx := pf.ccstack[len(pf.ccstack)-1]
			pf.ccstack = pf.ccstack[:len(pf.ccstack)-1]
			pos := idxToPos(idx, w)
			for _, npos := range nb.Neighbors(pos) {
				if !npos.In(pf.rg) {
					continue
				}
				nidx := pf.idx(npos)
				if pf.cc[nidx].idx == pf.ccidx {
					continue
				}
				pf.cc[nidx].id = ccid
				pf.cc[nidx].idx = pf.ccidx
				pf.ccstack = append(pf.ccstack, nidx)
			}
		}
		ccid++
	}
}

// ComputeCC computes the connected component which contains a given position.
// It makes the assumption that the paths are bidirectional, allowing for
// efficient computation.
func (pf *PathRange) ComputeCC(nb Neighborer, pos gruid.Position) {
	w, h := pf.rg.Size()
	if pf.cc == nil {
		pf.cc = make([]ccNode, w*h)
	}
	pf.ccidx++
	if pf.ccIterCache == nil {
		pf.ccIterCache = make([]int, w*h)
	}
	pf.ccIterCache = pf.ccIterCache[:0]
	pf.neighbors = nb
	pf.ccstack = pf.ccstack[:0]
	ccid := 0
	idx := pf.idx(pos)
	pf.cc[idx].id = ccid
	pf.cc[idx].idx = pf.ccidx
	pf.ccstack = append(pf.ccstack, idx)
	for len(pf.ccstack) > 0 {
		idx = pf.ccstack[len(pf.ccstack)-1]
		pf.ccstack = pf.ccstack[:len(pf.ccstack)-1]
		pf.ccIterCache = append(pf.ccIterCache, idx)
		pos := idxToPos(idx, w)
		for _, npos := range nb.Neighbors(pos) {
			if !npos.In(pf.rg) {
				continue
			}
			nidx := pf.idx(npos)
			if pf.cc[nidx].idx == pf.ccidx {
				continue
			}
			pf.cc[nidx].id = ccid
			pf.cc[nidx].idx = pf.ccidx
			pf.ccstack = append(pf.ccstack, nidx)
		}
	}
}

// CCAt returns a positive number identifying the position's connected
// component as computed by either the last ComputeCC or ComputeCCAll call. It
// returns -1 on out of range positions.
func (pf *PathRange) CCAt(pos gruid.Position) int {
	if !pos.In(pf.rg) || pf.cc == nil {
		return -1
	}
	node := pf.cc[pf.idx(pos)]
	if node.idx != pf.ccidx {
		return -1
	}
	return node.id
}

// CCIter iterates a function on all the positions belonging to the connected
// component computed by the last ComputeCC call.  Caching is used for
// efficiency, so the iteration function should avoid calling CCIter and
// ComputeCC.
func (pf *PathRange) CCIter(fn func(gruid.Position)) {
	if pf.ccIterCache == nil {
		return
	}
	w, _ := pf.rg.Size()
	for _, idx := range pf.ccIterCache {
		pos := idxToPos(idx, w)
		fn(pos)
	}
}
