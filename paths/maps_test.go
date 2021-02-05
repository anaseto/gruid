package paths

import (
	"testing"

	"github.com/anaseto/gruid"
)

func TestPathMaps(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 10, 5))
	nb := npath{}
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
	pr.BreadthFirstMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 3)
	pr.DijkstraMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 6)
	for _, pc := range poscosts {
		if pc.cost != pr.BreadthFirstMapAt(pc.p) {
			t.Errorf("bad bf cost %d for %+v", pr.BreadthFirstMapAt(pc.p), pc.p)
		}
		if 2*pc.cost != pr.DijkstraMapAt(pc.p) && pc.cost < 4 || pc.cost >= 4 && pr.DijkstraMapAt(pc.p) != 7 {
			t.Errorf("bad dijkstra cost %d for %+v", pr.DijkstraMapAt(pc.p), pc.p)
		}
	}
	for _, n := range pr.DijkstraMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 9) {
		for _, pc := range poscosts {
			if pc.p == n.P && 2*pc.cost != n.Cost {
				t.Errorf("bad dijkstra cost %d for %+v", n.Cost, n.P)
			}
		}
	}
	for _, n := range pr.BreadthFirstMap(nb, []gruid.Point{{X: 2, Y: 0}, {X: 2, Y: 2}}, 4) {
		for _, pc := range poscosts {
			if pc.p == n.P && pc.cost != n.Cost {
				t.Errorf("bad bf cost %d for %+v", n.Cost, n.P)
			}
		}
	}
}

func BenchmarkDijkstraMapSmall(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.DijkstraMap(nb, []gruid.Point{{X: 2, Y: 2}}, 9)
	}
}

func BenchmarkDijkstraMapBig(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.DijkstraMap(nb, []gruid.Point{{X: 2, Y: 2}}, 80)
	}
}

func BenchmarkBfMapSmall(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.BreadthFirstMap(nb, []gruid.Point{{X: 2, Y: 2}}, 9)
	}
}

func BenchmarkBfMapBig(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.BreadthFirstMap(nb, []gruid.Point{{X: 2, Y: 2}}, 80)
	}
}

func BenchmarkAstar(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.AstarPath(nb, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20})
	}
}

func BenchmarkAstarShortPath(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.AstarPath(nb, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 15, Y: 10})
	}
}

type apath struct {
	nb       *Neighbors
	passable func(gruid.Point) bool
	diags    bool
}

func (ap apath) Neighbors(p gruid.Point) []gruid.Point {
	if ap.diags {
		return ap.nb.All(p, func(q gruid.Point) bool {
			return ap.passable(q)
		})
	}
	return ap.nb.Cardinal(p, func(q gruid.Point) bool {
		return ap.passable(q)
	})
}

func (ap apath) Cost(p, q gruid.Point) int {
	return 1
}

func (ap apath) Estimation(p, q gruid.Point) int {
	p = p.Sub(q)
	return max(abs(p.X), abs(p.Y))
	//return abs(p.X) + abs(p.Y)
}

func BenchmarkAstarPassable1(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	ap := apath{nb: &Neighbors{}, passable: passable1, diags: true}
	for i := 0; i < b.N; i++ {
		pr.AstarPath(ap, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20})
	}
}

func BenchmarkAstarPassable1NoDiags(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	ap := apath{nb: &Neighbors{}, passable: passable1}
	for i := 0; i < b.N; i++ {
		pr.AstarPath(ap, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20})
	}
}
