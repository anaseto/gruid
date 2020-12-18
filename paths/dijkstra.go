package paths

import (
	"container/heap"

	"github.com/anaseto/gruid"
)

// Dijkstra is the interface that allows to build a dijkstra map using the
// DijkstraMap function.
type Dijkstra interface {
	// Neighbors returns the available neighbor positions of a given
	// position. Implementations may use a cache to avoid allocations.
	Neighbors(gruid.Position) []gruid.Position

	// Cost represents the cost from one position to an adjacent one. It
	// should not produce paths with negative costs.
	Cost(gruid.Position, gruid.Position) int
}

// DijkstraMap computes a dijkstra map given a list of source positions and a
// maximal cost from those sources. The resulting map can then be iterated with
// Iter.
func (pr *PathRange) DijkstraMap(dij Dijkstra, sources []gruid.Position, maxCost int) {
	if pr.dijkstraNodes == nil {
		pr.dijkstraNodes = &nodeMap{}
		w, h := pr.rg.Size()
		pr.dijkstraNodes.Nodes = make([]node, w*h)
		pr.dijkstraQueue = make(priorityQueue, 0, w*h)
		pr.iterNodeCache = []Node{}
	}
	pr.iterNodeCache = pr.iterNodeCache[:0]
	nm := pr.dijkstraNodes
	pr.dijkstra = dij
	nm.Index++
	nqs := pr.dijkstraQueue[:0]
	nq := &nqs
	heap.Init(nq)
	for _, f := range sources {
		if !f.In(pr.rg) {
			continue
		}
		n := nm.get(pr, f)
		n.Open = true
		heap.Push(nq, n)
	}
	for {
		if nq.Len() == 0 {
			return
		}
		current := heap.Pop(nq).(*node)
		current.Open = false
		current.Closed = true
		pr.iterNodeCache = append(pr.iterNodeCache, Node{Pos: current.Pos, Cost: current.Cost})

		for _, neighbor := range dij.Neighbors(current.Pos) {
			if !neighbor.In(pr.rg) {
				continue
			}
			cost := current.Cost + dij.Cost(current.Pos, neighbor)
			neighborNode := nm.get(pr, neighbor)
			if cost < neighborNode.Cost {
				if neighborNode.Open {
					heap.Remove(nq, neighborNode.Index)
				}
				neighborNode.Open = false
				neighborNode.Closed = false
			}
			if !neighborNode.Open && !neighborNode.Closed {
				neighborNode.Cost = cost
				if cost < maxCost {
					neighborNode.Open = true
					neighborNode.Rank = cost
					heap.Push(nq, neighborNode)
				}
			}
		}
	}
}

// Node represents a position in a dijkstra map with a related distance cost
// relative to the most close source.
type Node struct {
	Pos  gruid.Position
	Cost int
}

// idxToPos returns a grid position given an index and the width of the grid.
func idxToPos(i, w int) gruid.Position {
	return gruid.Position{X: i - (i/w)*w, Y: i / w}
}

// MapIter iterates a function on the nodes of the last computed dijkstra map,
// in cost increasing order.  Note that you should not call the MapIter or
// DijkstraMap methods on the same PathFinder within the iteration function, as
// that could invalidate the iteration state.
func (pr *PathRange) MapIter(f func(Node)) {
	for _, n := range pr.iterNodeCache {
		f(n)
	}
}
