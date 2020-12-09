// code of this file is a strongly modified version of code from
// github.com/beefsack/go-astar, which has the following license:
//
// Copyright (c) 2014 Michael Charles Alexander
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package paths

import (
	"container/heap"

	"github.com/anaseto/gruid"
)

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

// PathFinder allows for efficient path finding within a range.
type PathFinder struct {
	rg               gruid.Range
	nodeCache        *nodeMap
	queueCache       priorityQueue
	iterVisitedCache []int
	iterQueueCache   []int
	dijkstrer        Dijkstrer
}

// NewPathFinder returns a new PathFinder for positions in a given range.
func NewPathFinder(rg gruid.Range) *PathFinder {
	pf := &PathFinder{}
	pf.rg = rg
	w, h := rg.Size()
	pf.nodeCache.Nodes = make([]node, w*h)
	pf.queueCache = make(priorityQueue, 0, w*h)
	return pf
}

func (pf *PathFinder) idx(pos gruid.Position) int {
	w, _ := pf.rg.Size()
	return pos.Y*w + pos.X
}

func (pf *PathFinder) get(p gruid.Position) *node {
	nm := pf.nodeCache
	n := &nm.Nodes[pf.idx(p)]
	if n.CacheIndex != nm.Index {
		nm.Nodes[pf.idx(p)] = node{Pos: p, CacheIndex: nm.Index}
	}
	return n
}

func (pf *PathFinder) at(p gruid.Position) (*node, bool) {
	nm := pf.nodeCache
	n := &nm.Nodes[pf.idx(p)]
	if n.CacheIndex != nm.Index {
		return nil, false
	}
	return n, true
}

// idxToPos returns a grid position given an index and the width of the grid.
func idxToPos(i, w int) gruid.Position {
	return gruid.Position{X: i - (i/w)*w, Y: i / w}
}

// Astar is the interface that has to be satisfied in order to use the A*
// algorithm used by the AstarPath function.
type Astar interface {
	// Neighbors returns the available neighbor positions of a given
	// position. Implementations may use a cache to avoid allocations.
	Neighbors(gruid.Position) []gruid.Position

	// Cost represents the cost from one position to an adjacent one. It
	// should not produce paths with negative costs.
	Cost(gruid.Position, gruid.Position) int

	// Estimation offers an estimation cost for a path from a position to
	// another one. The estimation should always give a value lower or
	// equal to the cost of the best possible path.
	Estimation(gruid.Position, gruid.Position) int
}

// AstarPath return a path from a position to another, including thoses
// positions. It returns nil if no path was found.
func (pf *PathFinder) AstarPath(ast Astar, from, to gruid.Position) []gruid.Position {
	if !from.In(pf.rg) || !to.In(pf.rg) {
		return nil
	}
	pf.nodeCache.Index++
	nqs := pf.queueCache[:0]
	nq := &nqs
	heap.Init(nq)
	fromNode := pf.get(from)
	fromNode.Open = true
	num := 0
	fromNode.Num = num
	heap.Push(nq, fromNode)
	for {
		if nq.Len() == 0 {
			// There's no path, return found false.
			return nil
		}
		current := heap.Pop(nq).(*node)
		current.Open = false
		current.Closed = true

		if current.Pos == to {
			// Found a path to the goal.
			p := []gruid.Position{}
			curr := current
			for {
				p = append(p, curr.Pos)
				if curr.Parent == nil {
					break
				}
				curr, _ = pf.at(*curr.Parent)
			}
			return p
		}

		for _, neighbor := range ast.Neighbors(current.Pos) {
			cost := current.Cost + ast.Cost(current.Pos, neighbor)
			neighborNode := pf.get(neighbor)
			if cost < neighborNode.Cost {
				if neighborNode.Open {
					heap.Remove(nq, neighborNode.Index)
				}
				neighborNode.Open = false
				neighborNode.Closed = false
			}
			if !neighborNode.Open && !neighborNode.Closed {
				neighborNode.Cost = cost
				neighborNode.Open = true
				neighborNode.Rank = cost + ast.Estimation(neighbor, to)
				neighborNode.Parent = &current.Pos
				num++
				neighborNode.Num = num
				heap.Push(nq, neighborNode)
			}
		}
	}
}

// A priorityQueue implements heap.Interface and holds Nodes.  The
// priorityQueue is used to track open nodes by rank.
type priorityQueue []*node

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	//return pq[i].Rank < pq[j].Rank
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
