// Package rl will provide some facilities for common roguelike programming
// needs. It is usable, but still EXPERIMENTAL, do not expect the API to be
// stable.
package rl

import (
	"bytes"
	"container/heap"
	"encoding/gob"
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
// EventQueue implements gob.Decoder and gob.Encoder for easy serialization.
type EventQueue struct {
	eventQueue
}

type eventQueue struct {
	Queue *evSliceQueue
	Idx   int
}

// NewEventQueue returns a new EventQueue suitable for use.
func NewEventQueue() *EventQueue {
	q := &evSliceQueue{}
	heap.Init(q)
	return &EventQueue{eventQueue{
		Queue: q,
	}}
}

// GobDecode implements gob.GobDecoder.
func (eq *EventQueue) GobDecode(bs []byte) error {
	r := bytes.NewReader(bs)
	gdec := gob.NewDecoder(r)
	ievq := &eventQueue{}
	err := gdec.Decode(ievq)
	if err != nil {
		return err
	}
	eq.eventQueue = *ievq
	return nil
}

// GobEncode implements gob.GobEncoder.
func (eq *EventQueue) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	ge := gob.NewEncoder(&buf)
	err := ge.Encode(&eq.eventQueue)
	return buf.Bytes(), err
}

// Push adds a new event to the heap with a given rank. Events with the same
// rank are processed in a first-in first-out order.
func (eq *EventQueue) Push(ev Event, rank int) {
	evr := event{Event: ev, Rank: rank, Idx: eq.Idx}
	eq.Idx++
	heap.Push(eq.Queue, evr)
	if eq.Idx == math.MaxInt32 {
		// should not happen in practical situations
		eq.Filter(func(ev Event) bool { return true })
	}
}

// Empty reports whether the event queue is empty.
func (eq *EventQueue) Empty() bool {
	return eq.Queue.Len() <= 0
}

// Filter removes events that do not satisfy a given predicate from the event
// queue.
func (eq *EventQueue) Filter(fn func(ev Event) bool) {
	eq.Idx = 0
	ievs := []event{}
	for !eq.Empty() {
		evr := eq.popIEvent()
		if fn(evr.Event) {
			ievs = append(ievs, evr)
		}
	}
	*eq.Queue = (*eq.Queue)[:0]
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
	evr := heap.Pop(eq.Queue).(event)
	return evr
}

// evSliceQueue implements heap.Interface.
type evSliceQueue []event

func (evq evSliceQueue) Len() int {
	return len(evq)
}

func (evq evSliceQueue) Less(i, j int) bool {
	return evq[i].Rank < evq[j].Rank ||
		evq[i].Rank == evq[j].Rank && evq[i].Idx < evq[j].Idx
}

func (evq evSliceQueue) Swap(i, j int) {
	evq[i], evq[j] = evq[j], evq[i]
}

func (evq *evSliceQueue) Push(x interface{}) {
	no := x.(event)
	*evq = append(*evq, no)
}

func (evq *evSliceQueue) Pop() interface{} {
	old := *evq
	n := len(old)
	no := old[n-1]
	*evq = old[0 : n-1]
	return no
}
