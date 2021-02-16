package rl

import (
	"math/rand"
	"testing"

	"github.com/anaseto/gruid"
)

func TestNewGrid(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	if max.X != 80 && max.Y != 24 {
		t.Errorf("bad default size: (%d,%d)", max.X, max.Y)
	}
	gd = NewGrid(50, 50)
	max = gd.Size()
	w, h := max.X, max.Y
	if w != 50 && h != 50 {
		t.Errorf("grid size does not match configuration: (%d,%d)", w, h)
	}
	max = gd.Bounds().Size()
	rw, rh := max.X, max.Y
	if w != rw || rh != h {
		t.Errorf("incompatible sizes: grid (%d,%d) range (%d,%d)", w, h, rw, rh)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := gruid.Point{x, y}
			c := gd.At(p)
			if c != Cell(0) {
				t.Errorf("cell: bad content %d at %+v", c, p)
			}
		}
	}
}

func TestSetCell(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell(1))
	for i := 0; i < w*h; i++ {
		p := gruid.Point{X: rand.Intn(2*w) - w/2, Y: rand.Intn(2*h) - h/2}
		if gd.Contains(p) {
			c := gd.At(p)
			if c != Cell(1) && c != Cell(2) {
				t.Errorf("Bad fill or setcell %+v at p %+v", c, p)
			}
		}
		gd.Set(p, Cell(2))
		c := gd.At(p)
		if gd.Contains(p) {
			if c != Cell(2) {
				t.Errorf("Bad content %+v at %+v", c, p)
			}
		} else if c != Cell(0) {
			t.Errorf("Bad out of range content: %+v at %+v", c, p)
		}
	}
}

func TestGridSlice(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell(1))
	slice := gd.Slice(gruid.NewRange(5, 5, 10, 10))
	slice.Fill(Cell(2))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := gruid.Point{x, y}
			c := gd.At(p)
			if p.In(slice.Bounds()) {
				if c != Cell(2) {
					t.Errorf("bad slice cell: %d at %+v", c, p)
				}
			} else if c != Cell(1) {
				t.Errorf("bad grid non-slice cell: %d at %+v", c, p)
			}
		}
	}
}

func TestGridSlice2(t *testing.T) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	slice := gd.Slice(gruid.NewRange(0, 0, 0, 0))
	if !slice.Bounds().Empty() {
		t.Errorf("non empty range %v", slice.Bounds())
	}
	slice = gd.Slice(gruid.NewRange(0, 0, -5, -5))
	if !slice.Bounds().Empty() {
		t.Errorf("non empty negative range %v", slice.Bounds())
	}
	slice = gd.Slice(gruid.NewRange(5, 5, 0, 0))
	rg := slice.Bounds()
	if rg.Max.X != 5 || rg.Max.Y != 5 || rg.Min.X != 0 || rg.Min.Y != 0 {
		t.Errorf("bad inversed range %+v", slice.Bounds())
	}
}

func TestGridSlice4(t *testing.T) {
	gd := NewGrid(10, 10)
	if gd.Slice(gruid.NewRange(-5, -5, 20, 20)).Range() != gd.Range() {
		t.Errorf("bad oversized slice")
	}
}

func TestCopy(t *testing.T) {
	gd := NewGrid(80, 30)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell(1))
	gd2 := NewGrid(10, 10)
	gd2.Fill(Cell(2))
	gd.Copy(gd2)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := gruid.Point{x, y}
			c := gd.At(p)
			if p.In(gd2.Range()) {
				if c != Cell(2) {
					t.Errorf("bad copy at cell: %d at %+v", c, p)
				}
			} else if c != Cell(1) {
				t.Errorf("bad grid non-slice cell: %d at %+v", c, p)
			}
		}
	}
}

func TestCopy8(t *testing.T) {
	gd := NewGrid(80, 10)
	if gd.Copy(gd) != gd.Range().Size() {
		t.Errorf("bad same range copy")
	}
}

func TestMap(t *testing.T) {
	gd := NewGrid(8, 8)
	gd.Map(func(p gruid.Point, c Cell) Cell {
		return Cell(1)
	})
	if gd.Count(Cell(1)) != 8*8 {
		t.Errorf("bad map")
	}
}

func TestCount(t *testing.T) {
	gd := NewGrid(80, 10)
	if gd.Count(Cell(0)) != 800 {
		t.Errorf("bad count")
	}
	gd.Slice(gd.Range().Line(0)).Fill(Cell(1))
	if gd.Count(Cell(1)) != 80 {
		t.Errorf("bad count")
	}
	if gd.Count(Cell(1)) != gd.Slice(gd.Range().Line(0)).Count(Cell(1)) {
		t.Errorf("bad count")
	}
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			gd := NewGrid(x, y)
			gd = gd.Slice(gruid.NewRange(1, 1, 10, 10))
			gd.FillFunc(func() Cell {
				return Cell(rand.Intn(15))
			})
			for i := 0; i < 15; i++ {
				if gd.Count(Cell(i)) != gd.CountFunc(func(c Cell) bool { return c == Cell(i) }) {
					t.Errorf("bad count")
				}
			}
		}
	}
}

func BenchmarkGridCount(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		gd.Count(Cell(1))
	}
}

func BenchmarkGridCountFunc(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		gd.CountFunc(func(c Cell) bool { return c == Cell(1) })
	}
}

func BenchmarkGridIter(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		n := 0
		gd.Iter(func(p gruid.Point, c Cell) {
			if c == Cell(1) {
				n++
			}
		})
	}
}

func BenchmarkGridIterator(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		n := 0
		it := gd.Iterator()
		for it.Next() {
			if it.Cell() == Cell(1) {
				n++
			}
		}
	}
}

func BenchmarkGridLoopAt(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		n := Cell(0)
		max := gd.Size()
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := gruid.Point{x, y}
				n += gd.At(p)
			}
		}
	}
}

func BenchmarkGridLoopAtU(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		n := Cell(0)
		max := gd.Size()
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := gruid.Point{x, y}
				n += gd.AtU(p)
			}
		}
	}
}

func BenchmarkGridIterSet(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		gd.Iter(func(p gruid.Point, c Cell) {
			gd.Set(p, Cell(2))
		})
	}
}

func BenchmarkGridRangeIterSet(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		gd.Range().Iter(func(p gruid.Point) {
			gd.Set(p, Cell(2))
		})
	}
}

func BenchmarkGridLoopSet(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		max := gd.Size()
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := gruid.Point{x, y}
				gd.Set(p, Cell(2))
			}
		}
	}
}

func BenchmarkGridIteratorSet(b *testing.B) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell(1))
	for i := 0; i < b.N; i++ {
		it := gd.Iterator()
		for it.Next() {
			it.SetCell(Cell(2))
		}
	}
}

func BenchmarkGridIteratorNew(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		it := gd.Iterator()
		it.Next()
	}
}

func BenchmarkGridFill(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell(1))
	}
}

func BenchmarkGridMap(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Map(func(p gruid.Point, c Cell) Cell { return Cell(1) })
	}
}

func BenchmarkGridFillFunc(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.FillFunc(func() Cell { return Cell(1) })
	}
}

func BenchmarkGridCopy(b *testing.B) {
	gd := NewGrid(80, 24)
	gd2 := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Copy(gd2)
	}
}

func BenchmarkGridVerticalCopy(b *testing.B) {
	gd := NewGrid(2, 40*24)
	gd2 := NewGrid(2, 40*24)
	for i := 0; i < b.N; i++ {
		gd.Copy(gd2)
	}
}

func BenchmarkGridFillVertical(b *testing.B) {
	gd := NewGrid(1, 24*80)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell(1))
	}
}

func BenchmarkGridFillVertical8(b *testing.B) {
	gd := NewGrid(8, 24*10)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell(1))
	}
}

func BenchmarkGridFillVertical16(b *testing.B) {
	gd := NewGrid(16, 12*10)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell(1))
	}
}
