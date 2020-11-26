// Package gorltk provides a model for building roguelike games that use a
// grid-like interface. The interface abstracts rendering and input for
// different platforms. The package provides drivers for terminal apps
// (driver/tcell), native graphical apps (driver/tk) and browser apps
// (driver/js).
//
// The package uses an architecture of message passing inspired from the
// bubbletea module for building terminal apps, which in turn is based on the
// Elm Architecture.
package gorltk

import (
	"time"
)

// Game is the game user interface.
type Game struct {
	driver Driver
	model  Model
	grid   Grid
}

// GameConfig contains the configuration for creating a new Game.
type GameConfig struct {
	Model  Model
	Driver Driver
	Grid   Grid
}

// NewGame creates a new Game.
func NewGame(cfg GameConfig) *Game {
	return &Game{
		driver: cfg.Driver,
		model:  cfg.Model,
		grid:   cfg.Grid,
	}
}

// TODO
func (g *Game) Start() error {
	// TODO
	return nil
}

// Msg represents an action and triggers the Update function of the model.
type Msg interface{}

// Cmd is a function that returns a message. Commands returned by Update are
// executed on their own goroutine. You can use them for things like timers and
// IO operations. A nil command acts as a no-op.
type Cmd func() Msg

// Model contains the game state.
type Model interface {
	// Init will be called to initialize the model. It may return an
	// initial command to perform.
	Init() Cmd
	// Update is called when a message is received. Use it to update the
	// model in response to messages and/or send commands.
	Update(Msg) Cmd
	// Draw is called after Init and then after every Update and sends the
	// UI grid state to the driver. Use this function to draw the UI
	// elements in the grid.
	Draw(*Grid)
}

// Driver handles both user input and rendering. When using the message
// architecture you will not have to call those methods directly.
type Driver interface {
	Init() error // XXX keep it in the interface ?
	// Flush sends last frame grid changes to the driver.
	Flush(*Grid)
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
	Rune  rune
	Style CustomStyle
}

type MouseButton int

const (
	ButtonMain      MouseButton = iota // left button
	ButtonAuxiliary                    // middle button
	ButtonSecondary                    // right button
)

// Event is an interface for passing events around.
type Event interface {
	// When reports the time when the event was generated.
	When() time.Time
}

// EventKeyDown represents a key press.
type EventKeyDown struct {
	Key   Key  // name of the key in EventKeyDown event
	shift bool //
	Time  time.Time
}

// When returns the time when this event was generated.
func (ev EventKeyDown) When() time.Time {
	return ev.Time
}

// ShiftKey indicates if the shift key was pressed when the event occured. This
// may be not supported equally well accross all platforms.
func (ev EventKeyDown) ShiftKey() bool {
	return ev.shift
}

// EventMouseDown represents a mouse click.
type EventMouseDown struct {
	Button   MouseButton // which button was pressed
	MousePos Position    // mouse position in the grid
	Time     time.Time
}

// When returns the time when this event was generated.
func (ev EventMouseDown) When() time.Time {
	return ev.Time
}

// EventMouseMove represents a mouse motion.
type EventMouseMove struct {
	MousePos Position // mouse position in the grid
	Time     time.Time
}

// When returns the time when this event was generated.
func (ev EventMouseMove) When() time.Time {
	return ev.Time
}

// EventInterrupt represents a wakeup. It can be used to end prematurely a
// PollEvent call, for example to signal the end of an animation.
type EventInterrupt struct {
	Time time.Time
}

// When returns the time when this event was generated.
func (ev EventInterrupt) When() time.Time {
	return ev.Time
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

// GridCfg is used to configure a new Grid with NewGrid.
type GridCfg struct {
	Driver    Driver // drawing instructions Driver
	Width     int    // width in cells
	Height    int    // height in cells
	Recording bool   // whether to record frames to enable replay
}

type Frame struct {
	Cells []FrameCell // cells that changed from previous frame
	Time  time.Time   // time of frame drawing: used for replay
}

type FrameCell struct {
	Cell GridCell
	Pos  Position
}

func NewGrid(cfg GridCfg) *Grid {
	gd := &Grid{}
	if cfg.Driver == nil {
		panic("cfg.Driver is nil")
	}
	gd.driver = cfg.Driver
	if cfg.Height <= 0 {
		panic("cfg.Height must be positive")
	}
	if cfg.Width <= 0 {
		panic("cfg.Width must be positive")
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
	if i >= gd.height*gd.width {
		return
	}
	gd.cellBuffer[i] = gc
}

func (gd *Grid) GetIndex(pos Position) int {
	return pos.Y*gd.width + pos.X
}

func (gd *Grid) GetPos(i int) Position {
	return Position{X: i - (i/gd.width)*gd.width, Y: i / gd.width}
}

func (gd *Grid) Frame() Frame {
	return gd.frame
}

// Draw draws computes next frame changes and sends them to the Driver for
// immediate display. If recording is activated the frame changes are recorded,
// and can be retrieved by calling Frames().
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
	gd.driver.Flush(gd)
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
