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
	neighbors     Pather
}

// NewPathRange returns a new PathFinder for positions in a given range,
// such as the range occupied by the whole map, or a part of it.
func NewPathRange(rg gruid.Range) *PathRange {
	pr := &PathRange{}
	pr.rg = rg
	return pr
}

// SetRange updates the range used by the PathFinder. If the size is the same,
// cached structures will be preserved, otherwise they will be reinitialized.
func (pr *PathRange) SetRange(rg gruid.Range) {
	org := pr.rg
	pr.rg = rg
	max := rg.Size()
	omax := org.Size()
	if max == omax {
		return
	}
	*pr = PathRange{rg: rg}
}

func (pr *PathRange) idx(p gruid.Point) int {
	p = p.Rel(pr.rg)
	w := pr.rg.Size().X
	return p.Y*w + p.X
}

func (nm nodeMap) get(pr *PathRange, p gruid.Point) *node {
	n := &nm.Nodes[pr.idx(p)]
	if n.CacheIndex != nm.Index {
		nm.Nodes[pr.idx(p)] = node{P: p, CacheIndex: nm.Index}
	}
	return n
}

func (nm nodeMap) at(pr *PathRange, p gruid.Point) (*node, bool) {
	n := &nm.Nodes[pr.idx(p)]
	if n.CacheIndex != nm.Index {
		return nil, false
	}
	return n, true
}

type node struct {
	P          gruid.Point
	Cost       int
	Rank       int
	Parent     *gruid.Point
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
