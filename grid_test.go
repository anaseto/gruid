package gruid

import (
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
	max = gd.rg.Size()
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
			if p.In(slice.rg) {
				if c.Rune != 's' {
					t.Errorf("bad slice cell: %c at %+v", c.Rune, p)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad grid non-slice cell: %c at %+v", c.Rune, p)
			}
		}
	}
	slice = gd.Slice(NewRange(0, 0, 0, 0))
	if !slice.rg.Empty() {
		t.Errorf("non empty range %v", slice.rg)
	}
	slice = gd.Slice(NewRange(0, 0, -5, -5))
	if !slice.rg.Empty() {
		t.Errorf("non empty negative range %v", slice.rg)
	}
	slice = gd.Slice(NewRange(5, 5, 0, 0))
	rg := slice.rg
	if rg.Max.X != 5 || rg.Max.Y != 5 || rg.Min.X != 0 || rg.Min.Y != 0 {
		t.Errorf("bad inversed range %+v", slice.rg)
	}
	slice = gd.Slice(gd.Range().Line(1))
	gd.Fill(Cell{Rune: '.'})
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
	rg := gd.rg
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
	rg := gd.rg
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
	rg := gd.rg
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
	rg := gd.rg
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
	rg := gd.rg
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
	rg := gd.rg
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
