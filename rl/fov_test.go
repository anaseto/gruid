package rl

import (
	"testing"

	"github.com/anaseto/gruid"
)

const maxLOS = 10

func TestFOV(t *testing.T) {
	fov := NewFOV(gruid.NewRange(-maxLOS, -maxLOS, maxLOS+1, maxLOS+1))
	lt := &lighter{max: maxLOS}
	lns := fov.VisionMap(lt, gruid.Point{0, 0})
	if len(lns) != (2*maxLOS+1)*(2*maxLOS+1) {
		t.Errorf("bad length: %d vs %d", len(lns), (2*maxLOS+1)*(2*maxLOS+1))
	}
	ray := fov.Ray(lt, gruid.Point{5, 0})
	if len(ray) != 6 {
		t.Errorf("bad ray length: %d", len(ray))
	}
	if ray[2].P.X != 2 {
		t.Errorf("bad position in ray: %d", ray[2].P.X)
	}
	lt.max = 4
	lns = fov.LightMap(lt, []gruid.Point{{-5, 0}, {5, 0}})
	if len(lns) != 2*(2*4+1)*(2*4+1) {
		t.Errorf("bad length: %d vs %d", len(lns), 2*(2*4+1)*(2*4+1))
	}
}

type lighter struct {
	max int
}

func (lt *lighter) Cost(src, from, to gruid.Point) int {
	if src == from {
		return 0
	}
	if lt.diagonalStep(from, to) {
		return 2
	}
	return 1
}

func (lt *lighter) MaxCost(src gruid.Point) int {
	return lt.max
}

func (lt *lighter) diagonalStep(from, to gruid.Point) bool {
	step := to.Sub(from)
	return step.X != 0 && step.Y != 0
}

func BenchmarkFOV(b *testing.B) {
	fov := NewFOV(gruid.NewRange(-maxLOS, -maxLOS, maxLOS+1, maxLOS+1))
	lt := &lighter{max: maxLOS}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{0, 0})
	}
}

func BenchmarkFOVBig(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 80, 24))
	lt := &lighter{max: maxLOS}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{20, 10})
	}
}

func BenchmarkFOVBigLights(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 80, 24))
	lt := &lighter{max: 7}
	for i := 0; i < b.N; i++ {
		fov.LightMap(lt, []gruid.Point{{20, 10}, {40, 10}, {70, 15}})
	}
}

func BenchmarkFOVBigBig(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 80, 24))
	lt := &lighter{max: 50}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{40, 10})
	}
}

func BenchmarkFOV20x20(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 20, 20))
	lt := &lighter{max: 10}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{10, 10})
	}
}

func BenchmarkFOV100x100(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 100, 100))
	lt := &lighter{max: 20}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{50, 50})
	}
}

//func BenchmarkFOV600x600(b *testing.B) {
//fov := NewFOV(gruid.NewRange(0, 0, 600, 600))
//lt := &lighter{max: 50}
//for i := 0; i < b.N; i++ {
//fov.VisionMap(lt, gruid.Point{200, 200})
//}
//}
