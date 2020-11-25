package tcell

import (
	"github.com/anaseto/gorltk"
	"github.com/gdamore/tcell/v2"
)

type StyleManager interface {
	// GetAttributes returns a mask of text attributes for a given cell
	// style.
	GetStyle(gorltk.GridCell) tcell.Style
}

type Driver struct {
	// AttributeGetter has to be provided to make use of tcell text
	// attributes.
	StyleManager StyleManager
	screen       tcell.Screen
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
	return nil
}

func (t *Driver) Close() {
	t.screen.Fini()
}

func (t *Driver) Flush(gd gorltk.GridDrawer) {
	for _, cdraw := range gd.FrameCells() {
		c := cdraw.Cell
		st := t.StyleManager.GetStyle(c)
		t.screen.SetContent(cdraw.Pos.X, cdraw.Pos.Y, c.R, nil, st)
	}
	//ui.g.Printf("%d %d %d", ui.g.DrawFrame, ui.g.DrawFrameStart, len(ui.g.DrawLog))
	t.screen.Show()
}

func (t *Driver) Interrupt() {
	t.screen.PostEvent(tcell.NewEventInterrupt(nil))
}

func (t *Driver) PollEvent() gorltk.Event {
	for {
		switch tev := t.screen.PollEvent().(type) {
		case *tcell.EventKey:
			ev := gorltk.EventKeyDown{}
			ev.Time = tev.When()
			switch tev.Key() {
			case tcell.KeyEsc:
				ev.Key = " "
			case tcell.KeyLeft:
				// TODO: will not work if user changes keybindings
				ev.Key = "4"
			case tcell.KeyDown:
				ev.Key = "2"
			case tcell.KeyUp:
				ev.Key = "8"
			case tcell.KeyRight:
				ev.Key = "6"
			case tcell.KeyPgUp:
				ev.Key = "u"
			case tcell.KeyPgDn:
				ev.Key = "d"
			case tcell.KeyDelete:
				ev.Key = "5"
			case tcell.KeyCtrlW:
				ev.Key = "W"
			case tcell.KeyCtrlQ:
				ev.Key = "Q"
			case tcell.KeyCtrlP:
				ev.Key = "m"
			case tcell.KeyEnter:
				ev.Key = "."
			}
			if tev.Rune() != 0 && ev.Key == "" {
				ev.Key = string(tev.Rune())
			}
			return ev
		case *tcell.EventMouse:
			ev := gorltk.EventMouseDown{}
			ev.Time = tev.When()
			x, y := tev.Position()
			ev.MousePos = gorltk.Position{X: x, Y: y}
			switch tev.Buttons() {
			case tcell.Button1:
				ev.Button = gorltk.ButtonMain
			case tcell.Button2:
				ev.Button = gorltk.ButtonAuxiliary
			case tcell.Button3:
				ev.Button = gorltk.ButtonSecondary
			}
			return ev
		case *tcell.EventInterrupt:
			ev := gorltk.EventMouseMove{}
			ev.Time = tev.When()
			return ev
		}
	}
}

func (t *Driver) ClearCache() {
	// no special cache
}
