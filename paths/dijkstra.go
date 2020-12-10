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
func (pf *PathFinder) DijkstraMap(dij Dijkstra, sources []gruid.Position, maxCost int) {
	if pf.dijkstraNodes == nil {
		pf.dijkstraNodes = &nodeMap{}
		w, h := pf.rg.Size()
		pf.dijkstraNodes.Nodes = make([]node, w*h)
		pf.dijkstraQueue = make(priorityQueue, 0, w*h)
	}
	nm := pf.dijkstraNodes
	pf.dijkstra = dij
	nm.Index++
	nqs := pf.dijkstraQueue[:0]
	nq := &nqs
	heap.Init(nq)
	for _, f := range sources {
		if !f.In(pf.rg) {
			continue
		}
		n := nm.get(pf, f)
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

		for _, neighbor := range dij.Neighbors(current.Pos) {
			if !neighbor.In(pf.rg) {
				continue
			}
			cost := current.Cost + dij.Cost(current.Pos, neighbor)
			neighborNode := nm.get(pf, neighbor)
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

// MapIter iterates last computed dijkstra map from a given position in breadth
// first order. Note that you should not call the MapIter or DijkstraMap
// methods on the same PathFinder within the iteration function, as that could
// invalidate the iteration state.
func (pf *PathFinder) MapIter(pos gruid.Position, f func(Node)) {
	if pf.dijkstra == nil || !pos.In(pf.rg) {
		return
	}
	if pf.iterQueueCache == nil {
		w, h := pf.rg.Size()
		pf.iterQueueCache = make([]int, w*h)
		pf.iterVisitedCache = make([]int, w*h)
	}
	nm := pf.dijkstraNodes
	var qstart, qend int
	pf.iterQueueCache[qend] = pf.idx(pos)
	pf.iterVisitedCache[qend] = nm.Index
	qend++
	_, w := pf.rg.Size()
	for qstart < qend {
		pos = idxToPos(pf.iterQueueCache[qstart], w)
		qstart++
		nb := pf.dijkstra.Neighbors(pos)
		for _, npos := range nb {
			if !npos.In(pf.rg) {
				continue
			}
			n := &nm.Nodes[pf.idx(npos)]
			if n.CacheIndex == nm.Index && pf.iterVisitedCache[pf.idx(npos)] != nm.Index {
				f(Node{Pos: n.Pos, Cost: n.Cost})
				pf.iterQueueCache[qend] = pf.idx(npos)
				qend++
				pf.iterVisitedCache[pf.idx(npos)] = nm.Index
			}
		}
	}
}
