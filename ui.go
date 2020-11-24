package gorltk

import (
	"time"
)

type GridDrawer interface {
	// FrameCells returns drawing instructions for next frame.
	FrameCells() []FrameCell
	// Size returns the (width, height) of the grid in cells.
	Size() (int, int)
}

type Driver interface {
	Init() error // XXX keep it in the interface ?
	// Flush draws sends current drawing frame to the driver.
	Flush(GridDrawer)
	// PollEvent waits for user input and returns it.
	PollEvent() Event
	// Interrupt sends an EventInterrupt event.
	Interrupt()
	// Close executes optional code before closing the UI.
	Close() // XXX keep it in the interface ?
	// ClearCache clears any cache from cell styles to tiles.
	ClearCache()
}

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
	R     rune
	Style CustomStyle
}

type EventType int

const (
	EventKeyDown EventType = iota
	EventMouseDown
	EventMouseMove
	EventInterrupt
)

type MouseButton int

const (
	ButtonMain      MouseButton = iota // left button
	ButtonAuxiliary                    // middle button
	ButtonSecondary                    // right button
)

// Event represents information from user input.
type Event struct {
	Type     EventType   // any of the Event* constants
	Button   MouseButton // mouse button number in EventMouseDown event
	Key      string      // name of the key in EventKeyDown event
	MousePos Position    // mouse position in EventMouseDown or EventMouseMove event
	Shift    bool        // whether shift key was pressed during input
}

type Grid struct {
	Width          int        // XXX maybe unexport?
	Height         int        // XXX maybe unexport?
	CellBuffer     []GridCell // TODO: do not export
	cellBackBuffer []GridCell
	FrameLog       []Frame
}

type Frame struct {
	Cells []FrameCell // cells that changed from previous frame
	Time  time.Time   // time of frame drawing: used for replay
}

type FrameCell struct {
	Cell GridCell
	Pos  Position
}

func (gd *Grid) Size() (int, int) {
	return gd.Width, gd.Height
}

func (gd *Grid) SetGenCell(fc FrameCell) {
	i := gd.GetIndex(fc.Pos)
	if i >= gd.Height*gd.Width {
		return
	}
	gd.CellBuffer[i] = fc.Cell
}

func (gd *Grid) GetIndex(pos Position) int {
	return pos.Y*gd.Width + pos.X
}

func (gd *Grid) GetPos(i int) Position {
	return Position{X: i - (i/gd.Width)*gd.Width, Y: i / gd.Width}
}

func (gd *Grid) FrameCells() []FrameCell {
	if len(gd.FrameLog) <= 0 {
		return nil
	}
	return gd.FrameLog[len(gd.FrameLog)-1].Cells
}

func (gd *Grid) ComputeNextFrame() {
	if len(gd.cellBackBuffer) != len(gd.CellBuffer) {
		gd.cellBackBuffer = make([]GridCell, len(gd.CellBuffer))
	}
	gd.FrameLog = append(gd.FrameLog, Frame{Time: time.Now()})
	for i := 0; i < len(gd.CellBuffer); i++ {
		if gd.CellBuffer[i] == gd.cellBackBuffer[i] {
			continue
		}
		c := gd.CellBuffer[i]
		pos := gd.GetPos(i)
		cdraw := FrameCell{Cell: c, Pos: pos}
		last := len(gd.FrameLog) - 1
		gd.FrameLog[last].Cells = append(gd.FrameLog[last].Cells, cdraw)
		gd.cellBackBuffer[i] = c
	}
}

func (gd *Grid) Init() {
	if len(gd.CellBuffer) == 0 {
		gd.CellBuffer = make([]GridCell, gd.Height*gd.Width)
	} else if len(gd.CellBuffer) != gd.Height*gd.Width {
		gd.CellBuffer = make([]GridCell, gd.Height*gd.Width)
	}
}

func (gd *Grid) ClearCache() {
	for i := 0; i < len(gd.cellBackBuffer); i++ {
		gd.cellBackBuffer[i] = GridCell{}
	}
}
