// Copyright (c) 2020 Yon <anaseto@bardinflor.perso.aquilenet.fr>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
//
// ----
//
// Some code from the core gruid package relative to models and commands in ui.go
// is strongly inspired from github.com/charmbracelet/bubbletea, which uses the
// following license:
//
// MIT License
//
// Copyright (c) 2020 Charmbracelet, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package gruid provides a model for building grid-based applications. The
// interface abstracts rendering and input for different platforms. The module
// provides drivers for terminal apps (driver/tcell), native graphical apps
// (driver/sdl) and browser apps (driver/js).
//
// The package uses an architecture of updating a model in response to messages
// strongly inspired from the bubbletea module for building terminal apps
// (https://github.com/charmbracelet/bubbletea), which in turn is based on the
// Elm Architecture (https://guide.elm-lang.org/architecture/).
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
//	func (m *model) Update(msg gruid.Msg) gruid.Effect {
//		// Update your application's state in response to messages.
//	}
//
//	func (m *model) Draw() gruid.Grid {
//		// Write your rendering into the grid and return it.
//	}
//
//	func main() {
//		gd := gruid.NewGrid(80, 24)
//		m := &model{grid: gd, ...}
//		// Specify a driver among the provided ones.
//		driver := tcell.NewDriver(...)
//		app := gruid.NewApp(gruid.AppConfig{
//			Driver: driver,
//			Model: m,
//		})
//		// Start the main loop of the application.
//		if err := app.Start(context.Background()); err != nil {
//			log.Fatal(err)
//		}
//	}
//
// The values of type gruid.Effect returned by Update are optional and
// represent concurrently executed functions that produce messages.  See the
// relevant types documentation for details and usage.
package gruid

import (
	"context"
	"fmt"
	"io"
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
	enc    *frameEncoder
	logger *log.Logger

	grid  Grid
	frame Frame
}

// AppConfig contains the configuration options for creating a new App.
type AppConfig struct {
	Model  Model  // application state
	Driver Driver // input and rendering driver

	// FrameWriter is an optional io.Writer for recording frames. They can
	// be decoded after a successful Start session with a FrameDecoder. If
	// nil, no frame recording will be done. It is your responsibility to
	// call Close on the Writer after Start returns.
	FrameWriter io.Writer

	// Logger is optional and is used to log non-fatal IO errors.
	Logger *log.Logger
}

// NewApp creates a new App with the given configuration options.
func NewApp(cfg AppConfig) *App {
	app := &App{
		model:       cfg.Model,
		driver:      cfg.Driver,
		logger:      cfg.Logger,
		CatchPanics: true,
	}
	if cfg.FrameWriter != nil {
		app.enc = newFrameEncoder(cfg.FrameWriter)
	}
	return app
}

// Start initializes the application and runs its main loop. The context
// argument can be used as a means to prematurely cancel the loop. You can
// usually use an empty context here.
func (app *App) Start(ctx context.Context) (err error) {
	var (
		effects  = make(chan Effect, 4)
		msgs     = make(chan Msg, 4)
		errs     = make(chan error)    // for driver input errors
		polldone = make(chan struct{}) // PollMsgs subscription finished
	)

	// frame encoder finalization
	defer func() {
		if app.enc != nil {
			nerr := app.enc.gzw.Close()
			if err == nil {
				err = nerr
			} else if app.logger != nil {
				app.logger.Printf("error closing gzip encoder: %v", err)
			}
		}
	}()

	// driver and context initialization
	err = app.driver.Init()
	if err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	if app.CatchPanics {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
				cancel()
				<-polldone
				app.driver.Close()
				log.Printf("Caught panic: %v\nStack Trace:\n", r)
				debug.PrintStack()
			} else {
				<-polldone
				app.driver.Close()
			}
		}()
	} else {
		defer func() {
			<-polldone
			app.driver.Close()
		}()
	}
	defer cancel()

	// initialization message (non-blocking, buffered)
	msgs <- MsgInit{}

	// input messages queueing
	go func(ctx context.Context) {
		defer func() {
			close(polldone)
		}()
		err := app.driver.PollMsgs(ctx, msgs)
		if err != nil {
			select {
			case errs <- err:
			case <-ctx.Done():
			}
		}
	}(ctx)

	// effect processing
	go processEffects(ctx, msgs, effects)

	// Update on message then Draw main loop
	for {
		select {
		case <-ctx.Done():
			return err
		case err := <-errs:
			cancel()
			return err
		case msg := <-msgs:
			if msg == nil {
				continue
			}

			// Handle quit message
			if _, ok := msg.(msgEnd); ok {
				cancel()
				return err
			}

			// Process batched effects
			if batchedEffects, ok := msg.(msgBatch); ok {
				for _, eff := range batchedEffects {
					select {
					case effects <- eff:
					case <-ctx.Done():
						break
					}
				}
				continue
			}

			// force redraw on screen message
			_, exposed := msg.(MsgScreen)

			eff := app.model.Update(msg)
			if eff != nil {
				select {
				case effects <- eff: // process effect (if any)
				case <-ctx.Done():
					continue
				}
			}

			gd := app.model.Draw()
			frame := app.computeFrame(gd, exposed)
			if len(frame.Cells) > 0 {
				app.flush(frame)
			}
		}
	}
}

func (app *App) flush(frame Frame) {
	app.driver.Flush(frame)
	if app.enc != nil {
		err := app.enc.encode(frame)
		if err != nil && app.logger != nil {
			app.logger.Printf("frame encoding: %v", err)
		}
	}
}

func processEffects(ctx context.Context, msgs chan Msg, effects chan Effect) {
	for {
		select {
		case eff := <-effects:
			switch eff := eff.(type) {
			case Cmd:
				if eff != nil {
					go func(ctx context.Context, cmd Cmd) {
						select {
						case msgs <- cmd():
						case <-ctx.Done():
						}
					}(ctx, eff)
				}
			case Sub:
				if eff != nil {
					go eff(ctx, msgs)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// Model contains the application's state.
type Model interface {
	// Update is called when a message is received. Use it to update your
	// model in response to messages and/or send commands or subscriptions.
	// It is always called the first time with a MsgInit message.
	Update(Msg) Effect

	// Draw is called after every Update. Use this function to draw the UI
	// elements in a grid to be returned. If only parts of the grid are to
	// be updated, you can return a smaller grid slice, or an empty grid
	// slice to skip any drawing work. Note that the contents of the grid
	// slice are then compared to the previous state at the same bounds,
	// and only the changes are sent to the driver anyway.
	Draw() Grid
}

// Driver handles both user input and rendering. When creating an App and using
// the Start main loop, you will not have to call those methods directly. You
// may reuse the same driver for another application after the current
// application's Start loop ends.
type Driver interface {
	// Init initializes the driver, so that you can then call its other
	// methods.
	Init() error

	// PollMsgs is a subscription for input messages. It returns an error
	// in case the driver input loop suffered a non recoverable error. It
	// should handle cancellation of the passed context and return as
	// appropriate.
	PollMsgs(context.Context, chan<- Msg) error

	// Flush sends grid's last frame changes to the driver.
	Flush(Frame)

	// Close may execute needed code to finalize the screen and release
	// resources. Redundant Close() calls are ignored. After Close() it is
	// possible to call Init() again.
	Close()
}

// Msg represents an action and triggers the Update function of the model. Note
// that nil messages are discarded and do not trigger Update.
type Msg interface{}

// Effect is an interface type for representing either command or subscription
// functions.  Those functions generally represent IO operations, either
// producing a single message or several. They are executed on their own
// goroutine after being returned by the Update method of the model. A nil
// effect is discarded and does nothing.
//
// The types Cmd and Sub implement the Effect interface. See their respective
// documentation for specific usage details.
type Effect interface {
	implementsEffect()
}

// Cmd is an Effect that returns a message. Commands returned by Update are
// executed on their own goroutine. You can use them for things like single
// event timers and short-lived IO operations with a single result. A nil
// command is discarded and does nothing.
//
// Cmd implements the Effect interface.
type Cmd func() Msg

// Sub is similar to Cmd, but instead of returning a message, it sends messages
// to a channel. Subscriptions should only be used for long running functions
// where more than one message will be produced, for example to send messages
// delivered by a time.Ticker, or to report messages from listening on a
// socket. The function should handle the context and terminate as appropriate.
//
// Sub implements the Effect interface.
type Sub func(context.Context, chan<- Msg)

// implementsEffect makes Cmd satisfy Effect interface.
func (cmd Cmd) implementsEffect() {}

// implementsEffect makes Sub satisfy Effect interface.
func (sub Sub) implementsEffect() {}

// End returns a special command that signals the application to end its Start
// loop. Note that the application does not wait for pending effects to
// complete before exiting the Start loop, so you may have to wait for any of
// those commands messages before using End.
func End() Cmd {
	return func() Msg {
		return msgEnd{}
	}
}

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
