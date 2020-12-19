package paths

import (
	"testing"

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

func TestCCBf(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
	pr.ComputeCCAll(nb)
	rg := pr.rg
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
	pr.BreadthFirstMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 3)
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
	for _, pc := range poscosts {
		if pc.cost != pr.CostAt(pc.p) {
			t.Errorf("bad cost %d for %+v", pc.cost, pc.p)
		}
	}
}
