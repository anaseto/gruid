// Package gorltk provides a model for building roguelike games that use a
// grid-like interface. The interface abstracts rendering and input for
// different platforms. The package provides drivers for terminal apps
// (driver/tcell), native graphical apps (driver/tk) and browser apps
// (driver/js).
//
// The package uses an architecture of updating a model in response to messages
// inspired from the bubbletea module for building terminal apps, which in turn
// is based on the Elm Architecture.
package gorltk

// Game is the game user interface.
type Game struct {
	driver Driver
	model  Model
	grid   *Grid
}

// GameConfig contains the configuration for creating a new Game.
type GameConfig struct {
	Model  Model  // game state
	Grid   *Grid  // game UI grid
	Driver Driver // input and rendering driver
}

// NewGame creates a new Game.
func NewGame(cfg GameConfig) *Game {
	return &Game{
		model:  cfg.Model,
		grid:   cfg.Grid,
		driver: cfg.Driver,
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
	Init() error
	// Flush sends last frame grid changes to the driver.
	Flush(*Grid)
	// PollMsg waits for user input and returns it.
	PollMsg() Msg
	// Interrupt sends an EventInterrupt event.
	Interrupt()
	// Close executes optional code before closing the UI.
	Close()
	// ClearCache clears any cache from cell styles to tiles.
	ClearCache()
}
