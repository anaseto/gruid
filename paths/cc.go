package paths

import (
	"math"

	"github.com/anaseto/gruid"
)

type ccNode struct {
	Idx int // map number (for caching)
	ID  int // component identifier
}

// ComputeCCAll computes a map of the connected components. It makes the
// assumption that the paths are bidirectional, allowing for efficient
// computation. It uses the same caching structures as ComputeCC.
func (pr *PathRange) ComputeCCAll(nb Pather) {
	max := pr.Rg.Size()
	w, h := max.X, max.Y
	if pr.CC == nil {
		pr.CC = make([]ccNode, w*h)
	}
	pr.CCidx++
	defer pr.checkCCIdx()
	pr.CCstack = pr.CCstack[:0]
	ccid := 0
	for i := 0; i < len(pr.CC); i++ {
		if pr.CC[i].Idx == pr.CCidx {
			continue
		}
		pr.CC[i].ID = ccid
		pr.CC[i].Idx = pr.CCidx
		pr.CCstack = append(pr.CCstack, i)
		for len(pr.CCstack) > 0 {
			idx := pr.CCstack[len(pr.CCstack)-1]
			pr.CCstack = pr.CCstack[:len(pr.CCstack)-1]
			p := idxToPos(idx, w)
			for _, q := range nb.Neighbors(p) {
				if !q.In(pr.Rg) {
					continue
				}
				nidx := pr.idx(q)
				if pr.CC[nidx].Idx == pr.CCidx {
					continue
				}
				pr.CC[nidx].ID = ccid
				pr.CC[nidx].Idx = pr.CCidx
				pr.CCstack = append(pr.CCstack, nidx)
			}
		}
		ccid++
	}
}

// ComputeCC computes the connected component which contains a given position.
// It makes the assumption that the paths are bidirectional, allowing for
// efficient computation. It makes uses of the same caching structures as
// ComputeCCAll, so CCAt will return -1 on all unreachable positions from p.
func (pr *PathRange) ComputeCC(nb Pather, p gruid.Point) {
	max := pr.Rg.Size()
	w, h := max.X, max.Y
	if pr.CC == nil {
		pr.CC = make([]ccNode, w*h)
	}
	pr.CCidx++
	defer pr.checkCCIdx()
	if pr.CCIterCache == nil {
		pr.CCIterCache = make([]int, w*h)
	}
	pr.CCIterCache = pr.CCIterCache[:0]
	pr.CCstack = pr.CCstack[:0]
	ccid := 0
	if !p.In(pr.Range()) {
		return
	}
	idx := pr.idx(p)
	pr.CC[idx].ID = ccid
	pr.CC[idx].Idx = pr.CCidx
	pr.CCstack = append(pr.CCstack, idx)
	for len(pr.CCstack) > 0 {
		idx = pr.CCstack[len(pr.CCstack)-1]
		pr.CCstack = pr.CCstack[:len(pr.CCstack)-1]
		pr.CCIterCache = append(pr.CCIterCache, idx)
		p := idxToPos(idx, w)
		for _, q := range nb.Neighbors(p) {
			if !q.In(pr.Rg) {
				continue
			}
			nidx := pr.idx(q)
			if pr.CC[nidx].Idx == pr.CCidx {
				continue
			}
			pr.CC[nidx].ID = ccid
			pr.CC[nidx].Idx = pr.CCidx
			pr.CCstack = append(pr.CCstack, nidx)
		}
	}
}

// CCAt returns a positive number identifying the position's connected
// component as computed by either the last ComputeCC or ComputeCCAll call. It
// returns -1 on out of range positions.
func (pr *PathRange) CCAt(p gruid.Point) int {
	if !p.In(pr.Rg) || pr.CC == nil {
		return -1
	}
	node := pr.CC[pr.idx(p)]
	if node.Idx != pr.CCidx {
		return -1
	}
	return node.ID
}

// CCIter iterates a function on all the positions belonging to the connected
// component computed by the last ComputeCC call.  Caching is used for
// efficiency, so the iteration function should avoid calling CCIter and
// ComputeCC.
func (pr *PathRange) CCIter(fn func(gruid.Point)) {
	if pr.CCIterCache == nil {
		return
	}
	w := pr.Rg.Size().X
	for _, idx := range pr.CCIterCache {
		p := idxToPos(idx, w)
		fn(p)
	}
}

func (pr *PathRange) checkCCIdx() {
	if pr.CCidx < math.MaxInt32 {
		return
	}
	for i, n := range pr.CC {
		idx := 0
		if n.Idx == pr.CCidx {
			idx = 1
		}
		pr.CC[i] = ccNode{ID: n.ID, Idx: idx}
	}
	pr.CCidx = 1
}
