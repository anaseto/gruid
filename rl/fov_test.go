package rl

import (
	"testing"

	"github.com/anaseto/gruid"
)

const maxLOS = 10

func TestFOV(t *testing.T) {
	fov := NewFOV(gruid.NewRange(-maxLOS, -maxLOS, maxLOS+1, maxLOS+1))
	lt := &lighter{}
	lns := fov.VisionMap(lt, gruid.Point{0, 0}, maxLOS)
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
	lns = fov.LightMap(lt, []gruid.Point{{-5, 0}, {5, 0}}, 4)
	if len(lns) != 2*(2*4+1)*(2*4+1) {
		t.Errorf("bad length: %d vs %d", len(lns), 2*(2*4+1)*(2*4+1))
	}
}

type lighter struct {
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

func (lt *lighter) diagonalStep(from, to gruid.Point) bool {
	step := to.Sub(from)
	return step.X != 0 && step.Y != 0
}

func BenchmarkFOV(b *testing.B) {
	fov := NewFOV(gruid.NewRange(-maxLOS, -maxLOS, maxLOS+1, maxLOS+1))
	lt := &lighter{}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{0, 0}, maxLOS)
	}
}

func BenchmarkFOVBig(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 80, 24))
	lt := &lighter{}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{20, 10}, maxLOS)
	}
}

func BenchmarkFOVBigLights(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 80, 24))
	lt := &lighter{}
	for i := 0; i < b.N; i++ {
		fov.LightMap(lt, []gruid.Point{{20, 10}, {40, 10}, {70, 15}}, 7)
	}
}

func BenchmarkFOVBigBig(b *testing.B) {
	fov := NewFOV(gruid.NewRange(0, 0, 80, 24))
	lt := &lighter{}
	for i := 0; i < b.N; i++ {
		fov.VisionMap(lt, gruid.Point{40, 10}, 50)
	}
}