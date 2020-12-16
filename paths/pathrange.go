// Package paths provides utilities for efficient pathfinding in rectangular maps.
package paths

import "github.com/anaseto/gruid"

// PathRange allows for efficient path finding within a range. It caches
// structures, so that they can be reused without further memory allocations.
type PathRange struct {
	rg            gruid.Range
	astarNodes    *nodeMap
	astarQueue    priorityQueue
	dijkstraNodes *nodeMap // dijkstra map
	dijkstraQueue priorityQueue
	iterNodeCache []Node
	dijkstra      Dijkstra // used by MapIter
	bfmap         []bfNode // breadth first map
	bfqueue       []int
	bfunreachable int      // last maxcost + 1
	bfidx         int      // bf map number
	cc            []ccNode // connected components
	ccstack       []int
	ccidx         int // cc map number
	ccIterCache   []int
	neighbors     Neighborer
}

// NewPathRange returns a new PathFinder for positions in a given range,
// such as the range occupied by the whole map, or a part of it.
func NewPathRange(rg gruid.Range) *PathRange {
	pf := &PathRange{}
	pf.rg = rg
	return pf
}

// SetRange updates the range used by the PathFinder. If the size is the same,
// cached structures will be preserved, otherwise they will be reinitialized.
func (pf *PathRange) SetRange(rg gruid.Range) {
	org := pf.rg
	pf.rg = rg
	w, h := rg.Size()
	ow, oh := org.Size()
	if w == ow && h == oh {
		return
	}
	*pf = PathRange{rg: rg}
}

func (pf *PathRange) idx(pos gruid.Position) int {
	pos = pos.Relative(pf.rg)
	w, _ := pf.rg.Size()
	return pos.Y*w + pos.X
}

func (nm nodeMap) get(pf *PathRange, pos gruid.Position) *node {
	n := &nm.Nodes[pf.idx(pos)]
	if n.CacheIndex != nm.Index {
		nm.Nodes[pf.idx(pos)] = node{Pos: pos, CacheIndex: nm.Index}
	}
	return n
}

func (nm nodeMap) at(pf *PathRange, pos gruid.Position) (*node, bool) {
	n := &nm.Nodes[pf.idx(pos)]
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
