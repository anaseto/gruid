package gorltk

import (
	"time"
)

// CustomStyle can be used to add custom styling information. It can for
// example be used to apply specific terminal attributes (with GetStyle),
// or use special images (with GetImage), when appropiate.
type CustomStyle int

const DefaultStyle CustomStyle = 0

// Color is a generic value for representing colors. Those have to be mapped to
// the concrete colors for each driver, as appropiate.
type Color uint

// GridCell contains all the styling information to represent a cell in the
// grid.
type GridCell struct {
	Fg    Color
	Bg    Color
	Rune  rune
	Style CustomStyle
}

// Grid represents the game grid that is used to draw to the screen.
type Grid struct {
	driver         Driver
	width          int        // XXX maybe unexport?
	height         int        // XXX maybe unexport?
	cellBuffer     []GridCell // TODO: do not export
	cellBackBuffer []GridCell
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
	Cells []FrameCell // cells that changed from previous frame
	Time  time.Time   // time of frame drawing: used for replay
}

type FrameCell struct {
	Cell GridCell
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
	gd.resize(cfg.Width, cfg.Height)
	gd.recording = cfg.Recording
	return gd
}

func (gd *Grid) Size() (int, int) {
	return gd.width, gd.height
}

func (gd *Grid) resize(w, h int) {
	gd.width = w
	gd.height = h
	if len(gd.cellBuffer) != gd.height*gd.width {
		gd.cellBuffer = make([]GridCell, gd.height*gd.width)
	}
}

func (gd *Grid) SetCell(pos Position, gc GridCell) {
	i := gd.GetIndex(pos)
	if i >= len(gd.cellBuffer) || i < 0 {
		return
	}
	gd.cellBuffer[i] = gc
}

func (gd *Grid) GetIndex(pos Position) int {
	// TODO: unexport
	return pos.Y*gd.width + pos.X
}

func (gd *Grid) GetPos(i int) Position {
	// TODO: unexport
	return Position{X: i - (i/gd.width)*gd.width, Y: i / gd.width}
}

func (gd *Grid) Frame() Frame {
	return gd.frame
}

// Draw draws computes next frame changes. If recording is activated the frame
// changes are recorded, and can be retrieved later by calling Frames().
func (gd *Grid) Draw() {
	if len(gd.cellBackBuffer) != len(gd.cellBuffer) {
		gd.cellBackBuffer = make([]GridCell, len(gd.cellBuffer))
	}
	gd.frame = Frame{Time: time.Now()}
	for i := 0; i < len(gd.cellBuffer); i++ {
		if gd.cellBuffer[i] == gd.cellBackBuffer[i] {
			continue
		}
		c := gd.cellBuffer[i]
		pos := gd.GetPos(i)
		cdraw := FrameCell{Cell: c, Pos: pos}
		gd.frame.Cells = append(gd.frame.Cells, cdraw)
		gd.cellBackBuffer[i] = c
	}
	if gd.recording {
		gd.frames = append(gd.frames, gd.frame)
	}
}

// Frames returns a recording of frames as produced by successive Draw() call,
// if recording was enabled for the grid. The frame recording can be used to
// watch a replay of the game.  Note that each frame contains only cells that
// changed since the previous one.
func (gd *Grid) Frames() []Frame {
	return gd.frames
}

func (gd *Grid) ClearCache() {
	for i := 0; i < len(gd.cellBackBuffer); i++ {
		gd.cellBackBuffer[i] = GridCell{}
	}
}
