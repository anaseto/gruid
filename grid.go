package gorltk

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

// Grid represents the game grid that is used to draw to the screen.
type Grid struct {
	width          int
	height         int
	cellBuffer     []Cell
	cellBackBuffer []Cell
	frame          Frame
	frames         []Frame
	recording      bool
}

// GridConfig is used to configure a new Grid with NewGrid.
type GridConfig struct {
	Width     int  // width in cells: default is 80
	Height    int  // height in cells: default is 24
	Recording bool // whether to record frames to enable replay
}

type Frame struct {
	Cells  []FrameCell // cells that changed from previous frame
	Time   time.Time   // time of frame drawing: used for replay
	Width  int         // width of the grid when the frame was produced
	Height int         // height of the grid when the frame was produced
}

type FrameCell struct {
	Cell Cell
	Pos  Position
}

func NewGrid(cfg GridConfig) *Grid {
	gd := &Grid{}
	if cfg.Height <= 0 {
		cfg.Height = 24
	}
	if cfg.Width <= 0 {
		cfg.Width = 80
	}
	gd.Resize(cfg.Width, cfg.Height)
	gd.recording = cfg.Recording
	return gd
}

func (gd *Grid) Size() (int, int) {
	return gd.width, gd.height
}

// Resize can be used to resize the grid. If dimensions changed, it clears the
// grid.
//
// Note that this only modifies the size of the grid, which may be different
// than the window screen size.
func (gd *Grid) Resize(w, h int) {
	if gd.width == w && gd.height == h {
		return
	}
	newBuf := make([]Cell, w*h)
	gd.width = w
	gd.height = h
	gd.cellBuffer = newBuf
}

// SetCell draws cell content and styling at a given position in the grid. It
// is a no-op if the given position is outside the grid.
func (gd *Grid) SetCell(pos Position, c Cell) {
	i := gd.getIdx(pos)
	if i >= len(gd.cellBuffer) || i < 0 {
		return
	}
	gd.cellBuffer[i] = c
}

func (gd *Grid) getIdx(pos Position) int {
	return pos.Y*gd.width + pos.X
}

func (gd *Grid) getPos(i int) Position {
	return Position{X: i - (i/gd.width)*gd.width, Y: i / gd.width}
}

// Frame returns the drawing instructions produced by last Draw call.
//
// This function may be used to implement new drivers. You should normally not
// call it by hand when implementing a game.
func (gd *Grid) Frame() Frame {
	return gd.frame
}

// Draw computes next frame changes. If recording is activated the frame
// changes are recorded, and can be retrieved later by calling Frames().
//
// This function is automatically called after each Draw of the Model. You
// should normally not call it by hand when implementing a game using a Model.
func (gd *Grid) Draw() {
	if len(gd.cellBackBuffer) != len(gd.cellBuffer) {
		gd.cellBackBuffer = make([]Cell, len(gd.cellBuffer))
	}
	gd.frame = Frame{Time: time.Now(), Width: gd.width, Height: gd.height}
	for i := 0; i < len(gd.cellBuffer); i++ {
		if gd.cellBuffer[i] == gd.cellBackBuffer[i] {
			continue
		}
		c := gd.cellBuffer[i]
		pos := gd.getPos(i)
		cdraw := FrameCell{Cell: c, Pos: pos}
		gd.frame.Cells = append(gd.frame.Cells, cdraw)
		gd.cellBackBuffer[i] = c
	}
	if gd.recording {
		gd.frames = append(gd.frames, gd.frame)
	}
}

// Frames returns a recording of frame changes as produced by successive Draw()
// calls, if recording was enabled for the grid. The frame recording can be
// used to watch a replay of the game.
func (gd *Grid) Frames() []Frame {
	return gd.frames
}

func (gd *Grid) ClearCache() {
	for i := 0; i < len(gd.cellBackBuffer); i++ {
		gd.cellBackBuffer[i] = Cell{}
	}
}
