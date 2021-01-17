package gruid

import (
	//"log"
	"math/rand"
	"testing"
)

func randInt(n int) int {
	if n <= 0 {
		return 0
	}
	x := rand.Intn(n)
	return x
}

func TestRange(t *testing.T) {
	rg := NewRange(2, 3, 20, 30)
	max := rg.Size()
	w, h := max.X, max.Y
	count := 0
	rg.Iter(func(p Point) {
		if !p.In(rg) {
			t.Errorf("bad position: %+v", p)
		}
		count++
	})
	if count != w*h {
		t.Errorf("bad count: %d", count)
	}
	rg = rg.Sub(rg.Min)
	max = rg.Size()
	nw, nh := max.X, max.Y
	if nw != w || nh != h {
		t.Errorf("bad size for range %+v", rg)
	}
	if rg.Min.X != 0 || rg.Min.Y != 0 {
		t.Errorf("bad min for range %+v", rg)
	}
	nrg := rg.Shift(1, 2, 3, 4)
	if rg.Min.Shift(1, 2) != nrg.Min {
		t.Errorf("bad min shift for range %+v", nrg)
	}
	if rg.Max.Shift(3, 4) != nrg.Max {
		t.Errorf("bad max shift for range %+v", nrg)
	}
}

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
			p := Point{x, y}
			c := gd.At(p)
			if c.Rune != ' ' {
				t.Errorf("cell: bad content %c at %+v", c.Rune, p)
			}
		}
	}
}

func TestSetCell(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	w, h := max.X, max.Y
	st := Style{}
	gd.Fill(Cell{Rune: '.'})
	for i := 0; i < w*h; i++ {
		p := Point{X: randInt(2*w) - w/2, Y: randInt(2*h) - h/2}
		if gd.Contains(p) {
			c := gd.At(p)
			if c.Rune != '.' && c.Rune != 'x' {
				t.Errorf("Bad fill or setcell %+v at p %+v", c, p)
			}
		}
		gd.Set(p, Cell{Rune: 'x', Style: st.WithFg(2)})
		c := gd.At(p)
		if gd.Contains(p) {
			if c.Rune != 'x' || c.Style.Fg != 2 {
				t.Errorf("Bad content %+v at %+v", c, p)
			}
		} else if c.Rune != 0 || c.Style.Fg != 0 {
			t.Errorf("Bad out of range content: %+v at %+v", c, p)
		}
	}
}

func TestGridSlice(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	slice := gd.Slice(NewRange(5, 5, 10, 10))
	slice.Fill(Cell{Rune: 's'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			if p.In(slice.Bounds()) {
				if c.Rune != 's' {
					t.Errorf("bad slice cell: %c at %+v", c.Rune, p)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad grid non-slice cell: %c at %+v", c.Rune, p)
			}
		}
	}
}

func TestGridSlice2(t *testing.T) {
	gd := NewGrid(80, 24)
	gd.Fill(Cell{Rune: '.'})
	slice := gd.Slice(NewRange(0, 0, 0, 0))
	if !slice.Bounds().Empty() {
		t.Errorf("non empty range %v", slice.Bounds())
	}
	slice = gd.Slice(NewRange(0, 0, -5, -5))
	if !slice.Bounds().Empty() {
		t.Errorf("non empty negative range %v", slice.Bounds())
	}
	slice = gd.Slice(NewRange(5, 5, 0, 0))
	rg := slice.Bounds()
	if rg.Max.X != 5 || rg.Max.Y != 5 || rg.Min.X != 0 || rg.Min.Y != 0 {
		t.Errorf("bad inversed range %+v", slice.Bounds())
	}
}

func TestGridSlice3(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	slice := gd.Slice(gd.Range().Line(1))
	slice.Fill(Cell{Rune: '1'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			if y == 1 {
				if c.Rune != '1' {
					t.Errorf("bad line slice rune: %c", c.Rune)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad outside line slice rune: %c", c.Rune)
			}
		}
	}
	slice = gd.Slice(gd.Range().Column(2))
	gd.Fill(Cell{Rune: '.'})
	slice.Fill(Cell{Rune: '2'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			if x == 2 {
				if c.Rune != '2' {
					t.Errorf("bad column slice rune: %c", c.Rune)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad outside column slice rune: %c", c.Rune)
			}
		}
	}
}

func TestGridSlice4(t *testing.T) {
	gd := NewGrid(10, 10)
	if gd.Slice(NewRange(-5, -5, 20, 20)).Range() != gd.Range() {
		t.Errorf("bad oversized slice")
	}
}

func TestGridSlice5(t *testing.T) {
	gd := NewGrid(80, 24)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	slice := gd.Slice(NewRange(5, 5, 15, 15))
	slice.Fill(Cell{Rune: 's'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			if p.In(slice.Bounds()) {
				if c.Rune != 's' {
					t.Errorf("bad slice cell: %c at %+v", c.Rune, p)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad grid non-slice cell: %c at %+v", c.Rune, p)
			}
		}
	}
}

func TestIterMap(t *testing.T) {
	gd := NewGrid(10, 10)
	gd.Map(func(p Point, c Cell) Cell { return Cell{Rune: '+'} })
	gd.Iter(func(p Point, c Cell) {
		if c.Rune != '+' {
			t.Errorf("bad cell %c at %v", c.Rune, p)
		}
	})
}

func TestCopy(t *testing.T) {
	gd := NewGrid(80, 30)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	gd2 := NewGrid(10, 10)
	gd2.Fill(Cell{Rune: '+'})
	gd.Copy(gd2)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			if p.In(gd2.Range()) {
				if c.Rune != '+' {
					t.Errorf("bad copy at cell: %c at %+v", c.Rune, p)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad grid non-slice cell: %c at %+v", c.Rune, p)
			}
		}
	}
}

func TestCopy2(t *testing.T) {
	gd := NewGrid(80, 10)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Bounds()
	slice := gd.Slice(rg.Lines(1, 3))
	slice2 := gd.Slice(rg.Line(2))
	slice3 := gd.Slice(rg.Lines(2, 4))
	slice.Fill(Cell{Rune: '1'})  // line 1
	slice3.Fill(Cell{Rune: '3'}) // line 3
	slice2.Fill(Cell{Rune: '2'}) // line 2
	slice.Copy(slice3)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			switch {
			case p.In(rg.Line(1)):
				if c.Rune != '2' {
					t.Errorf("bad line 1: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(2)):
				if c.Rune != '3' {
					t.Errorf("bad line 2: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(3)):
				if c.Rune != '3' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			default:
				if c.Rune != '.' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			}
		}
	}
}

func TestCopy3(t *testing.T) {
	gd := NewGrid(80, 10)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Bounds()
	slice := gd.Slice(rg.Lines(1, 3))
	slice2 := gd.Slice(rg.Line(2))
	slice3 := gd.Slice(rg.Lines(2, 4))
	slice.Fill(Cell{Rune: '1'})  // line 1
	slice3.Fill(Cell{Rune: '3'}) // line 3
	slice2.Fill(Cell{Rune: '2'}) // line 2
	slice.Copy(slice2)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			switch {
			case p.In(rg.Line(1)):
				if c.Rune != '2' {
					t.Errorf("bad line 1: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(2)):
				if c.Rune != '2' {
					t.Errorf("bad line 2: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(3)):
				if c.Rune != '3' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			default:
				if c.Rune != '.' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			}
		}
	}
}

func TestCopy4(t *testing.T) {
	gd := NewGrid(80, 10)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Bounds()
	slice := gd.Slice(rg.Lines(1, 3))
	slice2 := gd.Slice(rg.Line(2))
	slice3 := gd.Slice(rg.Lines(2, 4))
	slice.Fill(Cell{Rune: '1'})  // line 1
	slice3.Fill(Cell{Rune: '3'}) // line 3
	slice2.Fill(Cell{Rune: '2'}) // line 2
	slice3.Copy(slice2)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			switch {
			case p.In(rg.Line(1)):
				if c.Rune != '1' {
					t.Errorf("bad line 1: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(2)):
				if c.Rune != '2' {
					t.Errorf("bad line 2: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(3)):
				if c.Rune != '3' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			default:
				if c.Rune != '.' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			}
		}
	}
}

func TestCopy5(t *testing.T) {
	gd := NewGrid(80, 10)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Bounds()
	slice := gd.Slice(rg.Lines(1, 3))
	slice2 := gd.Slice(rg.Line(2))
	slice3 := gd.Slice(rg.Lines(2, 4))
	slice.Fill(Cell{Rune: '1'})  // line 1
	slice3.Fill(Cell{Rune: '3'}) // line 3
	slice2.Fill(Cell{Rune: '2'}) // line 2
	slice2.Copy(slice)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			switch {
			case p.In(rg.Line(1)):
				if c.Rune != '1' {
					t.Errorf("bad line 1: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(2)):
				if c.Rune != '1' {
					t.Errorf("bad line 2: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(3)):
				if c.Rune != '3' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			default:
				if c.Rune != '.' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			}
		}
	}
}

func TestCopy6(t *testing.T) {
	gd := NewGrid(80, 10)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Bounds()
	slice := gd.Slice(rg.Lines(1, 3))
	slice2 := gd.Slice(rg.Line(2))
	slice3 := gd.Slice(rg.Lines(2, 4))
	slice.Fill(Cell{Rune: '1'})  // line 1
	slice3.Fill(Cell{Rune: '3'}) // line 3
	slice2.Fill(Cell{Rune: '2'}) // line 2
	slice2.Copy(slice3)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			switch {
			case p.In(rg.Line(1)):
				if c.Rune != '1' {
					t.Errorf("bad line 1: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(2)):
				if c.Rune != '2' {
					t.Errorf("bad line 2: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(3)):
				if c.Rune != '3' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			default:
				if c.Rune != '.' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			}
		}
	}
}

func TestCopy7(t *testing.T) {
	gd := NewGrid(80, 10)
	max := gd.Size()
	w, h := max.X, max.Y
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Bounds()
	slice := gd.Slice(rg.Lines(1, 3))
	slice2 := gd.Slice(rg.Line(2))
	slice3 := gd.Slice(rg.Lines(2, 4))
	slice.Fill(Cell{Rune: '1'})  // line 1
	slice3.Fill(Cell{Rune: '3'}) // line 3
	slice2.Fill(Cell{Rune: '2'}) // line 2
	slice3.Copy(slice)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := Point{x, y}
			c := gd.At(p)
			switch {
			case p.In(rg.Line(1)):
				if c.Rune != '1' {
					t.Errorf("bad line 1: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(2)):
				if c.Rune != '1' {
					t.Errorf("bad line 2: %c at %+v", c.Rune, p)
				}
			case p.In(rg.Line(3)):
				if c.Rune != '2' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			default:
				if c.Rune != '.' {
					t.Errorf("bad line 3: %c at %+v", c.Rune, p)
				}
			}
		}
	}
}

func TestCopy8(t *testing.T) {
	gd := NewGrid(3, 10)
	gd.Fill(Cell{Rune: '.'})
	gd2 := NewGrid(3, 10)
	gd2.Fill(Cell{Rune: '+'})
	gd.Copy(gd2)
	gd.Iter(func(p Point, c Cell) {
		if c.Rune != '+' {
			t.Errorf("bad rune %c at %v", c.Rune, p)
		}
	})
}

func TestCopy9(t *testing.T) {
	gd := NewGrid(80, 10)
	if gd.Copy(gd) != gd.Range().Size() {
		t.Errorf("bad same range copy")
	}
}

func TestResize(t *testing.T) {
	gd := NewGrid(20, 10)
	gd.Fill(Cell{Rune: '.'})
	rg := gd.Range()
	gd = gd.Resize(30, 20)
	if gd.Size().X != 30 || gd.Size().Y != 20 {
		t.Errorf("bad size: %v", gd.Size())
	}
	gd.Iter(func(p Point, c Cell) {
		if p.In(rg) {
			if c.Rune != '.' {
				t.Error("bad preservation of content")
			}
		} else if c.Rune != ' ' {
			t.Error("bad new content")
		}
	})
}

func TestIterator(t *testing.T) {
	gd := NewGrid(10, 10)
	slice := gd.Slice(NewRange(2, 2, 5, 5))
	it := slice.Iterator()
	for it.Next() {
		//log.Printf("pos: %v, i: %d, c: %c", it.P(), it.i, it.Cell())
		if it.Cell().Rune != ' ' {
			t.Errorf("not space rune: %c", it.Cell().Rune)
		}
		it.SetCell(it.Cell().WithRune('x'))
		if it.Cell().Rune != 'x' {
			t.Errorf("not x rune: %c", it.Cell().Rune)
		}
		if slice.At(it.P()).Rune != 'x' {
			t.Errorf("not x rune at %v: %c", it.P(), slice.At(it.P()).Rune)
		}
	}
	gd.Iter(func(p Point, c Cell) {
		if p.In(slice.Bounds()) {
			if c.Rune != 'x' {
				t.Errorf("bad rune at %v: %c", p, c.Rune)
			}
		} else if c.Rune != ' ' {
			t.Errorf("not space rune at %v: %c", p, c.Rune)
		}

	})
	it.SetP(Point{1, 1})
	if it.P().X != 1 || it.P().Y != 1 {
		t.Errorf("bad SetP: %v", it.P())
	}
	it.SetCell(it.Cell().WithRune('z'))
	if slice.At(Point{1, 1}).Rune != 'z' {
		t.Errorf("not z: %c", slice.At(Point{1, 1}).Rune)
	}
	it.Reset()
	for it.Next() {
		it.SetCell(it.Cell().WithRune('y'))
	}
}

func TestCell(t *testing.T) {
	c := Cell{}
	if c.WithRune('x').Rune != 'x' {
		t.Errorf("bad rune: %c", c.WithRune('x').Rune)
	}
	st := Style{Fg: 1, Bg: 2, Attrs: 3}
	if c.WithStyle(st).Style != st {
		t.Errorf("bad style: %+v", st)
	}
	if st.WithFg(4).Fg != Color(4) {
		t.Errorf("bad fg: %+v", st.WithFg(4).Fg)
	}
	if st.WithBg(4).Bg != Color(4) {
		t.Errorf("bad fg: %+v", st.WithBg(4).Bg)
	}
	if st.WithAttrs(4).Attrs != AttrMask(4) {
		t.Errorf("bad fg: %+v", st.WithBg(4).Bg)
	}
}

func TestPoint(t *testing.T) {
	p := Point{2, 3}
	if p.Mul(3).X != 6 {
		t.Errorf("bad mul: %v", p.Mul(3))
	}
	if p.Div(2).X != 1 {
		t.Errorf("bad mul: %v", p.Div(2))
	}
}

func TestRangeShift(t *testing.T) {
	rg := NewRange(1, 2, 3, 4)
	nrg := NewRange(2, 3, 4, 5)
	if rg.Shift(1, 1, 1, 1) != nrg {
		t.Errorf("bad shift: %v", rg.Shift(1, 1, 1, 1))
	}
	empty := Range{}
	if rg.Shift(0, 0, -5, 0) != empty {
		t.Errorf("bad shift: %v", rg.Shift(0, 0, -5, 0))
	}
	if rg.Shift(0, 0, 0, -5) != empty {
		t.Errorf("bad shift: %v", rg.Shift(0, 0, 0, -5))
	}
	if rg.Add(Point{1, 1}) != nrg {
		t.Errorf("bad add: %v", rg.Add(Point{1, 1}))
	}
}

func TestRangeColumnsLines(t *testing.T) {
	rg := NewRange(1, 1, 30, 30)
	if rg.Columns(4, 10).Size().X != 6 {
		t.Errorf("bad number of columns for range %v", rg.Columns(4, 10).Size().Y)
	}
	if rg.Columns(4, 10).Min.X != 5 {
		t.Errorf("bad min.X for range %v", rg.Columns(4, 10).Min.X)
	}
	if rg.Lines(4, 10).Size().Y != 6 {
		t.Errorf("bad number of columns for range %v", rg.Columns(4, 10).Size().Y)
	}
	if rg.Lines(4, 10).Min.Y != 5 {
		t.Errorf("bad min.X for range %v", rg.Columns(4, 10).Min.Y)
	}
	if !rg.Column(200).Empty() {
		t.Errorf("not empty column")
	}
	if !rg.Line(200).Empty() {
		t.Errorf("not empty line")
	}
}

func TestRangeEq(t *testing.T) {
	rg := NewRange(1, 2, 3, 4)
	if !rg.Eq(rg) {
		t.Errorf("bad reflexive Eq for %v", rg)
	}
	if rg.Eq(rg.Shift(1, 0, 0, 0)) {
		t.Errorf("bad shift Eq for %v", rg)
	}
	erg := Range{Point{2, 3}, Point{-1, -4}}
	empty := Range{}
	if !erg.Eq(empty) {
		t.Errorf("bad empty range equality")
	}
}

func TestRangeUnion(t *testing.T) {
	rg := NewRange(1, 2, 3, 4)
	org := NewRange(11, 12, 13, 14)
	union := NewRange(1, 2, 13, 14)
	if rg.Union(org) != union {
		t.Errorf("bad Union")
	}
	if !rg.In(union) {
		t.Errorf("bad In")
	}
}

func TestBounds(t *testing.T) {
	gd := NewGrid(10, 10)
	slice := gd.Slice(NewRange(2, 2, 4, 4))
	if slice.Bounds() != NewRange(2, 2, 4, 4) {
		t.Errorf("bad Bounds %v", slice.Bounds())
	}
}

func BenchmarkGridRangeIterSet(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Range().Iter(func(p Point) {
			gd.Set(p, Cell{}.WithRune('x'))
		})
	}
}

func BenchmarkGridLoopSet(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		max := gd.Size()
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := Point{x, y}
				gd.Set(p, Cell{}.WithRune('x'))
			}
		}
	}
}

func BenchmarkGridIteratorSet(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		it := gd.Iterator()
		for it.Next() {
			it.SetCell(Cell{}.WithRune('x'))
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

func BenchmarkGridCopy(b *testing.B) {
	gd := NewGrid(80, 24)
	gd2 := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Copy(gd2)
	}
}

func BenchmarkGridFill(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell{}.WithRune('x'))
	}
}

func BenchmarkGridMap(b *testing.B) {
	gd := NewGrid(80, 24)
	for i := 0; i < b.N; i++ {
		gd.Map(func(p Point, c Cell) Cell { return Cell{}.WithRune('x') })
	}
}

func BenchmarkGridFillVertical(b *testing.B) {
	gd := NewGrid(1, 24*80)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell{}.WithRune('x'))
	}
}

func BenchmarkGridFillVertical8(b *testing.B) {
	gd := NewGrid(8, 24*10)
	for i := 0; i < b.N; i++ {
		gd.Fill(Cell{}.WithRune('x'))
	}
}
