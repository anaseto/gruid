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
	Neighbors(gruid.Point) []gruid.Point

	// Cost represents the cost from one position to an adjacent one. It
	// should not produce paths with negative costs.
	Cost(gruid.Point, gruid.Point) int
}

// DijkstraMap computes a dijkstra map given a list of source positions and a
// maximal cost from those sources. The resulting map can then be iterated with
// Iter.
func (pr *PathRange) DijkstraMap(dij Dijkstra, sources []gruid.Point, maxCost int) {
	if pr.DijkstraNodes == nil {
		pr.DijkstraNodes = &nodeMap{}
		max := pr.Rg.Size()
		pr.DijkstraNodes.Nodes = make([]node, max.X*max.Y)
		pr.DijkstraQueue = make(priorityQueue, 0, max.X*max.Y)
		pr.IterNodeCache = []Node{}
	}
	pr.IterNodeCache = pr.IterNodeCache[:0]
	nm := pr.DijkstraNodes
	nm.Idx++
	defer checkNodesIdx(nm)
	nqs := pr.DijkstraQueue[:0]
	nq := &nqs
	heap.Init(nq)
	for _, f := range sources {
		if !f.In(pr.Rg) {
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
		n := heap.Pop(nq).(*node)
		n.Open = false
		n.Closed = true
		pr.IterNodeCache = append(pr.IterNodeCache, Node{P: n.P, Cost: n.Cost})

		for _, nb := range dij.Neighbors(n.P) {
			if !nb.In(pr.Rg) {
				continue
			}
			cost := n.Cost + dij.Cost(n.P, nb)
			nbNode := nm.get(pr, nb)
			if cost < nbNode.Cost {
				if nbNode.Open {
					heap.Remove(nq, nbNode.Idx)
				}
				nbNode.Open = false
				nbNode.Closed = false
			}
			if !nbNode.Open && !nbNode.Closed {
				nbNode.Cost = cost
				if cost <= maxCost {
					nbNode.Open = true
					nbNode.Rank = cost
					heap.Push(nq, nbNode)
				}
			}
		}
	}
}

// Node represents a position in a dijkstra map with a related distance cost
// relative to the most close source.
type Node struct {
	P    gruid.Point
	Cost int
}

// idxToPos returns a grid position given an index and the width of the grid.
func idxToPos(i, w int) gruid.Point {
	return gruid.Point{X: i - (i/w)*w, Y: i / w}
}

// MapIter iterates a function on the nodes of the last computed dijkstra map,
// in cost increasing order.  Note that you should not call the MapIter or
// DijkstraMap methods on the same PathFinder within the iteration function, as
// that could invalidate the iteration state.
func (pr *PathRange) MapIter(f func(Node)) {
	for _, n := range pr.IterNodeCache {
		f(n)
	}
}
