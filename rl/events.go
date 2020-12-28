// Package rl will provide some facilities for common roguelike programming
// needs. It is still EXPERIMENTAL, do not use for production.
package rl

import (
	"container/heap"
	"math"
)

// Event represents a message or an action.
type Event interface{}

type event struct {
	Event Event // underlying event
	Rank  int   // rank in the priority queue (lower is better)

	// Idx is a secondary rank that allows to ensure FIFO order on events
	// of same rank.
	Idx int
}

// EventQueue manages a priority queue of Event elements sorted by their
// rank in ascending order. The rank may represent for example a turn number or
// some more fine-grained priority measure.
//
// EventQueue must be created with NewEventQueue.
//
// TODO: make it possible to encode it with encoding/gob.
type EventQueue struct {
	queue *eventQueue
	idx   int
}

// NewEventQueue returns a new EventQueue suitable for use.
func NewEventQueue() *EventQueue {
	q := &eventQueue{}
	heap.Init(q)
	return &EventQueue{
		queue: q,
	}
}

// Push adds a new event to the heap with a given rank.
func (eq *EventQueue) Push(ev Event, rank int) {
	evr := event{Event: ev, Rank: rank, Idx: eq.idx}
	eq.idx++
	heap.Push(eq.queue, evr)
	if eq.idx == math.MaxInt32 {
		// should not happen in practical situations
		eq.Filter(func(ev Event) bool { return true })
	}
}

// Empty reports whether the event queue is empty.
func (eq *EventQueue) Empty() bool {
	return eq.queue.Len() <= 0
}

// Filter removes events that do not satisfy a given predicate from the event
// queue.
func (eq *EventQueue) Filter(fn func(ev Event) bool) {
	eq.idx = 0
	ievs := []event{}
	for !eq.Empty() {
		evr := eq.popIEvent()
		if fn(evr.Event) {
			ievs = append(ievs, evr)
		}
	}
	*eq.queue = (*eq.queue)[:0]
	for _, evr := range ievs {
		eq.Push(evr.Event, evr.Rank)
	}
}

// Pop removes the first element rank-wise (lowest rank) in the event queue and
// returns it.
func (eq *EventQueue) Pop() Event {
	return eq.popIEvent().Event
}

func (eq *EventQueue) popIEvent() event {
	evr := heap.Pop(eq.queue).(event)
	return evr
}

// eventQueue implements heap.Interface.
type eventQueue []event

func (evq eventQueue) Len() int {
	return len(evq)
}

func (evq eventQueue) Less(i, j int) bool {
	return evq[i].Rank < evq[j].Rank ||
		evq[i].Rank == evq[j].Rank && evq[i].Idx < evq[j].Idx
}

func (evq eventQueue) Swap(i, j int) {
	evq[i], evq[j] = evq[j], evq[i]
}

func (evq *eventQueue) Push(x interface{}) {
	no := x.(event)
	*evq = append(*evq, no)
}

func (evq *eventQueue) Pop() interface{} {
	old := *evq
	n := len(old)
	no := old[n-1]
	*evq = old[0 : n-1]
	return no
}
