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
	w, h := rg.Size()
	count := 0
	rg.Iter(func(pos Position) {
		if !pos.In(rg) {
			t.Errorf("bad position: %+v", pos)
		}
		count++
	})
	if count != w*h {
		t.Errorf("bad count: %d", count)
	}
	rg = rg.Origin()
	nw, nh := rg.Size()
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
	w, h := gd.Size()
	if w != 80 && h != 24 {
		t.Errorf("bad default size: (%d,%d)", w, h)
	}
	gd = NewGrid(50, 50)
	w, h = gd.Size()
	if w != 50 && h != 50 {
		t.Errorf("grid size does not match configuration: (%d,%d)", w, h)
	}
	rw, rh := gd.Range().Size()
	if w != rw || rh != h {
		t.Errorf("incompatible sizes: grid (%d,%d) range (%d,%d)", w, h, rw, rh)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pos := Position{x, y}
			c := gd.GetCell(pos)
			if c.Rune != ' ' {
				t.Errorf("cell: bad content %c at %+v", c.Rune, pos)
			}
		}
	}
}

func TestSetCell(t *testing.T) {
	gd := NewGrid(80, 24)
	w, h := gd.Size()
	st := Style{}
	gd.Fill(Cell{Rune: '.'})
	for i := 0; i < w*h; i++ {
		pos := Position{X: randInt(2*w) - w/2, Y: randInt(2*h) - h/2}
		if gd.Contains(pos) {
			c := gd.GetCell(pos)
			if c.Rune != '.' && c.Rune != 'x' {
				t.Errorf("Bad fill or setcell %+v at pos %+v", c, pos)
			}
		}
		gd.SetCell(pos, Cell{Rune: 'x', Style: st.WithFg(2)})
		c := gd.GetCell(pos)
		if gd.Contains(pos) {
			if c.Rune != 'x' || c.Style.Fg != 2 {
				t.Errorf("Bad content %+v at %+v", c, pos)
			}
		} else if c.Rune != 0 || c.Style.Fg != 0 {
			t.Errorf("Bad out of range content: %+v at %+v", c, pos)
		}
	}
}

func TestGridSlice(t *testing.T) {
	gd := NewGrid(80, 24)
	w, h := gd.Size()
	gd.Fill(Cell{Rune: '.'})
	slice := gd.Slice(NewRange(5, 5, 10, 10))
	slice.Fill(Cell{Rune: 's'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pos := Position{x, y}
			c := gd.GetCell(pos)
			if pos.In(slice.Range()) {
				if c.Rune != 's' {
					t.Errorf("bad slice cell: %c at %+v", c.Rune, pos)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad grid non-slice cell: %c at %+v", c.Rune, pos)
			}
		}
	}
	slice = gd.Slice(NewRange(0, 0, 0, 0))
	if !slice.Range().Empty() {
		t.Errorf("non empty range %v", slice.rg)
	}
	slice = gd.Slice(NewRange(0, 0, -5, -5))
	if !slice.Range().Empty() {
		t.Errorf("non empty negative range %v", slice.rg)
	}
	slice = gd.Slice(NewRange(5, 5, 0, 0))
	rg := slice.Range()
	if rg.Max.X != 5 || rg.Max.Y != 5 || rg.Min.X != 0 || rg.Min.Y != 0 {
		t.Errorf("bad inversed range %+v", slice.rg)
	}
	slice = gd.Slice(gd.Range().Origin().Line(1))
	gd.Fill(Cell{Rune: '.'})
	slice.Fill(Cell{Rune: '1'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pos := Position{x, y}
			c := gd.GetCell(pos)
			if y == 1 {
				if c.Rune != '1' {
					t.Errorf("bad line slice rune: %c", c.Rune)
				}
			} else if c.Rune != '.' {
				t.Errorf("bad outside line slice rune: %c", c.Rune)
			}
		}
	}
	slice = gd.Slice(gd.Range().Origin().Column(2))
	gd.Fill(Cell{Rune: '.'})
	slice.Fill(Cell{Rune: '2'})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pos := Position{x, y}
			c := gd.GetCell(pos)
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
