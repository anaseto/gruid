// Package gruid provides a model for building grid-based applications. The
// interface abstracts rendering and input for different platforms. The package
// provides drivers for terminal apps (driver/tcell), native graphical apps
// (driver/tk) and browser apps (driver/js).
//
// The package uses an architecture of updating a model in response to messages
// strongly inspired from the bubbletea module for building terminal apps (see
// github.com/charmbracelet/bubbletea), which in turn is based on the Elm
// Architecture (https://guide.elm-lang.org/architecture/).
package gruid

import (
	"fmt"
	"log"
	"runtime/debug"
)

// App represents a message and model-driven application with a grid-based user
// interface.
type App struct {
	// CatchPanics ensures that Close is called on the driver before ending
	// the Start loop. When a panic occurs, it will be recovered, the stack
	// trace will be printed and an error will be returned. It defaults to
	// true.
	CatchPanics bool

	driver Driver
	model  Model
}

// AppConfig contains the configuration for creating a new App.
type AppConfig struct {
	Model  Model  // application state
	Driver Driver // input and rendering driver
}

// NewApp creates a new App.
func NewApp(cfg AppConfig) *App {
	return &App{
		model:       cfg.Model,
		driver:      cfg.Driver,
		CatchPanics: true,
	}
}

// Start initializes the program and runs the application's main loop.
func (app *App) Start() error {
	var err error
	var (
		cmds = make(chan Cmd)
		msgs = make(chan Msg)
		done = make(chan struct{})
	)

	// driver initialization
	err = app.driver.Init()
	if err != nil {
		return err
	}
	if app.CatchPanics {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
				app.driver.Close()
				log.Printf("Caught panic: %v\nStack Trace:\n", r)
				debug.PrintStack()
			} else {
				app.driver.Close()
			}
		}()
	} else {
		defer func() {
			app.driver.Close()
		}()
	}

	// model initialization
	initCmd := app.model.Init()
	if initCmd != nil {
		go func() {
			cmds <- initCmd
		}()
	}

	// first drawing
	app.driver.Flush(app.model.Draw().ComputeFrame())

	// input messages queueing
	go func() {
		for {
			msgs <- app.driver.PollMsg()
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

		// Handle quit message
		if _, ok := msg.(msgQuit); ok {
			close(done)
			return err
		}

		// Process batch commands
		if batchedCmds, ok := msg.(msgBatch); ok {
			for _, cmd := range batchedCmds {
				cmds <- cmd
			}
			continue
		}

		cmd := app.model.Update(msg)                      // run update
		cmds <- cmd                                       // process command (if any)
		app.driver.Flush(app.model.Draw().ComputeFrame()) // send drawings to driver
	}
}

// Msg represents an action and triggers the Update function of the model.
type Msg interface{}

// Cmd is a function that returns a message. Commands returned by Update are
// executed on their own goroutine. You can use them for things like timers and
// IO operations. A nil command acts as a no-op.
type Cmd func() Msg

// Batch peforms a bunch of commands concurrently with no ordering guarantees
// about the results.
func Batch(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return nil
	}
	return func() Msg {
		return msgBatch(cmds)
	}
}

// Model contains the application's state.
type Model interface {
	// Init will be called first by Start. It may return an initial command
	// to perform.
	Init() Cmd

	// Update is called when a message is received. Use it to update the
	// model in response to messages and/or send commands.
	Update(Msg) Cmd

	// Draw is called after Init and then after every Update.  Use this
	// function to draw the UI elements in a grid to be returned.  The
	// returned grid will then automatically be sent to the driver for
	// immediate display.
	Draw() Grid
}

// Driver handles both user input and rendering. When creating an App and using
// the Start main loop, you will not have to call those methods directly.
type Driver interface {
	Init() error

	// Flush sends last frame grid changes to the driver.
	Flush(Grid)

	// PollMsg waits for user input messages.
	PollMsg() Msg

	// Close may execute needed code to finalize the screen and release
	// resources.
	Close()
}
