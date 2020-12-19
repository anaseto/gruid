package paths

import "github.com/anaseto/gruid"

type ccNode struct {
	idx int // map number (for caching)
	id  int // component identifier
}

// ComputeCCAll computes a map of the connected components. It makes the
// assumption that the paths are bidirectional, allowing for efficient
// computation. It uses the same caching structures as ComputeCC.
func (pr *PathRange) ComputeCCAll(nb Pather) {
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
			p := idxToPos(idx, w)
			for _, q := range nb.Neighbors(p) {
				if !q.In(pr.rg) {
					continue
				}
				nidx := pr.idx(q)
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
func (pr *PathRange) ComputeCC(nb Pather, p gruid.Point) {
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
	idx := pr.idx(p)
	pr.cc[idx].id = ccid
	pr.cc[idx].idx = pr.ccidx
	pr.ccstack = append(pr.ccstack, idx)
	for len(pr.ccstack) > 0 {
		idx = pr.ccstack[len(pr.ccstack)-1]
		pr.ccstack = pr.ccstack[:len(pr.ccstack)-1]
		pr.ccIterCache = append(pr.ccIterCache, idx)
		p := idxToPos(idx, w)
		for _, q := range nb.Neighbors(p) {
			if !q.In(pr.rg) {
				continue
			}
			nidx := pr.idx(q)
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
func (pr *PathRange) CCAt(p gruid.Point) int {
	if !p.In(pr.rg) || pr.cc == nil {
		return -1
	}
	node := pr.cc[pr.idx(p)]
	if node.idx != pr.ccidx {
		return -1
	}
	return node.id
}

// CCIter iterates a function on all the positions belonging to the connected
// component computed by the last ComputeCC call.  Caching is used for
// efficiency, so the iteration function should avoid calling CCIter and
// ComputeCC.
func (pr *PathRange) CCIter(fn func(gruid.Point)) {
	if pr.ccIterCache == nil {
		return
	}
	w, _ := pr.rg.Size()
	for _, idx := range pr.ccIterCache {
		p := idxToPos(idx, w)
		fn(p)
	}
}
