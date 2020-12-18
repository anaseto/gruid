package paths

import (
	"testing"

	"github.com/anaseto/gruid"
)

type npath struct {
	nb Neighbors
}

func (p npath) Neighbors(pos gruid.Position) []gruid.Position {
	return p.nb.All(pos, func(npos gruid.Position) bool {
		// strange Neighborer that allows only horizontal moves
		return npos.Y == pos.Y
	})
}

func TestCCBf(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
	pr.ComputeCCAll(nb)
	rg := pr.rg
	pos := gruid.Position{X: rg.Min.X, Y: rg.Min.Y}
	id := pr.CCAt(pos)
	for y := rg.Min.Y + 1; y < rg.Max.Y; y++ {
		pos := gruid.Position{X: rg.Min.X, Y: y}
		nid := pr.CCAt(pos)
		if id == nid {
			t.Errorf("same id on different lines: %d, %d", id, nid)
		}
		if nid != y-rg.Min.Y {
			t.Errorf("bad id: %d, %d", id, y-rg.Min.Y)
		}
		id = nid
	}
	id = pr.CCAt(pos)
	for y := rg.Min.Y; y < rg.Max.Y; y++ {
		pos := gruid.Position{X: rg.Min.X, Y: y}
		id := pr.CCAt(pos)
		for x := rg.Min.X; x < rg.Max.X; x++ {
			if id != pr.CCAt(gruid.Position{X: x, Y: y}) {
				t.Errorf("different id on same line: %d, %d", id, pr.CCAt(gruid.Position{X: x, Y: y}))
			}
		}
	}
	pr.ComputeCC(nb, gruid.Position{X: 1, Y: 1})
	count := 0
	pr.CCIter(func(pos gruid.Position) {
		count++
		if pos.Y != 1 {
			t.Errorf("bad id on line 1: %d", id)
		}
	})
	if count != 10 {
		t.Errorf("bad count: %d", count)
	}
	pr.BreadthFirstMap(nb, []gruid.Position{{X: 2, Y: 0}, {X: 2, Y: 2}}, 3)
	poscosts := []struct {
		pos  gruid.Position
		cost int
	}{
		{gruid.Position{0, 0}, 2},
		{gruid.Position{1, 0}, 1},
		{gruid.Position{2, 0}, 0},
		{gruid.Position{3, 0}, 1},
		{gruid.Position{4, 0}, 2},
		{gruid.Position{5, 0}, 3},
		{gruid.Position{6, 0}, 4},
		{gruid.Position{7, 0}, 4},
		{gruid.Position{0, 2}, 2},
		{gruid.Position{1, 2}, 1},
		{gruid.Position{2, 2}, 0},
		{gruid.Position{3, 2}, 1},
		{gruid.Position{4, 2}, 2},
		{gruid.Position{5, 2}, 3},
		{gruid.Position{6, 2}, 4},
		{gruid.Position{7, 2}, 4},
		{gruid.Position{0, 1}, 4},
		{gruid.Position{1, 1}, 4},
		{gruid.Position{2, 1}, 4},
		{gruid.Position{3, 1}, 4},
		{gruid.Position{4, 1}, 4},
		{gruid.Position{5, 1}, 4},
		{gruid.Position{6, 1}, 4},
	}
	for _, pc := range poscosts {
		if pc.cost != pr.CostAt(pc.pos) {
			t.Errorf("bad cost %d for %+v", pc.cost, pc.pos)
		}
	}
}
