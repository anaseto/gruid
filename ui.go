// Package gruid provides a model for building grid-based applications. The
// interface abstracts rendering and input for different platforms. The package
// provides drivers for terminal apps (driver/tcell), native graphical apps
// (driver/tk) and browser apps (driver/js).
//
// The package uses an architecture of updating a model in response to messages
// strongly inspired from the bubbletea module for building terminal apps (see
// github.com/charmbracelet/bubbletea), which in turn is based on the Elm
// Architecture (https://guide.elm-lang.org/architecture/).
//
// The typical usage looks like this:
//
//	// model implements gruid.Model interface and represents the
//	// application's state.
//	type model struct {
//		grid: gruid.Grid // user interface grid
//		// other fields with the state of the application
//	}
//
//	func (m *model) Init() gruid.Cmd {
//		// Write your model's initialization.
//	}
//
//	func (m *model) Update(msg gruid.Msg) gruid.Cmd {
//		// Update your application's state in response to messages.
//	}
//
//	func (m *model) Draw() gruid.Grid {
//		// Write your rendering into the grid and return it or a grid slice.
//	}
//
//	func main() {
//		grid := gruid.NewGrid(gruid.GridConfig{})
//		m := &model{grid: grid}
//		// Specify a driver among the provided ones.
//		driver := &tcell.Driver{...}
//		app := gruid.NewApp(gruid.AppConfig{
//			Driver: driver,
//			Model: m,
//		})
//		// Start the main loop of the application.
//		if err := app.Start(); err != nil {
//			log.Fatal(err)
//		}
//	}
//
// The values of type gruid.Cmd returned by Init and Update are optional and
// represent concurrently executed functions that produce a message.  See the
// API documentation for details and usage.
package gruid

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
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
	rec    bool
	fps    time.Duration

	renderer renderer
}

// AppConfig contains the configuration for creating a new App.
type AppConfig struct {
	Model  Model  // application state
	Driver Driver // input and rendering driver

	// FrameRecording instructs the application to record the frames to
	// enable replay. They can be retrieved after a successful Start()
	// session with Frames().
	FrameRecording bool

	// FPS specifies the maximum number of frames per second. Should be a
	// value between 60 and 240. Default is 60.
	FPS time.Duration
}

// NewApp creates a new App.
func NewApp(cfg AppConfig) *App {
	return &App{
		model:       cfg.Model,
		driver:      cfg.Driver,
		rec:         cfg.FrameRecording,
		fps:         cfg.FPS,
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

	// initialize renderer
	r := renderer{
		driver: app.driver,
		rec:    app.rec,
		fps:    app.fps,
	}
	r.Init()
	go r.Listen()

	// first drawing
	r.frames <- app.model.Draw().ComputeFrame()

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
			r.Stop()
			<-r.done
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

		cmd := app.model.Update(msg) // run update
		cmds <- cmd                  // process command (if any)
		frame := app.model.Draw().ComputeFrame()
		if len(frame.Cells) > 0 {
			r.frames <- frame // send frame with changes to driver
		}
	}
}

// Frames returns the successive frames recorded by the application. It can be
// used for a replay of the session.
func (app *App) Frames() []Frame {
	return app.renderer.framerec
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
	Flush(Frame)

	// PollMsg waits for user input messages.
	PollMsg() Msg

	// Close may execute needed code to finalize the screen and release
	// resources.
	Close()
}
