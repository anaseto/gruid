// Package paths provides utilities for efficient pathfinding in rectangular maps.
package paths

import "github.com/anaseto/gruid"

// PathFinder allows for efficient path finding within a range.
type PathFinder struct {
	rg               gruid.Range
	astarNodes       *nodeMap
	astarQueue       priorityQueue
	dijkstraNodes    *nodeMap // dijkstra map
	dijkstraQueue    priorityQueue
	iterVisitedCache []int
	iterQueueCache   []int
	dijkstra         Dijkstra // used by MapIter
	bfmap            []int    // breadth first map
	bfvisited        []bool
	bfqueue          []int
}

// NewPathFinder returns a new PathFinder for positions in a given range.
func NewPathFinder(rg gruid.Range) *PathFinder {
	pf := &PathFinder{}
	pf.rg = rg
	return pf
}

// SetRange updates the range used by the PathFinder. If the size is the same,
// cached structures will be preserved, otherwise they will be reinitialized.
func (pf *PathFinder) SetRange(rg gruid.Range) {
	org := pf.rg
	pf.rg = rg
	w, h := rg.Size()
	ow, oh := org.Size()
	if w == ow && h == oh {
		return
	}
	*pf = PathFinder{rg: rg}
}

func (pf *PathFinder) idx(pos gruid.Position) int {
	w, _ := pf.rg.Size()
	return pos.Y*w + pos.X
}

func (nm nodeMap) get(pf *PathFinder, p gruid.Position) *node {
	n := &nm.Nodes[pf.idx(p)]
	if n.CacheIndex != nm.Index {
		nm.Nodes[pf.idx(p)] = node{Pos: p, CacheIndex: nm.Index}
	}
	return n
}

func (nm nodeMap) at(pf *PathFinder, p gruid.Position) (*node, bool) {
	n := &nm.Nodes[pf.idx(p)]
	if n.CacheIndex != nm.Index {
		return nil, false
	}
	return n, true
}

type node struct {
	Pos        gruid.Position
	Cost       int
	Rank       int
	Parent     *gruid.Position
	Open       bool
	Closed     bool
	Index      int
	Num        int
	CacheIndex int
}

type nodeMap struct {
	Nodes []node
	Index int
}

// A priorityQueue implements heap.Interface with Node elements.
type priorityQueue []*node

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Rank < pq[j].Rank || pq[i].Rank == pq[j].Rank && pq[i].Num < pq[j].Num
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	no := x.(*node)
	no.Index = n
	*pq = append(*pq, no)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	no := old[n-1]
	no.Index = -1
	*pq = old[0 : n-1]
	return no
}
