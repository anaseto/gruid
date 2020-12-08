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
//		grid gruid.Grid // user interface grid
//		// other fields with the state of the application
//	}
//
//	func (m *model) Init() gruid.Effect {
//		// Write your model's initialization.
//	}
//
//	func (m *model) Update(msg gruid.Msg) gruid.Effect {
//		// Update your application's state in response to messages.
//	}
//
//	func (m *model) Draw() gruid.Grid {
//		// Write your rendering into the grid and return it or a grid slice.
//	}
//
//	func main() {
//		gd := gruid.NewGrid(gruid.GridConfig{})
//		m := &model{grid: gd}
//		// Specify a driver among the provided ones.
//		driver := &tcell.Driver{...}
//		app := gruid.NewApp(gruid.AppConfig{
//			Driver: driver,
//			Model: m,
//		})
//		// Start the main loop of the application.
//		if err := app.Start(nil); err != nil {
//			log.Fatal(err)
//		}
//	}
//
// The values of type gruid.Effect returned by Init and Update are optional and
// represent concurrently executed functions that produce messages.  See the
// API documentation for details and usage.
package gruid

import (
	"context"
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

	renderer *renderer
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

// Start initializes the application and runs its main loop. The context
// argument can be used as a means to prematurely cancel the loop. You can
// usually use nil here for client applications.
func (app *App) Start(ctx context.Context) (err error) {
	var (
		effects = make(chan Effect)
		msgs    = make(chan Msg)
		errs    = make(chan error)
	)

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
		go func(ctx context.Context) {
			select {
			case effects <- initCmd:
			case <-ctx.Done():
				return
			}
		}(ctx)
	}

	// initialize renderer
	app.renderer = &renderer{
		driver: app.driver,
		rec:    app.rec,
		fps:    app.fps,
	}
	app.renderer.Init()
	go app.renderer.Listen(ctx)

	// first drawing (buffered, non blocking)
	app.renderer.frames <- app.model.Draw().ComputeFrame()

	// input messages queueing
	go func(ctx context.Context) {
		for {
			msg, err := app.driver.PollMsg()
			if err != nil {
				select {
				case errs <- err:
					return
				case <-ctx.Done():
					return
				}
			}
			if msg == nil {
				// Close has been sent to the driver.
				return
			}
			select {
			case msgs <- msg:
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	// effect processing
	go func(ctx context.Context) {
		for {
			select {
			case eff := <-effects:
				if eff != nil {
					switch eff := eff.(type) {
					case Cmd:
						go func(ctx context.Context, cmd Cmd) {
							select {
							case msgs <- cmd():
							case <-ctx.Done():
							}
						}(ctx, eff)
					case Sub:
						go eff(ctx, msgs)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	// main loop
	for {
		select {
		case <-ctx.Done():
			<-app.renderer.done
			return ctx.Err()
		case err := <-errs:
			cancel()
			<-app.renderer.done
			return err
		case msg := <-msgs:
			if msg == nil {
				continue
			}

			// Handle quit message
			if _, ok := msg.(msgQuit); ok {
				cancel()
				<-app.renderer.done
				return err
			}

			// Process batched effects
			if batchedEffects, ok := msg.(msgBatch); ok {
				for _, eff := range batchedEffects {
					effects <- eff
				}
				continue
			}

			eff := app.model.Update(msg) // run update
			effects <- eff               // process effect (if any)
			frame := app.model.Draw().ComputeFrame()
			if len(frame.Cells) > 0 {
				app.renderer.frames <- frame // send frame with changes to driver
			}
		}
	}
}

// Frames returns the successive frames recorded by the application. It can be
// used for a replay of the session.
func (app *App) Frames() []Frame {
	return app.renderer.framerec
}

// Msg represents an action and triggers the Update function of the model. Note
// that nil messages are discarded and do not trigger Update.
type Msg interface{}

// Effect is an interface value for representing either a Cmd or Sub type.
// They generally represent IO operations, either producing a single message or
// several. See Cmd and Sub documentation for details.
type Effect interface {
	ImplementsEffect()
}

// Cmd is a function that returns a message. Commands returned by Update are
// executed on their own goroutine. You can use them for things like timers and
// IO operations. A nil command is discarded and does nothing.
type Cmd func() Msg

// Sub is similar to Cmd, but instead of returning a message, it sends messages
// to a channel. Subscriptions should only be used for long running processes
// where more than one message will be produced, for example to send messages
// delivered by a time.Ticker, or to report messages from listening on a
// socket. The function should handle the context and terminate as appropiate.
type Sub func(context.Context, chan<- Msg)

// ImplementsEffect makes Cmd satisfy Effect.
func (cmd Cmd) ImplementsEffect() {}

// ImplementsEffect makes Sub satisfy Effect.
func (sub Sub) ImplementsEffect() {}

// Batch peforms a bunch of effects concurrently with no ordering guarantees
// about the potential results.
func Batch(effs ...Effect) Effect {
	if len(effs) == 0 {
		return nil
	}
	return Cmd(func() Msg {
		return msgBatch(effs)
	})
}

// Model contains the application's state.
type Model interface {
	// Init will be called first by Start. It may return an initial command
	// or subscription to perform.
	Init() Effect

	// Update is called when a message is received. Use it to update the
	// model in response to messages and/or send commands or subscriptions.
	Update(Msg) Effect

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

	// PollMsg waits for user input messages. It returns nil after Close
	// has been sent to the driver. It returns an error in case the driver
	// input loop suffered a non recoverable error.
	PollMsg() (Msg, error)

	// Close may execute needed code to finalize the screen and release
	// resources.
	Close()
}
