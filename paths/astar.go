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

// Astar is the interface that allows to use the A* algorithm used by the
// AstarPath function.
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
func (pr *PathRange) AstarPath(ast Astar, from, to gruid.Position) []gruid.Position {
	if !from.In(pr.rg) || !to.In(pr.rg) {
		return nil
	}
	if pr.astarNodes == nil {
		pr.astarNodes = &nodeMap{}
		w, h := pr.rg.Size()
		pr.astarNodes.Nodes = make([]node, w*h)
		pr.astarQueue = make(priorityQueue, 0, w*h)
	}
	nm := pr.astarNodes
	nm.Index++
	nqs := pr.astarQueue[:0]
	nq := &nqs
	heap.Init(nq)
	fromNode := nm.get(pr, from)
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
				curr, _ = nm.at(pr, *curr.Parent)
			}
			return p
		}

		for _, neighbor := range ast.Neighbors(current.Pos) {
			if !neighbor.In(pr.rg) {
				continue
			}
			cost := current.Cost + ast.Cost(current.Pos, neighbor)
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
