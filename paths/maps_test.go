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

func BenchmarkDijktraMapSmall(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	nb := bpath{&Neighbors{}}
	for i := 0; i < b.N; i++ {
		pr.DijkstraMap(nb, []gruid.Point{{X: 2, Y: 2}}, 9)
	}
}

func BenchmarkDijktraMapBig(b *testing.B) {
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
