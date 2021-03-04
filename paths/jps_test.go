package paths

import (
	//"fmt"
	"math/rand"
	"testing"

	"github.com/anaseto/gruid"
)

func passable1(p gruid.Point) bool {
	if p.X == 20 {
		return p.Y == 0
	}
	if p.X == 40 {
		return p.Y == 23
	}
	if p.X == 60 {
		return p.Y == 0
	}
	return true
}

func passable2(p gruid.Point) bool {
	if p.X == 1 {
		return p.Y == 0 || p.Y == 11 || p.Y == 23
	}
	if p.X <= 20 {
		return p.Y%3 != 0 || p.X%3 != 1
	}
	if p.X <= 35 && p.X >= 25 {
		return p.Y%2 != 1
	}
	if p.X == 40 {
		return p.Y == 23 || p.Y == 15
	}
	if p.X == 50 {
		return p.Y%2 != 0
	}
	if p.X == 60 {
		return p.Y%3 != 1
	}
	if p.X <= 70 && p.X > 65 {
		return (p.Y+p.X)%3 != 1
	}
	if p.X == 78 {
		return p.Y == 0 || p.Y == 11 || p.Y == 23
	}
	return true
}

func TestJPS(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	pr.JPSPath(path, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20}, func(gruid.Point) bool { return true }, true)
}

func TestPassable1Diags(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	path = pr.JPSPath(path, gruid.Point{X: 0, Y: 23}, gruid.Point{X: 79, Y: 23}, passable1, true)
	ap := apath{nb: &Neighbors{}, passable: passable1, diags: true}
	patha := pr.AstarPath(ap, gruid.Point{X: 0, Y: 23}, gruid.Point{X: 79, Y: 23})
	if len(path) != len(patha) {
		t.Errorf("bad path:\n%v\n%v", path, patha)
	}
	//fmt.Printf("%s\n\n", logrid)
}

func TestPassable1(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	path = pr.JPSPath(path, gruid.Point{X: 0, Y: 23}, gruid.Point{X: 79, Y: 23}, passable1, false)
	ap := apath{nb: &Neighbors{}, passable: passable1, diags: false}
	patha := pr.AstarPath(ap, gruid.Point{X: 0, Y: 23}, gruid.Point{X: 79, Y: 23})
	if len(path) != len(patha) {
		t.Errorf("bad path:\n%v\n%v", path, patha)
	}
	//fmt.Printf("%s\n\n", logrid)
}

func TestPassableRand(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	for i := 0; i < 1000; i++ {
		path := []gruid.Point{}
		from := gruid.Point{rand.Intn(80), rand.Intn(24)}
		to := gruid.Point{rand.Intn(80), rand.Intn(24)}
		path = pr.JPSPath(path, from, to, passable1, false)
		ap := apath{nb: &Neighbors{}, passable: passable1, diags: false}
		patha := pr.AstarPath(ap, from, to)
		if len(path) != len(patha) {
			t.Errorf("bad path:\n%v\n%v", path, patha)
		}
	}
	//fmt.Printf("%s\n\n", logrid)
}

func TestPassableRandDiags(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	for i := 0; i < 1000; i++ {
		path := []gruid.Point{}
		from := gruid.Point{rand.Intn(80), rand.Intn(24)}
		to := gruid.Point{rand.Intn(80), rand.Intn(24)}
		path = pr.JPSPath(path, from, to, passable1, true)
		ap := apath{nb: &Neighbors{}, passable: passable1, diags: true}
		patha := pr.AstarPath(ap, from, to)
		if len(path) != len(patha) {
			t.Errorf("bad path:\n%v\n%v", path, patha)
		}
	}
	//fmt.Printf("%s\n\n", logrid)
}

func TestPassable2Rand(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	for i := 0; i < 1000; i++ {
		path := []gruid.Point{}
		from := gruid.Point{rand.Intn(80), rand.Intn(24)}
		to := gruid.Point{rand.Intn(80), rand.Intn(24)}
		path = pr.JPSPath(path, from, to, passable2, false)
		ap := apath{nb: &Neighbors{}, passable: passable2, diags: false}
		patha := pr.AstarPath(ap, from, to)
		if len(path) != len(patha) {
			t.Errorf("bad path:\n%v\n%v", path, patha)
		}
	}
	//fmt.Printf("%s\n\n", logrid)
}

func TestPassable2RandDiags(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	for i := 0; i < 1000; i++ {
		path := []gruid.Point{}
		from := gruid.Point{rand.Intn(80), rand.Intn(24)}
		to := gruid.Point{rand.Intn(80), rand.Intn(24)}
		path = pr.JPSPath(path, from, to, passable2, true)
		ap := apath{nb: &Neighbors{}, passable: passable2, diags: true}
		patha := pr.AstarPath(ap, from, to)
		if len(path) != len(patha) {
			t.Errorf("bad path:\n%v\n%v", path, patha)
		}
	}
	//fmt.Printf("%s\n\n", logrid)
}

func TestPassableBorder(t *testing.T) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	path = pr.JPSPath(path, gruid.Point{X: 4, Y: 4}, gruid.Point{X: 4, Y: 0}, func(gruid.Point) bool { return true }, false)
	ap := apath{nb: &Neighbors{}, passable: func(gruid.Point) bool { return true }, diags: false}
	patha := pr.AstarPath(ap, gruid.Point{X: 4, Y: 4}, gruid.Point{X: 4, Y: 0})
	if len(path) != len(patha) {
		t.Errorf("bad path:\n%v\n%v", path, patha)
	}
	//fmt.Printf("%s\n\n", logrid)
}

func BenchmarkJPS(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	for i := 0; i < b.N; i++ {
		path = pr.JPSPath(path, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20}, func(gruid.Point) bool { return true }, true)
	}
}

func BenchmarkJPSShortPath(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	for i := 0; i < b.N; i++ {
		path = pr.JPSPath(path, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 15, Y: 10}, func(gruid.Point) bool { return true }, true)
	}
}

func BenchmarkJPSPassable1(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	for i := 0; i < b.N; i++ {
		path = pr.JPSPath(path, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20}, passable1, true)
	}
}

func BenchmarkJPSPassable1NoDiags(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	for i := 0; i < b.N; i++ {
		path = pr.JPSPath(path, gruid.Point{X: 2, Y: 2}, gruid.Point{X: 70, Y: 20}, passable1, false)
	}
}

func BenchmarkPassable2RandJPS(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	path := []gruid.Point{}
	for i := 0; i < b.N; i++ {
		from := gruid.Point{rand.Intn(80), rand.Intn(24)}
		to := gruid.Point{rand.Intn(80), rand.Intn(24)}
		path = pr.JPSPath(path, from, to, passable2, false)
	}
	//fmt.Printf("%s\n\n", logrid)
}

func BenchmarkPassable2RandAstar(b *testing.B) {
	pr := NewPathRange(gruid.NewRange(0, 0, 80, 24))
	for i := 0; i < b.N; i++ {
		from := gruid.Point{rand.Intn(80), rand.Intn(24)}
		to := gruid.Point{rand.Intn(80), rand.Intn(24)}
		ap := apath{nb: &Neighbors{}, passable: passable2, diags: false}
		pr.AstarPath(ap, from, to)
	}
	//fmt.Printf("%s\n\n", logrid)
}
