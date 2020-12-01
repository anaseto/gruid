package gruid

import (
	"time"
)

// AttrMask can be used to add custom styling information. It can for example
// be used to map to specific terminal attributes (with GetStyle), or use
// special images (with GetImage), when appropiate.
//
// It may be used as a bitmask, like terminal attributes, or as a generic
// value for constants.
type AttrMask uint

// Color is a generic value for representing colors. Those have to be mapped to
// concrete colors for each driver, as appropiate.
type Color uint

// Cell contains all the content and styling information to represent a cell in
// the grid.
type Cell struct {
	Fg    Color    // foreground color
	Bg    Color    // background color
	Rune  rune     // cell character
	Attrs AttrMask // custom styling attributes
}

// Grid represents the grid that is used to draw a model logical contents that
// are then sent to the driver. It is a slice type, so it represents a
// rectangular range within the main grid of the application, which can be
// smaller than the whole grid.
type Grid struct {
	ug *grid // underlying main grid
	rg Range // range within the main grid
}

// Position represents an (X,Y) position in a grid.
type Position struct {
	X int
	Y int
}

// Range represents a rectangle in a grid with upper left position Pos, and a
// given Width and Height.
type Range struct {
	Pos    Position // upper left position
	Width  int      // width in cells
	Height int      // height in cells
}

// Relative returns a position relative to the range, given an absolute
// position in the grid. You may use it for example when dealing with mouse
// coordinates from a MsgMouseDown or a MsgMouseMove message.
func (rg Range) Relative(pos Position) Position {
	return Position{X: pos.X - rg.Pos.X, Y: pos.Y - rg.Pos.Y}
}

// Absolute returns an absolute position in a grid, given a position relative
// to the range.
func (rg Range) Absolute(pos Position) Position {
	return Position{X: pos.X + rg.Pos.X, Y: pos.Y + rg.Pos.Y}
}

type grid struct {
	width          int
	height         int
	cellBuffer     []Cell
	cellBackBuffer []Cell
	frame          Frame
	frames         []Frame
	recording      bool
}

// GridConfig is used to configure a new main grid with NewGrid.
type GridConfig struct {
	Width     int  // width in cells (default is 80)
	Height    int  // height in cells (default is 24)
	Recording bool // whether to record frames to enable replay
}

// Frame contains the necessary information to draw the frame changes from a
// frame to the next.
type Frame struct {
	Cells  []FrameCell // cells that changed from previous frame
	Time   time.Time   // time of frame drawing: used for replay
	Width  int         // width of the grid when the frame was produced
	Height int         // height of the grid when the frame was produced
}

// FrameCell represents a cell drawing instruction at a specific position in
// the main grid.
type FrameCell struct {
	Cell Cell
	Pos  Position
}

// NewGrid returns a new grid.
func NewGrid(cfg GridConfig) Grid {
	gd := Grid{}
	gd.ug = &grid{}
	if cfg.Height <= 0 {
		cfg.Height = 24
	}
	if cfg.Width <= 0 {
		cfg.Width = 80
	}
	gd = gd.Resize(cfg.Width, cfg.Height)
	gd.ug.recording = cfg.Recording
	return gd
}

// Range returns the range that is represented by this grid within the
// application's main grid.
func (gd Grid) Range() Range {
	return gd.rg
}

// Slice returns a rectangular slice of the grid given by a range. If the range
// is out of bounds of the parent grid, it will be reduced to fit.
//
// This makes it easy to use relative coordinates when working with UI
// elements.
func (gd Grid) Slice(rg Range) Grid {
	if rg.Width < 0 {
		rg.Width = 0
	}
	if rg.Height < 0 {
		rg.Height = 0
	}
	if rg.Pos.X+rg.Width > gd.rg.Width {
		gd.rg.Width = rg.Pos.X + rg.Width
	}
	if rg.Pos.Y+rg.Height > gd.rg.Height {
		gd.rg.Height = rg.Pos.Y + rg.Height
	}
	rg.Pos.X = gd.rg.Pos.X + rg.Pos.X
	rg.Pos.Y = gd.rg.Pos.Y + rg.Pos.Y
	return Grid{ug: gd.ug, rg: rg}
}

// Size returns the (width, height) parts of the grid range.
func (gd Grid) Size() (int, int) {
	return gd.rg.Width, gd.rg.Height
}

// Resize is similar to Slice, but it only specifies new dimensions, and if the
// range goes beyond the underlying grid range, it will grow the underlying
// grid.
//
// Note that this only modifies the size of the grid, which may be different
// than the window screen size.
func (gd Grid) Resize(w, h int) Grid {
	if gd.rg.Width == w && gd.rg.Height == h {
		return gd
	}
	gd.rg.Width = w
	gd.rg.Height = h
	uw := gd.ug.width
	uh := gd.ug.height
	grow := false
	if w+gd.rg.Pos.X > uw {
		gd.ug.width = w + gd.rg.Pos.X
		grow = true
	}
	if h+gd.rg.Pos.Y > uh {
		gd.ug.height = h + gd.rg.Pos.Y
		grow = true
	}
	if grow {
		newBuf := make([]Cell, gd.ug.width*gd.ug.height)
		for i := 0; i < len(gd.ug.cellBuffer); i++ {
			pos := idxToPos(i, uw) // old absolute position
			idx := gd.getIdx(pos)
			if idx >= 0 && idx < len(newBuf) { // should always be the case
				newBuf[idx] = gd.ug.cellBuffer[i]
			}
		}
		gd.ug.cellBuffer = newBuf
		gd.ug.cellBackBuffer = make([]Cell, len(gd.ug.cellBuffer))
	}
	return gd
}

func (gd Grid) Valid(pos Position) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < gd.rg.Width && pos.Y < gd.rg.Height
}

// SetCell draws cell content and styling at a given position in the grid. If
// the position is out of range, the function does nothing.
func (gd Grid) SetCell(pos Position, c Cell) {
	if !gd.Valid(pos) {
		return
	}
	i := gd.getIdx(pos)
	if i >= len(gd.ug.cellBuffer) || i < 0 {
		return
	}
	gd.ug.cellBuffer[i] = c
}

// GetCell returns the cell content and styling at a given position. If the
// position is out of range, it returns de zero value. The returned cell is the
// content as it is in the logical grid, which may be different from what is
// currently displayed on the screen.
func (gd Grid) GetCell(pos Position) Cell {
	if !gd.Valid(pos) {
		return Cell{}
	}
	i := gd.getIdx(pos)
	if i >= len(gd.ug.cellBuffer) || i < 0 {
		return Cell{}
	}
	return gd.ug.cellBuffer[i]
}

// Iter calls a given function for all the positions of the grid.
func (gd Grid) Iter(fn func(Position)) {
	xmax, ymax := gd.Size()
	for y := 0; y < ymax; y++ {
		for x := 0; x < xmax; x++ {
			pos := Position{X: x, Y: y}
			fn(pos)
		}
	}
}

// getIdx returns the buffer index of a relative position.
func (gd Grid) getIdx(pos Position) int {
	pos = gd.rg.Absolute(pos)
	return pos.Y*gd.ug.width + pos.X
}

// idxToPos returns a grid position given an index and the width of the grid.
func idxToPos(i, w int) Position {
	return Position{X: i - (i/w)*w, Y: i / w}
}

// Frame returns the drawing instructions produced by last Draw call.
//
// This function may be used to implement new drivers. You should normally not
// call it by hand in your application code.
func (gd Grid) Frame() Frame {
	return gd.ug.frame
}

// Draw computes next frame changes which can be retrieved by calling Frame().
// If recording is activated the frame changes are recorded, and can be
// retrieved later by calling Frames().
//
// This function is automatically called after each Draw of the Model. You
// should normally not call it by hand when implementing an application using a
// Model. It is provided just in case you want to use a grid without an
// application and a model.
func (gd Grid) ComputeFrame() Grid {
	// XXX: unexport?
	if len(gd.ug.cellBackBuffer) != len(gd.ug.cellBuffer) {
		gd.ug.cellBackBuffer = make([]Cell, len(gd.ug.cellBuffer))
	}
	gd.ug.frame = Frame{Time: time.Now(), Width: gd.ug.width, Height: gd.ug.height}
	for i := 0; i < len(gd.ug.cellBuffer); i++ {
		if gd.ug.cellBuffer[i] == gd.ug.cellBackBuffer[i] {
			continue
		}
		c := gd.ug.cellBuffer[i]
		pos := idxToPos(i, gd.ug.width)
		cdraw := FrameCell{Cell: c, Pos: pos}
		gd.ug.frame.Cells = append(gd.ug.frame.Cells, cdraw)
		gd.ug.cellBackBuffer[i] = c
	}
	if gd.ug.recording {
		gd.ug.frames = append(gd.ug.frames, gd.ug.frame)
	}
	return gd
}

// Frames returns a recording of frame changes successively computed, if
// recording was enabled for the grid. The frame recording can be used to watch
// a replay of the application's session.
func (gd Grid) Frames() []Frame {
	return gd.ug.frames
}

// ClearCache clears internal cache buffers, forcing a complete redraw of the
// screen with the next Draw call, even for cells that did not change. This can
// be used in the case the physical display and the internal model are not in
// sync: for example after a resize, or after a change of the GetImage function
// of the driver (on the fly change of the tileset).
func (gd Grid) ClearCache() {
	for i := 0; i < len(gd.ug.cellBackBuffer); i++ {
		gd.ug.cellBackBuffer[i] = Cell{}
	}
}
