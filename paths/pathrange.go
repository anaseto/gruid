// Package paths provides utilities for efficient pathfinding in rectangular maps.
package paths

import (
	"bytes"
	"encoding/gob"

	"github.com/anaseto/gruid"
)

// PathRange allows for efficient path finding within a range. It caches
// structures, so that they can be reused without further memory allocations.
//
// It implements gob.Encoder and gob.Decoder for easy serialization.
type PathRange struct {
	pathRange
}

type pathRange struct {
	Rg            gruid.Range
	AstarNodes    *nodeMap
	AstarQueue    priorityQueue
	DijkstraNodes *nodeMap // dijkstra map
	DijkstraQueue priorityQueue
	IterNodeCache []Node
	Bfmap         []bfNode // breadth first map
	Bfqueue       []int
	Bfunreachable int      // last maxcost + 1
	Bfidx         int      // bf map number
	CC            []ccNode // connected components
	CCstack       []int
	CCidx         int // cc map number
	CCIterCache   []int
}

// GobDecode implements gob.GobDecoder.
func (pr *PathRange) GobDecode(bs []byte) error {
	r := bytes.NewReader(bs)
	gd := gob.NewDecoder(r)
	ipr := &pathRange{}
	err := gd.Decode(ipr)
	if err != nil {
		return err
	}
	pr.pathRange = *ipr
	return nil
}

// GobEncode implements gob.GobEncoder.
func (pr *PathRange) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	ge := gob.NewEncoder(&buf)
	err := ge.Encode(&pr.pathRange)
	return buf.Bytes(), err
}

// NewPathRange returns a new PathFinder for positions in a given range,
// such as the range occupied by the whole map, or a part of it.
func NewPathRange(rg gruid.Range) *PathRange {
	pr := &PathRange{}
	pr.Rg = rg
	return pr
}

// SetRange updates the range used by the PathFinder. If the size is the same,
// cached structures will be preserved, otherwise they will be reinitialized.
func (pr *PathRange) SetRange(rg gruid.Range) {
	org := pr.Rg
	pr.Rg = rg
	max := rg.Size()
	omax := org.Size()
	if max == omax {
		return
	}
	*pr = PathRange{}
	pr.Rg = rg
}

// Range returns the current PathRange's range of positions.
func (pr *PathRange) Range() gruid.Range {
	return pr.Rg
}

func (pr *PathRange) idx(p gruid.Point) int {
	p = p.Sub(pr.Rg.Min)
	w := pr.Rg.Max.X - pr.Rg.Min.X
	return p.Y*w + p.X
}

func (nm nodeMap) get(pr *PathRange, p gruid.Point) *node {
	n := &nm.Nodes[pr.idx(p)]
	if n.CacheIndex != nm.Idx {
		nm.Nodes[pr.idx(p)] = node{P: p, CacheIndex: nm.Idx}
	}
	return n
}

func (nm nodeMap) at(pr *PathRange, p gruid.Point) (*node, bool) {
	n := &nm.Nodes[pr.idx(p)]
	if n.CacheIndex != nm.Idx {
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
	Idx        int
	Num        int
	CacheIndex int
}

type nodeMap struct {
	Nodes []node
	Idx   int
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
	pq[i].Idx = i
	pq[j].Idx = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	no := x.(*node)
	no.Idx = n
	*pq = append(*pq, no)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	no := old[n-1]
	no.Idx = -1
	*pq = old[0 : n-1]
	return no
}
