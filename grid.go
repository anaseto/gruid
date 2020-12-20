package gruid

import (
	"fmt"
	"time"
)

// AttrMask can be used to add custom styling information. It can for example
// be used to map to specific terminal attributes (with GetStyle), or use
// special images (with GetImage), when appropiate.
//
// It may be used as a bitmask, like terminal attributes, or as a generic
// value for constants.
type AttrMask uint

// Color is a generic value for representing colors. Except for the zero value,
// which gets special treatment, those have to be mapped to concrete foreground
// and background colors for each driver, as appropiate.
type Color uint

// ColorDefault gets special treatment by drivers and is mapped, when it makes
// sense, to a default color, both for foreground and background.
const ColorDefault Color = 0

// Cell contains all the content and styling information to represent a cell in
// the grid.
type Cell struct {
	Rune  rune  // cell content character
	Style Style // cell style
}

// WithRune returns a derived Cell with a new Rune.
func (c Cell) WithRune(r rune) Cell {
	c.Rune = r
	return c
}

// WithStyle returns a derived Cell with a new Style.
func (c Cell) WithStyle(st Style) Cell {
	c.Style = st
	return c
}

// Style represents the styling information of a cell: foreground color,
// background color and custom attributes.
type Style struct {
	Fg    Color    // foreground color
	Bg    Color    // background color
	Attrs AttrMask // custom styling attributes
}

// WithFg returns a derived style with a new foreground color.
func (st Style) WithFg(cl Color) Style {
	st.Fg = cl
	return st
}

// WithBg returns a derived style with a new background color.
func (st Style) WithBg(cl Color) Style {
	st.Bg = cl
	return st
}

// WithAttrs returns a derived style with new attributes.
func (st Style) WithAttrs(cl Color) Style {
	st.Bg = cl
	return st
}

// Point represents an (X,Y) position in a grid.
type Point struct {
	X int
	Y int
}

// Shift returns a new point with coordinates shifted by (x,y).
func (p Point) Shift(x, y int) Point {
	return Point{X: p.X + x, Y: p.Y + y}
}

// Add returns vector p+q.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// Sub returns vector p-q.
func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

// In reports whether the absolute position is withing the given range.
func (p Point) In(rg Range) bool {
	return p.X >= rg.Min.X && p.Y >= rg.Min.Y && p.X < rg.Max.X && p.Y < rg.Max.Y
}

// Rel changes an absolute position into a position relative to a given
// range. It's the same as p.Sub(rg.Min). You may use it for example when
// dealing with mouse coordinates from a MsgMouse message. See also the method
// of the same name for the Range type, which may serve a similar purpose.
func (p Point) Rel(rg Range) Point {
	return p.Sub(rg.Min)
}

// Abs returns the absolute position given a range. It's the same as
// p.Add(rg.Min).
func (p Point) Abs(rg Range) Point {
	return p.Add(rg.Min)
}

// Range represents a rectangle in a grid with upper left position Min and
// bottom right position Max (excluded). In other terms, it contains all the
// positions P such that Min <= P < Max. A range is well-formed if Min <= Max.
// Min represents the upper-left position in the range, and Max-(1,1) the
// lower-right one.
type Range struct {
	Min, Max Point
}

// NewRange returns a new Range with coordinates (x0, y0) for Min and (x1, y1)
// for Max. The returned range will have minumum and maximum coordinates
// swapped if necessary, so that the range is well-formed.
func NewRange(x0, y0, x1, y1 int) Range {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	return Range{Min: Point{X: x0, Y: y0}, Max: Point{X: x1, Y: y1}}
}

// Size returns the (width, height) of the range in cells.
func (rg Range) Size() Point {
	return rg.Max.Sub(rg.Min)
}

// Shift returns a new range with coordinates shifted by (x0,y0) and (x1,y1).
func (rg Range) Shift(x0, y0, x1, y1 int) Range {
	rg = Range{Min: rg.Min.Shift(x0, y0), Max: rg.Max.Shift(x1, y1)}
	if rg.Min.X > rg.Max.X {
		rg.Min.X = rg.Max.X
	}
	if rg.Min.Y > rg.Max.Y {
		rg.Max.Y = rg.Min.Y
	}
	return rg
}

// Line reduces the range to relative line y, or an empty range if out of
// bounds.
func (rg Range) Line(y int) Range {
	if rg.Min.Shift(0, y).In(rg) {
		rg.Min.Y = rg.Min.Y + y
		rg.Max.Y = rg.Min.Y + 1
	} else {
		rg = Range{}
	}
	return rg
}

// Lines reduces the range to relative lines between y0 (included) and y1
// (excluded), or an empty range if out of bounds.
func (rg Range) Lines(y0, y1 int) Range {
	nrg := rg
	nrg.Min.Y = rg.Min.Y + y0
	nrg.Max.Y = rg.Min.Y + y1
	return rg.Intersect(nrg)
}

// Column reduces the range to relative column x, or an empty range if out of
// bounds.
func (rg Range) Column(x int) Range {
	if rg.Min.Shift(x, 0).In(rg) {
		rg.Min.X = rg.Min.X + x
		rg.Max.X = rg.Min.X + 1
	} else {
		rg = Range{}
	}
	return rg
}

// Columns reduces the range to relative columns between x0 (included) and x1
// (excluded), or an empty range if out of bounds.
func (rg Range) Columns(x0, x1 int) Range {
	nrg := rg
	nrg.Min.X = rg.Min.X + x0
	nrg.Max.X = rg.Min.X + x1
	return rg.Intersect(nrg)
}

// Empty reports whether the range contains no positions.
func (rg Range) Empty() bool {
	return rg.Min.X >= rg.Max.X || rg.Min.Y >= rg.Max.Y
}

// Origin returns a range of same size with Min = (0, 0). It may be useful to
// define grid slices as a Shift of a relative original range.
func (rg Range) Origin() Range {
	rg.Max = rg.Max.Sub(rg.Min)
	rg.Min = Point{}
	return rg
}

// Relative returns a range-relative version of messages defined by the gruid
// package. Currently, it only affects mouse messages, which are given
// positions relative to the range.
func (rg Range) Relative(msg Msg) Msg {
	if msg, ok := msg.(MsgMouse); ok {
		msg.P = msg.P.Rel(rg)
		return msg
	}
	return msg
}

// Intersect returns the largest range contained both by rg and r. If the two
// ranges dot not overlap, the zero range will be returned.
func (rg Range) Intersect(r Range) Range {
	if rg.Max.X > r.Max.X {
		rg.Max.X = r.Max.X
	}
	if rg.Max.Y > r.Max.Y {
		rg.Max.Y = r.Max.Y
	}
	if rg.Min.X < r.Min.X {
		rg.Min.X = r.Min.X
	}
	if rg.Min.Y < r.Min.Y {
		rg.Min.Y = r.Min.Y
	}
	if rg.Empty() {
		return Range{}
	}
	return rg
}

// Union returns the smallest range containing both rg and r.
func (rg Range) Union(r Range) Range {
	if rg.Max.X < r.Max.X {
		rg.Max.X = r.Max.X
	}
	if rg.Max.Y < r.Max.Y {
		rg.Max.Y = r.Max.Y
	}
	if rg.Min.X > r.Min.X {
		rg.Min.X = r.Min.X
	}
	if rg.Min.Y > r.Min.Y {
		rg.Min.Y = r.Min.Y
	}
	return rg
}

// Overlaps reports whether the two ranges have a non-zero intersection.
func (rg Range) Overlaps(r Range) bool {
	return !rg.Intersect(r).Empty()
}

// In reports whether range rg is completely contained in range r.
func (rg Range) In(r Range) bool {
	return rg.Intersect(r) == rg
}

// Iter calls a given function for all the positions of the range.
func (rg Range) Iter(fn func(Point)) {
	for y := rg.Min.Y; y < rg.Max.Y; y++ {
		for x := rg.Min.X; x < rg.Max.X; x++ {
			p := Point{X: x, Y: y}
			fn(p)
		}
	}
}

// Grid represents the grid that is used to draw a model logical contents that
// are then sent to the driver. It is a slice type, so it represents a
// rectangular range within the whole grid of the application. Due to how it is
// represented internally, it is more efficient to iterate whole lines first,
// as in the following pattern:
//
// 	max := gd.Size()
//	for y := 0; y < max.Y; y++ {
//		for x := 0; x < max.X; x++ {
//			p := Point{X: x, Y: y}
//			// do something with p and the grid gd
//		}
//	}
//
// You may also look into the Iter and Fill methods.
type Grid struct {
	ug *grid // underlying whole grid
	rg Range // range within the whole grid
}

type grid struct {
	width  int
	height int
	cells  []Cell
}

// Frame contains the necessary information to draw the frame changes from a
// frame to the next. One is sent to the driver after every Draw.
type Frame struct {
	Cells  []FrameCell // cells that changed from previous frame
	Time   time.Time   // time of frame drawing: used for replay
	Width  int         // width of the whole grid when the frame was issued
	Height int         // height of the whole grid when the frame was issued
}

// FrameCell represents a cell drawing instruction at a specific absolute
// position in the whole grid.
type FrameCell struct {
	Cell Cell  // cell content and styling
	P    Point // absolute position in the whole grid
}

// NewGrid returns a new grid with given width and height in cells. The width
// and height should be positive or null. The new grid contains all positions
// (X,Y) with 0 <= X < w and 0 <= Y < h.
func NewGrid(w, h int) Grid {
	gd := Grid{}
	gd.ug = &grid{}
	if w < 0 || h < 0 {
		panic(fmt.Sprintf("negative dimensions: NewGrid(%d,%d)", w, h))
	}
	gd = gd.Resize(w, h)
	return gd
}

// Range returns the range that is represented by this grid within the
// application's whole grid.
func (gd Grid) Range() Range {
	return gd.rg
}

// Slice returns a rectangular slice of the grid given by a range relative to
// the grid. If the range is out of bounds of the parent grid, it will be
// reduced to fit to the available space.
//
// This makes it easy to use relative coordinates when working with UI
// elements.
func (gd Grid) Slice(rg Range) Grid {
	if rg.Min.X < 0 {
		rg.Min.X = 0
	}
	if rg.Min.Y < 0 {
		rg.Min.Y = 0
	}
	max := gd.rg.Size()
	if rg.Max.X > max.X {
		rg.Max.X = max.X
	}
	if rg.Max.Y > max.Y {
		rg.Max.Y = max.Y
	}
	min := gd.rg.Min
	rg.Min = rg.Min.Add(min)
	rg.Max = rg.Max.Add(min)
	return Grid{ug: gd.ug, rg: rg}
}

// Size returns the grid (width, height) in cells, and is a shorthand for
// gd.Range().Size().
func (gd Grid) Size() Point {
	return gd.rg.Size()
}

// Resize is similar to Slice, but it only specifies new dimensions, and if the
// range goes beyond the underlying grid range, it will grow the underlying
// grid.
//
// Note that this only modifies the size of the grid, which may be different
// than the window screen size.
func (gd Grid) Resize(w, h int) Grid {
	max := gd.Size()
	ow, oh := max.X, max.Y
	if ow == w && oh == h {
		return gd
	}
	if w <= 0 || h <= 0 {
		gd.rg.Max = gd.rg.Min
		return gd
	}
	if gd.ug == nil {
		gd.ug = &grid{}
	}
	gd.rg.Max = gd.rg.Min.Shift(w, h)
	uw := gd.ug.width
	uh := gd.ug.height
	grow := false
	if w+gd.rg.Min.X > uw {
		gd.ug.width = w + gd.rg.Min.X
		grow = true
	}
	if h+gd.rg.Min.Y > uh {
		gd.ug.height = h + gd.rg.Min.Y
		grow = true
	}
	if grow {
		newBuf := make([]Cell, gd.ug.width*gd.ug.height)
		for i := range newBuf {
			newBuf[i] = Cell{Rune: ' '}
		}
		for i := range gd.ug.cells {
			p := idxToPos(i, uw)         // old absolute position
			idx := p.X + gd.ug.width*p.Y // new index
			newBuf[idx] = gd.ug.cells[i]
		}
		gd.ug.cells = newBuf
	}
	return gd
}

// Contains returns true if the relative position is within the grid range.
func (gd Grid) Contains(p Point) bool {
	return p.Abs(gd.rg).In(gd.rg)
}

// Set draws cell content and styling at a given position in the grid. If the
// position is out of range, the function does nothing.
func (gd Grid) Set(p Point, c Cell) {
	if !gd.Contains(p) {
		return
	}
	i := gd.getIdx(p)
	gd.ug.cells[i] = c
}

// At returns the cell content and styling at a given position. If the position
// is out of range, it returns de zero value. The returned cell is the content
// as it is in the logical grid, which may be different from what is currently
// displayed on the screen.
func (gd Grid) At(p Point) Cell {
	if !gd.Contains(p) {
		return Cell{}
	}
	i := gd.getIdx(p)
	return gd.ug.cells[i]
}

// getIdx returns the buffer index of a relative position.
func (gd Grid) getIdx(p Point) int {
	p = p.Abs(gd.rg)
	return p.Y*gd.ug.width + p.X
}

// idxToPos returns a grid position given an index and the width of the grid.
func idxToPos(i, w int) Point {
	return Point{X: i - (i/w)*w, Y: i / w}
}

// Fill sets the given cell as content for all the grid positions.
func (gd Grid) Fill(c Cell) {
	max := gd.Size()
	upos := gd.Range().Min
	for y := 0; y < max.Y; y++ {
		yidx := (upos.Y + y) * gd.ug.width
		for x := 0; x < max.X; x++ {
			xidx := x + upos.X
			gd.ug.cells[xidx+yidx] = c
		}
	}
}

// Iter iterates a function on all the grid positions and cells.
func (gd Grid) Iter(fn func(Point, Cell)) {
	max := gd.Size()
	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			p := Point{X: x, Y: y}
			c := gd.At(p)
			fn(p, c)
		}
	}
}

// Copy copies elements from a source grid src into the destination grid gd,
// and returns the copied grid-slice size, which is the minimum of both grids
// for each dimension. The result is independent of whether the two grids
// referenced memory overlaps or not.
func (gd Grid) Copy(src Grid) Point {
	if gd.ug != src.ug {
		return gd.cp(src)
	}
	if gd.Range() == src.Range() {
		return gd.Range().Size()
	}
	if !gd.Range().Overlaps(src.Range()) || gd.Range().Min.Y <= src.Range().Min.Y {
		return gd.cp(src)
	}
	return gd.cprev(src)
}

func (gd Grid) cp(src Grid) Point {
	rg := gd.Range()
	rgsrc := src.Range()
	max := rg.Origin().Intersect(rgsrc.Origin()).Size()
	for j := 0; j < max.Y; j++ {
		idx := (rg.Min.Y+j)*gd.ug.width + rg.Min.X
		idxsrc := (rgsrc.Min.Y+j)*src.ug.width + rgsrc.Min.X
		copy(gd.ug.cells[idx:idx+max.X], src.ug.cells[idxsrc:idxsrc+max.X])
	}
	return max
}

func (gd Grid) cprev(src Grid) Point {
	rg := gd.Range()
	rgsrc := src.Range()
	max := rg.Origin().Intersect(rgsrc.Origin()).Size()
	for j := max.Y - 1; j >= 0; j-- {
		idx := (rg.Min.Y+j)*gd.ug.width + rg.Min.X
		idxsrc := (rgsrc.Min.Y+j)*src.ug.width + rgsrc.Min.X
		copy(gd.ug.cells[idx:idx+max.X], src.ug.cells[idxsrc:idxsrc+max.X])
	}
	return max
}

// computeFrame computes next frame minimal changes and returns them.
func (app *App) computeFrame(gd Grid) Frame {
	if gd.ug == nil {
		app.frame.Cells = app.frame.Cells[:0]
		return Frame{}
	}
	ug := gd.ug
	if len(app.cellbuf) < len(ug.cells) {
		app.cellbuf = make([]Cell, len(ug.cells))
	}
	app.frame.Time = time.Now()
	app.frame.Width = ug.width
	app.frame.Height = ug.height
	app.frame.Cells = app.frame.Cells[:0]
	for i, c := range ug.cells {
		if c == app.cellbuf[i] {
			continue
		}
		p := idxToPos(i, ug.width)
		cdraw := FrameCell{Cell: c, P: p}
		app.frame.Cells = append(app.frame.Cells, cdraw)
		app.cellbuf[i] = c
	}
	return app.frame
}

// clearCache clears internal cache buffers, forcing a complete redraw of the
// screen with the next Draw call, even for cells that did not change.
func (app *App) clearCellCache() {
	for i := range app.cellbuf {
		app.cellbuf[i] = Cell{}
	}
}
