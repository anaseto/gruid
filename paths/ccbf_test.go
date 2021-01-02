package paths

import (
	"bytes"
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
	pr.ComputeCCAll(nb)
	rg := pr.Rg
	p := gruid.Point{X: rg.Min.X, Y: rg.Min.Y}
	id := pr.CCAt(p)
	for y := rg.Min.Y + 1; y < rg.Max.Y; y++ {
		p := gruid.Point{X: rg.Min.X, Y: y}
		nid := pr.CCAt(p)
		if id == nid {
			t.Errorf("same id on different lines: %d, %d", id, nid)
		}
		if nid != y-rg.Min.Y {
			t.Errorf("bad id: %d, %d", id, y-rg.Min.Y)
		}
		id = nid
	}
	id = pr.CCAt(p)
	for y := rg.Min.Y; y < rg.Max.Y; y++ {
		p := gruid.Point{X: rg.Min.X, Y: y}
		id := pr.CCAt(p)
		for x := rg.Min.X; x < rg.Max.X; x++ {
			if id != pr.CCAt(gruid.Point{X: x, Y: y}) {
				t.Errorf("different id on same line: %d, %d", id, pr.CCAt(gruid.Point{X: x, Y: y}))
			}
		}
	}
	pr.ComputeCC(nb, gruid.Point{X: 1, Y: 1})
	count := 0
	pr.CCIter(func(p gruid.Point) {
		count++
		if p.Y != 1 {
			t.Errorf("bad id on line 1: %d", id)
		}
	})
	if count != 10 {
		t.Errorf("bad count: %d", count)
	}
	poscosts := []struct {
		p    gruid.Point
		cost int
	}{
		{gruid.Point{0, 0}, 2},
		{gruid.Point{1, 0}, 1},
		{gruid.Point{2, 0}, 0},
		{gruid.Point{3, 0}, 1},
		{gruid.Point{4, 0}, 2},
		{gruid.Point{5, 0}, 3},
		{gruid.Point{6, 0}, 4},
		{gruid.Point{7, 0}, 4},
		{gruid.Point{0, 2}, 2},
		{gruid.Point{1, 2}, 1},
		{gruid.Point{2, 2}, 0},
		{gruid.Point{3, 2}, 1},
		{gruid.Point{4, 2}, 2},
		{gruid.Point{5, 2}, 3},
		{gruid.Point{6, 2}, 4},
		{gruid.Point{7, 2}, 4},
		{gruid.Point{0, 1}, 4},
		{gruid.Point{1, 1}, 4},
		{gruid.Point{2, 1}, 4},
		{gruid.Point{3, 1}, 4},
		{gruid.Point{4, 1}, 4},
		{gruid.Point{5, 1}, 4},
		{gruid.Point{6, 1}, 4},
	}
	for i := 0; i < 2; i++ {
		pr.BreadthFirstMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 3)
		for _, pc := range poscosts {
			if pc.cost != pr.CostAt(pc.p) {
				t.Errorf("bad cost %d for %+v", pc.cost, pc.p)
			}
		}
		pr.DijkstraMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 9)
		pr.MapIter(func(n Node) {
			for _, pc := range poscosts {
				if pc.p == n.P && 2*pc.cost != n.Cost {
					t.Errorf("bad cost %d for %+v", n.Cost, n.P)
				}
			}
		})
	}
}
