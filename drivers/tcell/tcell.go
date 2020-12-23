// Package tcell provides a gruid Driver for making terminal apps.
package tcell

import (
	"context"
	"errors"

	"github.com/anaseto/gruid"
	"github.com/gdamore/tcell/v2"
)

// StyleManager allows for retrieving of styling information.
type StyleManager interface {
	// GetAttributes returns a mask of text attributes for a given cell
	// style.
	GetStyle(gruid.Style) tcell.Style
}

// Driver implements gruid.Driver using the tcell terminal library.
type Driver struct {
	sm        StyleManager
	screen    tcell.Screen
	mouse     bool
	mousedrag bool
	mousePos  gruid.Point
	init      bool
	noQuit    bool
}

// Config contains configurations options for the driver.
type Config struct {
	StyleManager StyleManager // for cell styling (required)
	DisableMouse bool         // disable mouse-related messages
}

// NewDriver returns a new driver with given configuration options.
func NewDriver(cfg Config) *Driver {
	return &Driver{sm: cfg.StyleManager, mouse: !cfg.DisableMouse}
}

// PreventQuit will make next call to Close keep the same tcell screen. It can
// be used to chain two applications with the same screen. It is then your
// reponsibility to either run another application or call Close manually to
// properly quit.
func (dr *Driver) PreventQuit() {
	dr.noQuit = true
}

// Init implements gruid.Driver.Init. It initializes a screen using the tcell
// terminal library.
func (dr *Driver) Init() error {
	if dr.sm == nil {
		return errors.New("no style manager provided")
	}
	if !dr.init {
		screen, err := tcell.NewScreen()
		dr.screen = screen
		if err != nil {
			return err
		}
		err = dr.screen.Init()
		if err != nil {
			return err
		}
		dr.screen.SetStyle(tcell.StyleDefault)
		if dr.mouse {
			dr.screen.EnableMouse()
		} else {
			dr.screen.DisableMouse()
		}
		dr.screen.HideCursor()
	}

	// try to send initial size
	w, h := dr.screen.Size()
	dr.screen.PostEvent(tcell.NewEventResize(w, h))

	dr.init = true
	return nil
}

// PollMsgs implements gruid.Driver.PollMsgs. It does not report KP_5 keypad
// key when numlock is off.
func (dr *Driver) PollMsgs(ctx context.Context, msgs chan<- gruid.Msg) error {
	go func(ctx context.Context) {
		<-ctx.Done()
		n := 0
		err := dr.screen.PostEvent(tcell.NewEventInterrupt(0))
		for err != nil && n < 10 {
			// should not happen in practice
			n++
			err = dr.screen.PostEvent(tcell.NewEventInterrupt(0))
		}
	}(ctx)
	send := func(msg gruid.Msg) {
		select {
		case msgs <- msg:
		case <-ctx.Done():
		}
	}
	for {
		ev := dr.screen.PollEvent()
		if ev == nil {
			// screen is finished, should not happen
			return errors.New("screen was finished")
		}
		switch tev := ev.(type) {
		case *tcell.EventInterrupt:
			return nil
		case *tcell.EventError:
			return tev
		case *tcell.EventKey:
			msg := gruid.MsgKeyDown{}
			mod := tev.Modifiers()
			if mod&tcell.ModShift != 0 {
				msg.Mod |= gruid.ModShift
			}
			if mod&tcell.ModCtrl != 0 {
				msg.Mod |= gruid.ModCtrl
			}
			if mod&tcell.ModAlt != 0 {
				msg.Mod |= gruid.ModAlt
			}
			if mod&tcell.ModMeta != 0 { // never reported
				msg.Mod |= gruid.ModMeta
			}
			msg.Time = tev.When()
			switch tev.Key() {
			case tcell.KeyDown:
				msg.Key = gruid.KeyArrowDown
			case tcell.KeyLeft:
				msg.Key = gruid.KeyArrowLeft
			case tcell.KeyRight:
				msg.Key = gruid.KeyArrowRight
			case tcell.KeyUp:
				msg.Key = gruid.KeyArrowUp
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				msg.Key = gruid.KeyBackspace
			case tcell.KeyDelete:
				msg.Key = gruid.KeyDelete
			case tcell.KeyEnd:
				msg.Key = gruid.KeyEnd
			case tcell.KeyEscape:
				msg.Key = gruid.KeyEscape
			case tcell.KeyEnter:
				msg.Key = gruid.KeyEnter
			case tcell.KeyHome:
				msg.Key = gruid.KeyHome
			case tcell.KeyInsert:
				msg.Key = gruid.KeyInsert
			case tcell.KeyPgUp:
				msg.Key = gruid.KeyPageUp
			case tcell.KeyPgDn:
				msg.Key = gruid.KeyPageDown
			case tcell.KeyTab:
				msg.Key = gruid.KeyTab
			case tcell.KeyBacktab:
				msg.Key = gruid.KeyTab
				msg.Mod = gruid.ModShift
			}
			if tev.Rune() != 0 && msg.Key == "" {
				msg.Key = gruid.Key(tev.Rune())
			}
			if msg.Key == "" {
				continue
			}
			send(msg)
		case *tcell.EventMouse:
			x, y := tev.Position()
			switch tev.Buttons() {
			case tcell.Button1:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.P = gruid.Point{X: x, Y: y}
				if dr.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseMain
					dr.mousedrag = true
				}
				send(msg)
			case tcell.Button3:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.P = gruid.Point{X: x, Y: y}
				if dr.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseAuxiliary
					dr.mousedrag = true
				}
				send(msg)
			case tcell.Button2:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.P = gruid.Point{X: x, Y: y}
				if dr.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseSecondary
					dr.mousedrag = true
				}
				send(msg)
			case tcell.WheelUp:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.P = gruid.Point{X: x, Y: y}
				msg.Action = gruid.MouseWheelUp
				send(msg)
			case tcell.WheelDown:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.P = gruid.Point{X: x, Y: y}
				msg.Action = gruid.MouseWheelDown
				send(msg)
			case tcell.ButtonNone:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.P = gruid.Point{X: x, Y: y}
				if dr.mousedrag {
					dr.mousedrag = false
					msg.Action = gruid.MouseRelease
				} else {
					if dr.mousePos == msg.P {
						continue
					}
					msg.Action = gruid.MouseMove
					dr.mousePos = msg.P
				}
				send(msg)
			}
		case *tcell.EventResize:
			msg := gruid.MsgScreen{}
			msg.Time = tev.When()
			msg.Width, msg.Height = tev.Size()
			send(msg)
		}
	}
}

// Flush implements gruid.Driver.Flush.
func (dr *Driver) Flush(frame gruid.Frame) {
	for _, fc := range frame.Cells {
		c := fc.Cell
		st := dr.sm.GetStyle(c.Style)
		dr.screen.SetContent(fc.P.X, fc.P.Y, c.Rune, nil, st)
	}
	dr.screen.Show()
}

// Close implements gruid.Driver.Close. It finalizes the screen and releases
// resources.
func (dr *Driver) Close() {
	if !dr.init {
		return
	}
	if !dr.noQuit {
		dr.screen.Fini()
		dr.init = false
	}
	dr.noQuit = false
}
