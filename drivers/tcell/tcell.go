package tcell

import (
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
	StyleManager StyleManager
	screen       tcell.Screen
	mousedrag    bool
	mousePos     gruid.Position
}

// Init implements gruid.Driver.Init. It initializes a screen using the tcell
// terminal library.
func (t *Driver) Init() error {
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

// PollMsg implements gruid.Driver.PollMsg.
func (t *Driver) PollMsg() (gruid.Msg, error) {
	for {
		ev := t.screen.PollEvent()
		if ev == nil {
			// screen is finished
			return nil, nil
		}
		switch tev := ev.(type) {
		case *tcell.EventError:
			return nil, tev
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
			return msg, nil
		case *tcell.EventMouse:
			x, y := tev.Position()
			switch tev.Buttons() {
			case tcell.Button1:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.MousePos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseMain
					t.mousedrag = true
				}
				return msg, nil
			case tcell.Button3:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.MousePos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseAuxiliary
					t.mousedrag = true
				}
				return msg, nil
			case tcell.Button2:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.MousePos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					msg.Action = gruid.MouseMove
				} else {
					msg.Action = gruid.MouseSecondary
					t.mousedrag = true
				}
				return msg, nil
			case tcell.WheelUp:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.MousePos = gruid.Position{X: x, Y: y}
				msg.Action = gruid.MouseWheelUp
				return msg, nil
			case tcell.WheelDown:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.MousePos = gruid.Position{X: x, Y: y}
				msg.Action = gruid.MouseWheelDown
				return msg, nil
			case tcell.ButtonNone:
				msg := gruid.MsgMouse{}
				msg.Time = tev.When()
				msg.MousePos = gruid.Position{X: x, Y: y}
				if t.mousedrag {
					t.mousedrag = false
					msg.Action = gruid.MouseRelease
				} else {
					if t.mousePos == msg.MousePos {
						continue
					}
					msg.Action = gruid.MouseMove
					t.mousePos = msg.MousePos
				}
				return msg, nil
			}
		case *tcell.EventResize:
			msg := gruid.MsgScreenSize{}
			msg.Time = tev.When()
			msg.Width, msg.Height = tev.Size()
			return msg, nil
		}
	}
}

// Flush implements gruid.Driver.Flush.
func (t *Driver) Flush(frame gruid.Frame) {
	for _, cdraw := range frame.Cells {
		c := cdraw.Cell
		st := t.StyleManager.GetStyle(c.Style)
		t.screen.SetContent(cdraw.Pos.X, cdraw.Pos.Y, c.Rune, nil, st)
	}
	t.screen.Show()
}

// Close implements gruid.Driver.Close. It finalizes the screen and releases
// resources.
func (t *Driver) Close() {
	t.screen.Fini()
}
