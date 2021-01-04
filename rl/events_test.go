package rl

import (
	"testing"
)

func TestEventsQueuePush(t *testing.T) {
	eq := NewEventQueue()
	eq.Push(3, 1)
	eq.Push(1, 3)
	eq.Push(2, 2)
	count := 3
	for !eq.Empty() {
		ev := eq.Pop()
		switch n := ev.(type) {
		case int:
			if n != count {
				t.Errorf("bad number: %d vs %d", n, count)
			}
			count--
		default:
			t.Errorf("bad event: %+v", n)
		}
	}
}

func TestEventsQueueFilter(t *testing.T) {
	eq := NewEventQueue()
	eq.Push(3, 1)
	eq.Push(1, 3)
	eq.Push(2, 2)
	eq.Filter(func(ev Event) bool {
		switch n := ev.(type) {
		case int:
			return n == 3
		default:
			return false
		}
	})
	if eq.Empty() {
		t.Errorf("empty: %+v", eq)
	}
	if eq.Queue.Len() != 1 {
		t.Errorf("bad length: %d vs 1", eq.Queue.Len())
	}
	n := eq.Pop().(int)
	if n != 3 {
		t.Errorf("bad number: %d vs 3", n)
	}
	if !eq.Empty() {
		t.Errorf("not empty: %+v", eq)
	}
}
