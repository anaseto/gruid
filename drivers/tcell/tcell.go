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
	// AttributeGetter has to be provided to make use of tcell text
	// attributes.
	sm        StyleManager
	screen    tcell.Screen
	mousedrag bool
	mousePos  gruid.Position
	closed    bool
}

// Config contains configurations options for the driver.
type Config struct {
	StyleManager StyleManager // for cell styling
}

// NewDriver returns a new driver with given configuration options.
func NewDriver(cfg Config) *Driver {
	return &Driver{sm: cfg.StyleManager}
}

// Init implements gruid.Driver.Init. It initializes a screen using the tcell
// terminal library.
func (t *Driver) Init() error {
	if t.sm == nil {
		return errors.New("no style manager provided")
	}
	screen, err := tcell.NewScreen()
	t.screen = screen
	if err != nil {
		return err
	}
	err = t.screen.Init()
	if err != nil {
		return err
	}
	t.screen.SetStyle(tcell.StyleDefault)
	t.screen.EnableMouse()
	t.screen.HideCursor()

	// try to send initial size
	w, h := t.screen.Size()
	t.screen.PostEvent(tcell.NewEventResize(w, h))
	return nil
}

// PollMsgs implements gruid.Driver.PollMsgs.
func (t *Driver) PollMsgs(ctx context.Context, msgs chan<- gruid.Msg) error {
	go func(ctx context.Context) {
		<-ctx.Done()
		t.screen.PostEvent(tcell.NewEventInterrupt(0))
	}(ctx)
	send := func(msg gruid.Msg) {
		select {
		case msgs <- msg:
		case <-ctx.Done():
		}
	}
	for {
		ev := t.screen.PollEvent()
		if ev == nil {
			// screen is finished, should not happen
			return nil
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
				msg.Pos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseMain
					t.mousedrag = true
				}
				send(msg)
			case tcell.Button3:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.Pos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseAuxiliary
					t.mousedrag = true
				}
				send(msg)
			case tcell.Button2:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.Pos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseSecondary
					t.mousedrag = true
				}
				send(msg)
			case tcell.WheelUp:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.Pos = gruid.Position{X: x, Y: y}
				msg.Action = gruid.MouseWheelUp
				send(msg)
			case tcell.WheelDown:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.Pos = gruid.Position{X: x, Y: y}
				msg.Action = gruid.MouseWheelDown
				send(msg)
			case tcell.ButtonNone:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.Pos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					t.mousedrag = false
					msg.Action = gruid.MouseRelease
				} else {
					if t.mousePos == msg.Pos {
						continue
					}
					msg.Action = gruid.MouseMove
					t.mousePos = msg.Pos
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
func (t *Driver) Flush(frame gruid.Frame) {
	for _, cdraw := range frame.Cells {
		c := cdraw.Cell
		st := t.sm.GetStyle(c.Style)
		t.screen.SetContent(cdraw.Pos.X, cdraw.Pos.Y, c.Rune, nil, st)
	}
	t.screen.Show()
}

// Close implements gruid.Driver.Close. It finalizes the screen and releases
// resources.
func (t *Driver) Close() {
	t.screen.Fini()
}
