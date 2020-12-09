package paths

import (
	"container/heap"

	"github.com/anaseto/gruid"
)

// Dijkstrer is the interface that has to be satisfied in order to build a
// dijkstra map using the DijkstraMap function.
type Dijkstrer interface {
	// Neighbors returns the available neighbor positions of a given
	// position. Implementations may use a cache to avoid allocations.
	Neighbors(gruid.Position) []gruid.Position

	// Cost represents the cost from one position to an adjacent one. It
	// should not produce paths with negative costs.
	Cost(gruid.Position, gruid.Position) int
}

// Dijkstra computes a dijkstra map given a list of source positions and a
// maximal cost from those sources. The resulting map can then be iterated with
// Iter.
func (pf *PathFinder) DijkstraMap(dij Dijkstrer, sources []gruid.Position, maxCost int) {
	pf.dijkstrer = dij
	pf.nodeCache.Index++
	nqs := pf.queueCache[:0]
	nq := &nqs
	heap.Init(nq)
	for _, f := range sources {
		n := pf.get(f)
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
				if cost < maxCost {
					neighborNode.Open = true
					neighborNode.Rank = cost
					heap.Push(nq, neighborNode)
				}
			}
		}
	}
}

// Node represents a position in a dijkstra map with a related cost.
type Node struct {
	Pos  gruid.Position
	Cost int
}

// Iter iterates last computed dijkstra map from a given position.
func (pf *PathFinder) Iter(pos gruid.Position, f func(Node)) {
	if pf.dijkstrer == nil || !pos.In(pf.rg) {
		return
	}
	nm := pf.nodeCache
	var qstart, qend int
	pf.iterQueueCache[qend] = pf.idx(pos)
	pf.iterVisitedCache[qend] = nm.Index
	qend++
	_, w := pf.rg.Size()
	for qstart < qend {
		pos = idxToPos(pf.iterQueueCache[qstart], w)
		qstart++
		nb := pf.dijkstrer.Neighbors(pos)
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

const unreachable = 9999

// AutoExploreDijkstra is an optimized version of the dijkstra algorithm for
// auto-exploration.
//func (pf *PathFinder) AutoExploreDijkstra(dij Dijkstrer, sources []int) {
//dmap := DijkstraMapCache[:]
//var visited [DungeonNCells]bool
//var queue [DungeonNCells]int
//var qstart, qend int
//w, h := pf.rg.Size()
//for i := 0; i < w*h; i++ {
//dmap[i] = unreachable
//}
//for _, s := range sources {
//dmap[s] = 0
//queue[qend] = s
//qend++
//visited[s] = true
//}
//for qstart < qend {
//cidx := queue[qstart]
//qstart++
//cpos := idxtopos(cidx)
//for _, npos := range dij.Neighbors(cpos) {
//nidx := pf.idx(npos)
//if !visited[nidx] {
//queue[qend] = nidx
//qend++
//visited[nidx] = true
//dmap[nidx] = 1 + dmap[cidx]
//}
//}
//}
//}
