package paths

import (
	"bytes"
	//"fmt"
	"testing"

	"encoding/gob"
	"github.com/anaseto/gruid"
)

type npath struct {
	nb Neighbors
}

func (nb npath) Neighbors(p gruid.Point) []gruid.Point {
	return nb.nb.All(p, func(q gruid.Point) bool {
		// strange Neighborer that allows only horizontal moves
		return q.Y == p.Y
	})
}

func (nb npath) Cost(p, q gruid.Point) int {
	return 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (nb npath) Estimation(p, q gruid.Point) int {
	r := p.Sub(q)
	return abs(r.X) + abs(r.Y)
}

func TestAstar(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
	path := pr.AstarPath(nb, gruid.Point{0, 0}, gruid.Point{4, 0})
	if len(path) != 5 {
		t.Errorf("bad length: %d", len(path))
	}
	path = pr.AstarPath(nb, gruid.Point{0, 0}, gruid.Point{0, 1})
	if len(path) != 0 {
		t.Errorf("not empty path: %d", len(path))
	}
}

func TestGob(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
	path := pr.AstarPath(nb, gruid.Point{0, 0}, gruid.Point{4, 0})
	if len(path) != 5 {
		t.Errorf("bad length: %d", len(path))
	}
	buf := bytes.Buffer{}
	ge := gob.NewEncoder(&buf)
	err := ge.Encode(pr)
	if err != nil {
		t.Error(err)
	}
	pr = &PathRange{}
	gd := gob.NewDecoder(&buf)
	err = gd.Decode(pr)
	if err != nil {
		t.Error(err)
	}
	path = pr.AstarPath(nb, gruid.Point{0, 0}, gruid.Point{5, 0})
	if len(path) != 6 {
		t.Errorf("bad length: %d", len(path))
	}
}

func TestCCBf(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
	pr.CCMapAll(nb)
	rg := pr.Rg
	p := gruid.Point{X: rg.Min.X, Y: rg.Min.Y}
	id := pr.CCMapAt(p)
	for y := rg.Min.Y + 1; y < rg.Max.Y; y++ {
		p := gruid.Point{X: rg.Min.X, Y: y}
		nid := pr.CCMapAt(p)
		if id == nid {
			t.Errorf("same id on different lines: %d, %d", id, nid)
		}
		if nid != y-rg.Min.Y {
			t.Errorf("bad id: %d, %d", id, y-rg.Min.Y)
		}
		id = nid
	}
	id = pr.CCMapAt(p)
	for y := rg.Min.Y; y < rg.Max.Y; y++ {
		p := gruid.Point{X: rg.Min.X, Y: y}
		id := pr.CCMapAt(p)
		for x := rg.Min.X; x < rg.Max.X; x++ {
			if id != pr.CCMapAt(gruid.Point{X: x, Y: y}) {
				t.Errorf("different id on same line: %d, %d", id, pr.CCMapAt(gruid.Point{X: x, Y: y}))
			}
		}
	}
	count := 0
	for _, p := range pr.CCMap(nb, gruid.Point{X: 1, Y: 1}) {
		count++
		if p.Y != 1 {
			t.Errorf("bad id on line 1: %d", id)
		}
	}
	if count != 10 {
		t.Errorf("bad count: %d", count)
	}
}

func TestCCBfOutOfRange(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
	p := gruid.Point{-1, -1}
	pr.CCMapAll(nb)
	pr.CCMap(nb, p)
	if pr.CCMapAt(p) != -1 {
		t.Errorf("bad out of range value: %v", pr.CCMapAt(p))
	}
	p = gruid.Point{4, 0}
	if pr.CCMapAt(p) != -1 {
		t.Errorf("bad unreachable value: %v", pr.CCMapAt(p))
	}
	q := gruid.Point{6, 2}
	pr.CCMap(nb, p)
	if pr.CCMapAt(q) != -1 {
		t.Errorf("bad unreachable value: %v", pr.CCMapAt(q))
	}
}

type bpath struct {
	nb *Neighbors
}

func (nb bpath) Neighbors(p gruid.Point) []gruid.Point {
	return nb.nb.All(p, func(q gruid.Point) bool {
		return true
	})
}

func (nb bpath) Cost(p, q gruid.Point) int {
	return 1
}

func (nb bpath) Estimation(p, q gruid.Point) int {
	p = p.Sub(q)
	return abs(p.X) + abs(p.Y)
}

func BenchmarkCCMapAll(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.CCMapAll(nb)
	}
}

func BenchmarkCCMap(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.CCMap(nb, gruid.Point{5, 5})
	}
}
