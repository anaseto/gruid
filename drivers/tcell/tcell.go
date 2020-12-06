package tcell

import (
	"github.com/anaseto/gruid"
	"github.com/gdamore/tcell/v2"
)

type StyleManager interface {
	// GetAttributes returns a mask of text attributes for a given cell
	// style.
	GetStyle(gruid.Style) tcell.Style
}

type Driver struct {
	// AttributeGetter has to be provided to make use of tcell text
	// attributes.
	StyleManager StyleManager
	screen       tcell.Screen
	mousedrag    bool
	mousePos     gruid.Position
}

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

func (t *Driver) Close() {
	t.screen.Fini()
}

func (t *Driver) Flush(frame gruid.Frame) {
	for _, cdraw := range frame.Cells {
		c := cdraw.Cell
		st := t.StyleManager.GetStyle(c.Style)
		t.screen.SetContent(cdraw.Pos.X, cdraw.Pos.Y, c.Rune, nil, st)
	}
	t.screen.Show()
}

func (t *Driver) PollMsg() gruid.Msg {
	for {
		switch tev := t.screen.PollEvent().(type) {
		case *tcell.EventKey:
			msg := gruid.MsgKeyDown{}
			if tev.Modifiers() == tcell.ModShift {
				msg.Shift = true
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
			case tcell.KeyBackspace:
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
			}
			if tev.Rune() != 0 && msg.Key == "" {
				msg.Key = gruid.Key(tev.Rune())
			}
			return msg
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
				return msg
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
				return msg
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
				return msg
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
				return msg
			}
		case *tcell.EventResize:
			msg := gruid.MsgScreenSize{}
			msg.Time = tev.When()
			msg.Width, msg.Height = tev.Size()
			return msg
		}
	}
}
