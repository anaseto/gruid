package paths

import (
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

// DijkstraMapAt returns the cost associated to a position in the last computed
// Dijkstra map. It returns maxCost + 1 if the position is out of range.
func (pr *PathRange) DijkstraMapAt(p gruid.Point) int {
	n := pr.DijkstraNodes.at(pr, p)
	if n == nil {
		return pr.DijkstraUnreachable
	}
	return n.Cost
}

// DijkstraMap computes a dijkstra map given a list of source positions and a
// maximal cost from those sources. It returns a slice with the nodes of the
// map, in cost increasing order. The resulting slice is cached for efficiency,
// so future calls to DijkstraMap will invalidate its contents.
func (pr *PathRange) DijkstraMap(dij Dijkstra, sources []gruid.Point, maxCost int) []Node {
	if pr.DijkstraNodes == nil {
		pr.DijkstraNodes = &nodeMap{}
		max := pr.Rg.Size()
		pr.DijkstraNodes.Nodes = make([]node, max.X*max.Y)
		pr.DijkstraQueue = make(priorityQueue, 0, max.X*max.Y)
		pr.DijkstraIterNodes = []Node{}
	}
	pr.DijkstraUnreachable = maxCost + 1
	pr.DijkstraIterNodes = pr.DijkstraIterNodes[:0]
	nm := pr.DijkstraNodes
	nm.Idx++
	defer checkNodesIdx(nm)
	nqs := pr.DijkstraQueue[:0]
	nq := &nqs
	pqInit(nq)
	for _, f := range sources {
		if !f.In(pr.Rg) {
			continue
		}
		n := nm.get(pr, f)
		n.Open = true
		pqPush(nq, n)
	}
	for {
		if nq.Len() == 0 {
			return pr.DijkstraIterNodes
		}
		n := pqPop(nq)
		n.Open = false
		n.Closed = true
		pr.DijkstraIterNodes = append(pr.DijkstraIterNodes, Node{P: n.P, Cost: n.Cost})

		for _, q := range dij.Neighbors(n.P) {
			if !q.In(pr.Rg) {
				continue
			}
			cost := n.Cost + dij.Cost(n.P, q)
			if cost > maxCost {
				continue
			}
			nbNode := nm.get(pr, q)
			if cost < nbNode.Cost {
				if nbNode.Open {
					pqRemove(nq, nbNode.Idx)
				}
				nbNode.Open = false
				nbNode.Closed = false
			}
			if !nbNode.Open && !nbNode.Closed {
				nbNode.Cost = cost
				nbNode.Open = true
				nbNode.Rank = cost
				pqPush(nq, nbNode)
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
