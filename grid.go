package gruid

import (
	"fmt"
	"strings"
	"time"
)

// AttrMask can be used to add custom styling information. It can for example
// be used to map to specific terminal attributes (with GetStyle), or use
// special images (with GetImage), when appropriate.
//
// It may be used as a bitmask, like terminal attributes, or as a generic
// value for constants.
type AttrMask uint32

// AttrsDefault represents the default styling attributes.
const AttrsDefault AttrMask = 0

// Color is a generic value for representing colors. Those have to be mapped to
// concrete foreground and background colors for each driver, as appropriate.
type Color uint32

// ColorDefault should get special treatment by drivers and be mapped, when it
// makes sense, to a default color, both for foreground and background.
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
func (st Style) WithAttrs(attrs AttrMask) Style {
	st.Attrs = attrs
	return st
}

// Point represents an (X,Y) position in a grid. It follows conventions similar
// to the ones used by the standard library image.Point.
type Point struct {
	X int
	Y int
}

// String returns a string representation of the form "(x,y)".
func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// Shift returns a new point with coordinates shifted by (x,y). It's a
// shorthand for p.Add(Point{x,y}).
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

// In reports whether the position is within the given range.
func (p Point) In(rg Range) bool {
	return p.X >= rg.Min.X && p.X < rg.Max.X && p.Y >= rg.Min.Y && p.Y < rg.Max.Y
}

// Mul returns the vector p*k.
func (p Point) Mul(k int) Point {
	return Point{X: p.X * k, Y: p.Y * k}
}

// Div returns the vector p/k.
func (p Point) Div(k int) Point {
	return Point{X: p.X / k, Y: p.Y / k}
}

// Range represents a rectangle in a grid that contains all the positions P
// such that Min <= P < Max coordinate-wise. A range is well-formed if Min <=
// Max. When non-empty, Min represents the upper-left position in the range,
// and Max-(1,1) the lower-right one.
type Range struct {
	Min, Max Point
}

// NewRange returns a new Range with coordinates (x0, y0) for Min and (x1, y1)
// for Max. The returned range will have minimum and maximum coordinates
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

// String returns a string representation of the form "(x0,y0)-(x1,y1)".
func (rg Range) String() string {
	return fmt.Sprintf("%s-%s", rg.Min, rg.Max)
}

// Size returns the (width, height) of the range in cells.
func (rg Range) Size() Point {
	return rg.Max.Sub(rg.Min)
}

// Shift returns a new range with coordinates shifted by (x0,y0) and (x1,y1).
func (rg Range) Shift(x0, y0, x1, y1 int) Range {
	rg = Range{Min: rg.Min.Shift(x0, y0), Max: rg.Max.Shift(x1, y1)}
	if rg.Empty() {
		return Range{}
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

// Eq reports whether the two ranges containt the same set of points. All empty
// ranges are considered equal.
func (rg Range) Eq(r Range) bool {
	return rg == r || rg.Empty() && r.Empty()
}

// Sub returns a range of same size translated by -p.
func (rg Range) Sub(p Point) Range {
	rg.Max = rg.Max.Sub(p)
	rg.Min = rg.Min.Sub(p)
	return rg
}

// Add returns a range of same size translated by +p.
func (rg Range) Add(p Point) Range {
	rg.Max = rg.Max.Add(p)
	rg.Min = rg.Min.Add(p)
	return rg
}

// RelMsg returns a range-relative version of messages defined by the gruid
// package. Currently, it only affects mouse messages, which are given
// positions relative to the range.
func (rg Range) RelMsg(msg Msg) Msg {
	if msg, ok := msg.(MsgMouse); ok {
		msg.P = msg.P.Sub(rg.Min)
		return msg
	}
	return msg
}

// Intersect returns the largest range contained both by rg and r. If the two
// ranges do not overlap, the zero range will be returned.
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
	if rg.Min.X >= rg.Max.X || rg.Min.Y >= rg.Max.Y {
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
// rectangular range within an underlying original grid. Due to how it is
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
// Most iterations can be performed using the Slice, Fill, Copy, Map and Iter
// methods. An alternative choice is to use the Iterator method.
//
// Grid elements must be created with NewGrid.
type Grid struct {
	innerGrid
}

type innerGrid struct {
	Ug *grid // underlying whole grid
	Rg Range // range within the whole grid
}

type grid struct {
	Cells  []Cell
	Width  int
	Height int
}

// Frame contains the necessary information to draw the frame changes from a
// frame to the next. One is sent to the driver after every Draw.
type Frame struct {
	Time   time.Time   // time of frame drawing: used for replay
	Cells  []FrameCell // cells that changed from previous frame
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
// (X,Y) with 0 <= X < w and 0 <= Y < h. The grid is filled with Cell{Rune: ' '}.
func NewGrid(w, h int) Grid {
	gd := Grid{}
	gd.Ug = &grid{}
	if w < 0 || h < 0 {
		panic(fmt.Sprintf("negative dimensions: NewGrid(%d,%d)", w, h))
	}
	gd.Rg.Max = Point{w, h}
	gd.Ug.Width = w
	gd.Ug.Height = h
	gd.Ug.Cells = make([]Cell, w*h)
	gd.Fill(Cell{Rune: ' '})
	return gd
}

// String returns a simplified string representation of the grid's runes,
// without the styling.
func (gd Grid) String() string {
	b := strings.Builder{}
	it := gd.Iterator()
	for it.Next() {
		b.WriteRune(it.Cell().Rune)
		p := it.P()
		if p.X == gd.Rg.Max.X-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

// Bounds returns the range that is covered by this grid slice within the
// underlying original grid.
func (gd Grid) Bounds() Range {
	return gd.Rg
}

// Range returns the range with Min set to (0,0) and Max set to gd.Size(). It
// may be convenient when using Slice with a range Shift.
func (gd Grid) Range() Range {
	return gd.Rg.Sub(gd.Rg.Min)
}

// Slice returns a rectangular slice of the grid given by a range relative to
// the grid. If the range is out of bounds of the parent grid, it will be
// reduced to fit to the available space. The returned grid shares memory with
// the parent.
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
func (gd Grid) Size() Point {
	return gd.Rg.Size()
}

// Resize is similar to Slice, but it only specifies new dimensions, and if the
// range goes beyond the underlying original grid range, it will grow the
// underlying grid. In case of growth, it preserves the content, and new cells
// are initialized to Cell{Rune: ' '}.
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
		ngd.Copy(Grid{innerGrid{Ug: gd.Ug, Rg: NewRange(0, 0, gd.Ug.Width, gd.Ug.Height)}})
		*gd.Ug = *ngd.Ug
	}
	return gd
}

// Contains returns true if the given relative position is within the grid.
func (gd Grid) Contains(p Point) bool {
	return p.Add(gd.Rg.Min).In(gd.Rg)
}

// Set draws cell content and styling at a given position in the grid. If the
// position is out of range, the function does nothing.
func (gd Grid) Set(p Point, c Cell) {
	q := p.Add(gd.Rg.Min)
	if !q.In(gd.Rg) {
		return
	}
	i := q.Y*gd.Ug.Width + q.X
	gd.Ug.Cells[i] = c
}

// At returns the cell content and styling at a given position. If the position
// is out of range, it returns the zero value.
func (gd Grid) At(p Point) Cell {
	q := p.Add(gd.Rg.Min)
	if !q.In(gd.Rg) {
		return Cell{}
	}
	i := q.Y*gd.Ug.Width + q.X
	return gd.Ug.Cells[i]
}

// Fill sets the given cell as content for all the grid positions.
func (gd Grid) Fill(c Cell) {
	if gd.Ug == nil {
		return
	}
	w := gd.Rg.Max.X - gd.Rg.Min.X
	switch {
	case w >= 8:
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

// Iter iterates a function on all the grid positions and cells.
func (gd Grid) Iter(fn func(Point, Cell)) {
	if gd.Ug == nil {
		return
	}
	w := gd.Ug.Width
	yimax := gd.Rg.Max.Y * w
	cells := gd.Ug.Cells
	for y, yi := 0, gd.Rg.Min.Y*w; yi < yimax; y, yi = y+1, yi+w {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := cells[xi]
			p := Point{X: x, Y: y}
			fn(p, c)
		}
	}
}

// Map updates the grid content using the given mapping function.
func (gd Grid) Map(fn func(Point, Cell) Cell) {
	if gd.Ug == nil {
		return
	}
	w := gd.Ug.Width
	cells := gd.Ug.Cells
	yimax := gd.Rg.Max.Y * w
	for y, yi := 0, gd.Rg.Min.Y*w; yi < yimax; y, yi = y+1, yi+w {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := cells[xi]
			p := Point{X: x, Y: y}
			cells[xi] = fn(p, c)
		}
	}
}

// Copy copies elements from a source grid src into the destination grid gd,
// and returns the copied grid-slice size, which is the minimum of both grids
// for each dimension. The result is independent of whether the two grids
// referenced memory overlaps or not.
func (gd Grid) Copy(src Grid) Point {
	if gd.Ug == nil {
		return Point{}
	}
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

func (gd Grid) cp(src Grid) Point {
	w := gd.Ug.Width
	wsrc := src.Ug.Width
	max := gd.Range().Intersect(src.Range()).Size()
	idxmin := gd.Rg.Min.Y*w + gd.Rg.Min.X
	idxsrcmin := src.Rg.Min.Y*w + src.Rg.Min.X
	idxmax := (gd.Rg.Min.Y + max.Y) * w
	for idx, idxsrc := idxmin, idxsrcmin; idx < idxmax; idx, idxsrc = idx+w, idxsrc+wsrc {
		copy(gd.Ug.Cells[idx:idx+max.X], src.Ug.Cells[idxsrc:idxsrc+max.X])
	}
	return max
}

func (gd Grid) cpv(src Grid) Point {
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

func (gd Grid) cprev(src Grid) Point {
	w := gd.Ug.Width
	wsrc := src.Ug.Width
	max := gd.Range().Intersect(src.Range()).Size()
	idxmax := (gd.Rg.Min.Y+max.Y-1)*w + gd.Rg.Min.X
	idxsrcmax := (src.Rg.Min.Y+max.Y-1)*w + src.Rg.Min.X
	idxmin := gd.Rg.Min.Y * w
	for idx, idxsrc := idxmax, idxsrcmax; idx >= idxmin; idx, idxsrc = idx-w, idxsrc-wsrc {
		copy(gd.Ug.Cells[idx:idx+max.X], src.Ug.Cells[idxsrc:idxsrc+max.X])
	}
	return max
}

// GridIterator represents a stateful iterator for a grid. They are created
// with the Iterator method.
type GridIterator struct {
	cells  []Cell // grid cells
	p      Point  // iterator's current position
	max    Point  // last position
	i      int    // current position's index
	w      int    // underlying grid's width
	nlstep int    // newline step
	rg     Range  // grid range
}

// Iterator returns an iterator that can be used to iterate on the grid. It may
// be convenient when more flexibility than the provided by the other iteration
// functions is needed. It is used as follows:
//
// 	it := gd.Iterator()
// 	for it.Next() {
// 		// call it.P() or it.Cell() or it.SetCell() as appropriate
// 	}
func (gd Grid) Iterator() GridIterator {
	if gd.Ug == nil {
		return GridIterator{}
	}
	w := gd.Ug.Width
	it := GridIterator{
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
	it.p = Point{-1, 0}
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
func (it *GridIterator) P() Point {
	return it.p
}

// SetP sets the iterator's current position.
func (it *GridIterator) SetP(p Point) {
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

// SetCell updates the grid cell at the iterator's current position.
func (it *GridIterator) SetCell(c Cell) {
	it.cells[it.i] = c
}

// computeFrame computes next frame minimal changes and returns them.
func (app *App) computeFrame(gd Grid, exposed bool) Frame {
	if gd.Ug == nil || gd.Rg.Empty() && !exposed {
		return Frame{}
	}
	if app.grid.Ug == nil {
		app.grid = NewGrid(gd.Ug.Width, gd.Ug.Height)
		app.frame.Width = gd.Ug.Width
		app.frame.Height = gd.Ug.Height
	} else if app.grid.Ug.Width != gd.Ug.Width || app.grid.Ug.Height != gd.Ug.Height {
		app.grid = app.grid.Resize(gd.Ug.Width, gd.Ug.Height)
		app.frame.Width = gd.Ug.Width
		app.frame.Height = gd.Ug.Height
	}
	app.frame.Time = time.Now()
	app.frame.Cells = app.frame.Cells[:0]
	if exposed {
		return app.refresh(gd)
	}
	w := gd.Ug.Width
	cells := gd.Ug.Cells
	pcells := app.grid.Ug.Cells // previous cells
	yimax := gd.Rg.Max.Y * w
	for y, yi := 0, gd.Rg.Min.Y*w; yi < yimax; y, yi = y+1, yi+w {
		ximax := yi + gd.Rg.Max.X
		for x, xi := 0, yi+gd.Rg.Min.X; xi < ximax; x, xi = x+1, xi+1 {
			c := cells[xi]
			if c == pcells[xi] {
				continue
			}
			pcells[xi] = c
			p := Point{X: x, Y: y}
			cdraw := FrameCell{Cell: c, P: p}
			app.frame.Cells = append(app.frame.Cells, cdraw)
		}
	}
	return app.frame
}

// refresh forces a complete redraw of the screen, even for cells that did not
// change.
func (app *App) refresh(gd Grid) Frame {
	gd.Rg.Min = Point{0, 0}
	gd.Rg.Max = gd.Rg.Min.Add(Point{gd.Ug.Width, gd.Ug.Height})
	app.grid.Copy(gd)
	it := gd.Iterator()
	for it.Next() {
		cdraw := FrameCell{Cell: it.Cell(), P: it.P()}
		app.frame.Cells = append(app.frame.Cells, cdraw)
	}
	return app.frame
}
