package rl

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/anaseto/gruid"
)

// Grid is modeled after gruid.Grid but with int Cells. It is suitable for
// representing a map. It is a slice type, so it represents a rectangular range
// within an underlying original grid. Due to how it is represented internally,
// it is more efficient to iterate whole lines first, as in the following
// pattern:
//
// 	max := gd.Size()
//	for y := 0; y < max.Y; y++ {
//		for x := 0; x < max.X; x++ {
//			p := Point{X: x, Y: y}
//			// do something with p and the grid gd
//		}
//	}
//
// Most iterations can be performed using the Slice, Fill, Copy, Map and Iter
// methods.
//
// Grid implements gob.Decoder and gob.Encoder for easy serialization.
type Grid struct {
	innerGrid
}

type innerGrid struct {
	Ug *grid       // underlying whole grid
	Rg gruid.Range // range within the whole grid
}

// Cell represents a cell in a rl.Grid.
type Cell int

type grid struct {
	Width  int
	Height int
	Cells  []Cell
}

// NewGrid returns a new grid with given width and height in cells. The width
// and height should be positive or null. The new grid contains all positions
// (X,Y) with 0 <= X < w and 0 <= Y < h. The grid is filled with the zero
// value for cells.
func NewGrid(w, h int) Grid {
	gd := Grid{}
	gd.Ug = &grid{}
	if w < 0 || h < 0 {
		panic(fmt.Sprintf("negative dimensions: NewGrid(%d,%d)", w, h))
	}
	gd = gd.Resize(w, h)
	return gd
}

// GobDecode implements gob.GobDecoder.
func (gd *Grid) GobDecode(bs []byte) error {
	r := bytes.NewReader(bs)
	gdec := gob.NewDecoder(r)
	igd := &innerGrid{}
	err := gdec.Decode(igd)
	if err != nil {
		return err
	}
	gd.innerGrid = *igd
	return nil
}

// GobEncode implements gob.GobEncoder.
func (gd *Grid) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	ge := gob.NewEncoder(&buf)
	err := ge.Encode(&gd.innerGrid)
	return buf.Bytes(), err
}

// Bounds returns the range that is covered by this grid slice within the
// underlying original grid.
func (gd Grid) Bounds() gruid.Range {
	return gd.Rg
}

// Range returns the range with Min set to (0,0) and Max set to gd.Size(). It
// may be convenient when using Slice with a range Shift.
func (gd Grid) Range() gruid.Range {
	return gd.Rg.Sub(gd.Rg.Min)
}

// Slice returns a rectangular slice of the grid given by a range relative to
// the grid. If the range is out of bounds of the parent grid, it will be
// reduced to fit to the available space. The returned grid shares memory with
// the parent.
//
// This makes it easy to use relative coordinates when working with UI
// elements.
func (gd Grid) Slice(rg gruid.Range) Grid {
	if rg.Min.X < 0 {
		rg.Min.X = 0
	}
	if rg.Min.Y < 0 {
		rg.Min.Y = 0
	}
	max := gd.Rg.Size()
	if rg.Max.X > max.X {
		rg.Max.X = max.X
	}
	if rg.Max.Y > max.Y {
		rg.Max.Y = max.Y
	}
	min := gd.Rg.Min
	rg.Min = rg.Min.Add(min)
	rg.Max = rg.Max.Add(min)
	return Grid{innerGrid{Ug: gd.Ug, Rg: rg}}
}

// Size returns the grid (width, height) in cells, and is a shorthand for
// gd.Range().Size().
func (gd Grid) Size() gruid.Point {
	return gd.Rg.Size()
}

// Resize is similar to Slice, but it only specifies new dimensions, and if the
// range goes beyond the underlying original grid range, it will grow the
// underlying grid. It preserves the content, and any new cells get the zero
// value.
func (gd Grid) Resize(w, h int) Grid {
	max := gd.Size()
	ow, oh := max.X, max.Y
	if ow == w && oh == h {
		return gd
	}
	if w <= 0 || h <= 0 {
		gd.Rg.Max = gd.Rg.Min
		return gd
	}
	if gd.Ug == nil {
		gd.Ug = &grid{}
	}
	gd.Rg.Max = gd.Rg.Min.Shift(w, h)
	uw := gd.Ug.Width
	uh := gd.Ug.Height
	grow := false
	if w+gd.Rg.Min.X > uw {
		gd.Ug.Width = w + gd.Rg.Min.X
		grow = true
	}
	if h+gd.Rg.Min.Y > uh {
		gd.Ug.Height = h + gd.Rg.Min.Y
		grow = true
	}
	if grow {
		newBuf := make([]Cell, gd.Ug.Width*gd.Ug.Height)
		for i := range gd.Ug.Cells {
			p := idxToPos(i, uw)         // old absolute position
			idx := p.X + gd.Ug.Width*p.Y // new index
			newBuf[idx] = gd.Ug.Cells[i]
		}
		gd.Ug.Cells = newBuf
	}
	return gd
}

// Contains returns true if the given relative position is within the grid.
func (gd Grid) Contains(p gruid.Point) bool {
	return p.Add(gd.Rg.Min).In(gd.Rg)
}

// Set draws a cell at a given position in the grid. If the position is out of
// range, the function does nothing.
func (gd Grid) Set(p gruid.Point, c Cell) {
	if !gd.Contains(p) {
		return
	}
	i := gd.getIdx(p)
	gd.Ug.Cells[i] = c
}

// At returns the cell at a given position. If the position is out of range, it
// returns the zero value.
func (gd Grid) At(p gruid.Point) Cell {
	if !gd.Contains(p) {
		return Cell(0)
	}
	i := gd.getIdx(p)
	return gd.Ug.Cells[i]
}

// getIdx returns the buffer index of a relative position.
func (gd Grid) getIdx(p gruid.Point) int {
	p = p.Add(gd.Rg.Min)
	return p.Y*gd.Ug.Width + p.X
}

// idxToPos returns a grid position given an index and the width of the grid.
func idxToPos(i, w int) gruid.Point {
	return gruid.Point{X: i % w, Y: i / w}
}

// Fill sets the given cell as content for all the grid positions.
func (gd Grid) Fill(c Cell) {
	w := gd.Rg.Max.X - gd.Rg.Min.X
	switch {
	case w > 8:
		gd.fillcp(c)
	case w == 1:
		gd.fillv(c)
	default:
		gd.fill(c)
	}
}

func (gd Grid) fillcp(c Cell) {
	ug := gd.Ug
	ymin := gd.Rg.Min.Y * ug.Width
	w := gd.Rg.Max.X - gd.Rg.Min.X
	for xi := ymin + gd.Rg.Min.X; xi < ymin+gd.Rg.Max.X; xi++ {
		ug.Cells[xi] = c
	}
	idxmax := (gd.Rg.Max.Y-1)*ug.Width + gd.Rg.Max.X
	for idx := ymin + ug.Width + gd.Rg.Min.X; idx < idxmax; idx += ug.Width {
		copy(ug.Cells[idx:idx+w], ug.Cells[ymin+gd.Rg.Min.X:ymin+gd.Rg.Max.X])
	}
}

func (gd Grid) fill(c Cell) {
	ug := gd.Ug
	yimax := gd.Rg.Max.Y * ug.Width
	for yi := gd.Rg.Min.Y * ug.Width; yi < yimax; yi += ug.Width {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			ug.Cells[xi] = c
		}
	}
}

func (gd Grid) fillv(c Cell) {
	ug := gd.Ug
	yimax := gd.Rg.Max.Y*ug.Width + gd.Rg.Min.X
	for xi := gd.Rg.Min.Y*ug.Width + gd.Rg.Min.X; xi < yimax; xi += ug.Width {
		ug.Cells[xi] = c
	}
}

// FillFunc updates the content for all the grid positions in order using the
// given function return value.
func (gd Grid) FillFunc(fn func() Cell) {
	ug := gd.Ug
	yimax := gd.Rg.Max.Y * ug.Width
	for yi := gd.Rg.Min.Y * ug.Width; yi < yimax; yi += ug.Width {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			ug.Cells[xi] = fn()
		}
	}
}

// Map updates the grid content using the given mapping function.
func (gd Grid) Map(fn func(gruid.Point, Cell) Cell) {
	ug := gd.Ug
	yimax := gd.Rg.Max.Y * ug.Width
	for y, yi := 0, gd.Rg.Min.Y*ug.Width; yi < yimax; y, yi = y+1, yi+ug.Width {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := ug.Cells[xi]
			p := gruid.Point{X: x, Y: y}
			ug.Cells[xi] = fn(p, c)
		}
	}
}

// CountFunc returns the number of cells for which the given function returns
// true.
func (gd Grid) CountFunc(fn func(c Cell) bool) int {
	ug := gd.Ug
	count := 0
	yimax := gd.Rg.Max.Y * ug.Width
	for yi := gd.Rg.Min.Y * ug.Width; yi < yimax; yi += ug.Width {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			c := ug.Cells[xi]
			if fn(c) {
				count++
			}
		}
	}
	return count
}

// Count returns the number of cells which are equal to the given one.
func (gd Grid) Count(c Cell) int {
	ug := gd.Ug
	count := 0
	yimax := gd.Rg.Max.Y * ug.Width
	for yi := gd.Rg.Min.Y * ug.Width; yi < yimax; yi += ug.Width {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			cc := ug.Cells[xi]
			if c == cc {
				count++
			}
		}
	}
	return count
}

// Iter iterates a function on all the grid positions and cells.
func (gd Grid) Iter(fn func(gruid.Point, Cell)) {
	ug := gd.Ug
	yimax := gd.Rg.Max.Y * ug.Width
	for y, yi := 0, gd.Rg.Min.Y*ug.Width; yi < yimax; y, yi = y+1, yi+ug.Width {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := ug.Cells[xi]
			p := gruid.Point{X: x, Y: y}
			fn(p, c)
		}
	}
}

// Copy copies elements from a source grid src into the destination grid gd,
// and returns the copied grid-slice size, which is the minimum of both grids
// for each dimension. The result is independent of whether the two grids
// referenced memory overlaps or not.
func (gd Grid) Copy(src Grid) gruid.Point {
	if gd.Ug != src.Ug {
		return gd.cp(src)
	}
	if gd.Rg == src.Rg {
		return gd.Rg.Size()
	}
	if !gd.Rg.Overlaps(src.Rg) || gd.Rg.Min.Y <= src.Rg.Min.Y {
		return gd.cp(src)
	}
	return gd.cprev(src)
}

func (gd Grid) cp(src Grid) gruid.Point {
	max := gd.Range().Intersect(src.Range()).Size()
	ug, ugsrc := gd.Ug, src.Ug
	idxmin := gd.Rg.Min.Y * ug.Width
	idxsrcmin := src.Rg.Min.Y * ug.Width
	idxmax := (gd.Rg.Min.Y + max.Y) * ug.Width
	for idx, idxsrc := idxmin, idxsrcmin; idx < idxmax; idx, idxsrc = idx+ug.Width, idxsrc+ugsrc.Width {
		copy(ug.Cells[idx:idx+max.X], ugsrc.Cells[idxsrc:idxsrc+max.X])
	}
	return max
}

func (gd Grid) cprev(src Grid) gruid.Point {
	max := gd.Range().Intersect(src.Range()).Size()
	ug, ugsrc := gd.Ug, src.Ug
	idxmax := (gd.Rg.Min.Y + max.Y - 1) * ug.Width
	idxsrcmax := (src.Rg.Min.Y + max.Y - 1) * ug.Width
	idxmin := gd.Rg.Min.Y * ug.Width
	for idx, idxsrc := idxmax, idxsrcmax; idx >= idxmin; idx, idxsrc = idx-ug.Width, idxsrc-ugsrc.Width {
		copy(ug.Cells[idx:idx+max.X], ugsrc.Cells[idxsrc:idxsrc+max.X])
	}
	return max
}
