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
// methods. An alternative choice is to use the Iterator method.
//
// Grid elements must be created with NewGrid.
//
// Grid implements gob.Decoder and gob.Encoder for easy serialization.
type Grid struct {
	innerGrid
}

type innerGrid struct {
	Ug *grid       // underlying whole grid
	Rg gruid.Range // range within the whole grid
}

// Cell represents a cell in a map Grid, commonly a terrain type or other
// information associated with a map position.
type Cell int

type grid struct {
	Cells  []Cell
	Width  int
	Height int
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
	gd.Rg.Max = gruid.Point{w, h}
	gd.Ug.Width = w
	gd.Ug.Height = h
	gd.Ug.Cells = make([]Cell, w*h)
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
	if w+gd.Rg.Min.X > gd.Ug.Width || h+gd.Rg.Min.Y > gd.Ug.Height {
		ngd := NewGrid(w+gd.Rg.Min.X, h+gd.Rg.Min.Y)
		ngd.Copy(Grid{innerGrid{Ug: gd.Ug, Rg: gruid.NewRange(0, 0, gd.Ug.Width, gd.Ug.Height)}})
		*gd.Ug = *ngd.Ug
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
	q := p.Add(gd.Rg.Min)
	if !q.In(gd.Rg) {
		return
	}
	i := q.Y*gd.Ug.Width + q.X
	gd.Ug.Cells[i] = c
}

// At returns the cell at a given position. If the position is out of range, it
// returns the zero value.
func (gd Grid) At(p gruid.Point) Cell {
	q := p.Add(gd.Rg.Min)
	if !q.In(gd.Rg) {
		return Cell(0)
	}
	i := q.Y*gd.Ug.Width + q.X
	return gd.Ug.Cells[i]
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
	w := gd.Ug.Width
	ymin := gd.Rg.Min.Y * w
	gdw := gd.Rg.Max.X - gd.Rg.Min.X
	cells := gd.Ug.Cells
	for xi := ymin + gd.Rg.Min.X; xi < ymin+gd.Rg.Max.X; xi++ {
		cells[xi] = c
	}
	idxmax := (gd.Rg.Max.Y-1)*w + gd.Rg.Max.X
	for idx := ymin + w + gd.Rg.Min.X; idx < idxmax; idx += w {
		copy(cells[idx:idx+gdw], cells[ymin+gd.Rg.Min.X:ymin+gd.Rg.Max.X])
	}
}

func (gd Grid) fill(c Cell) {
	w := gd.Ug.Width
	cells := gd.Ug.Cells
	yimax := gd.Rg.Max.Y * w
	for yi := gd.Rg.Min.Y * w; yi < yimax; yi += w {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			cells[xi] = c
		}
	}
}

func (gd Grid) fillv(c Cell) {
	w := gd.Ug.Width
	cells := gd.Ug.Cells
	ximax := gd.Rg.Max.Y*w + gd.Rg.Min.X
	for xi := gd.Rg.Min.Y*w + gd.Rg.Min.X; xi < ximax; xi += w {
		cells[xi] = c
	}
}

// FillFunc updates the content for all the grid positions in order using the
// given function return value.
func (gd Grid) FillFunc(fn func() Cell) {
	w := gd.Ug.Width
	yimax := gd.Rg.Max.Y * w
	cells := gd.Ug.Cells
	for yi := gd.Rg.Min.Y * w; yi < yimax; yi += w {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			cells[xi] = fn()
		}
	}
}

// Iter iterates a function on all the grid positions and cells.
func (gd Grid) Iter(fn func(gruid.Point, Cell)) {
	w := gd.Ug.Width
	yimax := gd.Rg.Max.Y * w
	cells := gd.Ug.Cells
	for y, yi := 0, gd.Rg.Min.Y*w; yi < yimax; y, yi = y+1, yi+w {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := cells[xi]
			p := gruid.Point{X: x, Y: y}
			fn(p, c)
		}
	}
}

// Map updates the grid content using the given mapping function.
func (gd Grid) Map(fn func(gruid.Point, Cell) Cell) {
	w := gd.Ug.Width
	cells := gd.Ug.Cells
	yimax := gd.Rg.Max.Y * w
	for y, yi := 0, gd.Rg.Min.Y*w; yi < yimax; y, yi = y+1, yi+w {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := cells[xi]
			p := gruid.Point{X: x, Y: y}
			cells[xi] = fn(p, c)
		}
	}
}

// CountFunc returns the number of cells for which the given function returns
// true.
func (gd Grid) CountFunc(fn func(c Cell) bool) int {
	w := gd.Ug.Width
	count := 0
	yimax := gd.Rg.Max.Y * w
	cells := gd.Ug.Cells
	for yi := gd.Rg.Min.Y * w; yi < yimax; yi += w {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			c := cells[xi]
			count += bool2int(fn(c))
		}
	}
	return count
}

// Count returns the number of cells which are equal to the given one.
func (gd Grid) Count(c Cell) int {
	w := gd.Ug.Width
	count := 0
	yimax := gd.Rg.Max.Y * w
	cells := gd.Ug.Cells
	for yi := gd.Rg.Min.Y * w; yi < yimax; yi += w {
		ximax := yi + gd.Rg.Max.X
		for xi := yi + gd.Rg.Min.X; xi < ximax; xi++ {
			cc := cells[xi]
			count += bool2int(cc == c)
		}
	}
	return count
}

func bool2int(b bool) int {
	var i int
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// Copy copies elements from a source grid src into the destination grid gd,
// and returns the copied grid-slice size, which is the minimum of both grids
// for each dimension. The result is independent of whether the two grids
// referenced memory overlaps or not.
func (gd Grid) Copy(src Grid) gruid.Point {
	if gd.Ug != src.Ug {
		if src.Rg.Max.X-src.Rg.Min.X <= 4 {
			return gd.cpv(src)
		}
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
	w := gd.Ug.Width
	wsrc := src.Ug.Width
	max := gd.Range().Intersect(src.Range()).Size()
	idxmin := gd.Rg.Min.Y * w
	idxsrcmin := src.Rg.Min.Y * w
	idxmax := (gd.Rg.Min.Y + max.Y) * w
	for idx, idxsrc := idxmin, idxsrcmin; idx < idxmax; idx, idxsrc = idx+w, idxsrc+wsrc {
		copy(gd.Ug.Cells[idx:idx+max.X], src.Ug.Cells[idxsrc:idxsrc+max.X])
	}
	return max
}

func (gd Grid) cpv(src Grid) gruid.Point {
	w := gd.Ug.Width
	wsrc := src.Ug.Width
	max := gd.Range().Intersect(src.Range()).Size()
	yimax := (gd.Rg.Min.Y + max.Y) * w
	cells := gd.Ug.Cells
	srccells := src.Ug.Cells
	for yi, yisrc := gd.Rg.Min.Y*w, src.Rg.Min.Y*wsrc; yi < yimax; yi, yisrc = yi+w, yisrc+wsrc {
		ximax := yi + max.X
		for xi, xisrc := yi+gd.Rg.Min.X, yisrc+src.Rg.Min.X; xi < ximax; xi, xisrc = xi+1, xisrc+1 {
			cells[xi] = srccells[xisrc]
		}
	}
	return max
}

func (gd Grid) cprev(src Grid) gruid.Point {
	w := gd.Ug.Width
	wsrc := src.Ug.Width
	max := gd.Range().Intersect(src.Range()).Size()
	idxmax := (gd.Rg.Min.Y + max.Y - 1) * w
	idxsrcmax := (src.Rg.Min.Y + max.Y - 1) * w
	idxmin := gd.Rg.Min.Y * w
	for idx, idxsrc := idxmax, idxsrcmax; idx >= idxmin; idx, idxsrc = idx-w, idxsrc-wsrc {
		copy(gd.Ug.Cells[idx:idx+max.X], src.Ug.Cells[idxsrc:idxsrc+max.X])
	}
	return max
}

// GridIterator represents a stateful iterator for a grid. They are created
// with the Iterator method.
type GridIterator struct {
	cells  []Cell      // grid cells
	p      gruid.Point // iterator's current position
	max    gruid.Point // last position
	i      int         // current position's index
	w      int         // underlying grid's width
	nlstep int         // newline step
	rg     gruid.Range // grid range
}

// Iterator returns an iterator that can be used to iterate on the grid. It may
// be convenient when more flexibility than the provided by the other iteration
// functions is needed. It is used as follows:
//
// 	it := gd.Iterator()
// 	for it.Next() {
// 		// call it.P() or it.Cell() or it.SetCell() as appropriate
// 	}
func (gd Grid) Iterator() *GridIterator {
	w := gd.Ug.Width
	it := &GridIterator{
		w:      w,
		cells:  gd.Ug.Cells,
		max:    gd.Size().Shift(-1, -1),
		rg:     gd.Rg,
		nlstep: gd.Rg.Min.X + (w - gd.Rg.Max.X + 1),
	}
	it.Reset()
	return it
}

// Reset resets the iterator's state so that it can be used again.
func (it *GridIterator) Reset() {
	it.p = gruid.Point{-1, 0}
	it.i = it.rg.Min.Y*it.w + it.rg.Min.X - 1
}

// Next advances the iterator the next position in the grid.
func (it *GridIterator) Next() bool {
	if it.p.X < it.max.X {
		it.p.X++
		it.i++
		return true
	}
	if it.p.Y < it.max.Y {
		it.p.Y++
		it.p.X = 0
		it.i += it.nlstep
		return true
	}
	return false
}

// P returns the iterator's current position.
func (it *GridIterator) P() gruid.Point {
	return it.p
}

// SetP sets the iterator's current position.
func (it *GridIterator) SetP(p gruid.Point) {
	q := p.Add(it.rg.Min)
	if !q.In(it.rg) {
		return
	}
	it.p = p
	it.i = q.Y*it.w + q.X
}

// Cell returns the Cell in the grid at the iterator's current position.
func (it *GridIterator) Cell() Cell {
	return it.cells[it.i]
}

// SetCell updates the grid cell at the iterator's current position. It's
// faster than calling Set on the grid.
func (it *GridIterator) SetCell(c Cell) {
	it.cells[it.i] = c
}
