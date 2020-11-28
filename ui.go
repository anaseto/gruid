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

import (
	"fmt"
	"log"
	"runtime/debug"
)

// Game is the game user interface.
type Game struct {
	// CatchPanics ensures that Close is called on the driver before ending
	// the game Start loop. When a panic occurs, it will be recovered, the
	// stack trace will be printed and an error will be returned. It
	// defaults to true.
	CatchPanics bool

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
		model:       cfg.Model,
		grid:        cfg.Grid,
		driver:      cfg.Driver,
		CatchPanics: true,
	}
}

// Start initializes the program and runs the main game loop.
func (g *Game) Start() (err error) {
	var (
		cmds = make(chan Cmd)
		msgs = make(chan Msg)
		done = make(chan struct{})
	)

	// driver initialization
	err = g.driver.Init()
	if err != nil {
		return err
	}
	if g.CatchPanics {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
				g.driver.Close()
				log.Printf("Caught panic: %v\nStack Trace:\n", r)
				debug.PrintStack()
			} else {
				g.driver.Close()
			}
		}()
	}

	// model initialization
	initCmd := g.model.Init()
	if initCmd != nil {
		go func() {
			cmds <- initCmd
		}()
	}

	// first drawing
	g.grid.Draw()
	g.driver.Flush(g.grid)

	// input messages queueing
	go func() {
		for {
			msg, ok := g.driver.PollMsg()
			if ok {
				msgs <- msg
			}
		}
	}()

	// command processing
	go func() {
		for {
			select {
			case <-done:
				return
			case cmd := <-cmds:
				if cmd != nil {
					go func() {
						msgs <- cmd()
					}()
				}
			}
		}
	}()

	// main loop
	for {
		msg := <-msgs
		if _, ok := msg.(msgQuit); ok {
			close(done)
			return nil
		}
		cmd := g.model.Update(msg)
		cmds <- cmd
		g.model.Draw(g.grid)
		g.grid.Draw()
		g.driver.Flush(g.grid)
	}
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

	// Draw is called after Init and then after every Update.  Use this
	// function to draw the UI elements in the grid. The grid resulting
	// changes will then automatically be sent to the driver.
	Draw(*Grid)
}

// Driver handles both user input and rendering. When using the message
// architecture you will not have to call those methods directly.
type Driver interface {
	Init() error
	// Flush sends last frame grid changes to the driver.
	Flush(*Grid)
	// PollMsg waits for user input. It returns (msg, true) on user input,
	// and (nil, false) if the waiting is terminated by calling Interrupt.
	PollMsg() (Msg, bool)
	// Interrupt terminates prematurely a PollMsg call.
	Interrupt()
	// Close executes optional code before closing the UI.
	Close()
	// ClearCache clears the cache from cell styles to tiles, if any.
	ClearCache()
}
