package tcell

import (
	"github.com/anaseto/gorltk"
	"github.com/gdamore/tcell/v2"
)

type StyleManager interface {
	// GetAttributes returns a mask of text attributes for a given cell
	// style.
	GetStyle(gorltk.Cell) tcell.Style
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

func (t *Driver) Flush(gd *gorltk.Grid) {
	for _, cdraw := range gd.Frame().Cells {
		c := cdraw.Cell
		st := t.StyleManager.GetStyle(c)
		t.screen.SetContent(cdraw.Pos.X, cdraw.Pos.Y, c.Rune, nil, st)
	}
	//ui.g.Printf("%d %d %d", ui.g.DrawFrame, ui.g.DrawFrameStart, len(ui.g.DrawLog))
	t.screen.Show()
}

func (t *Driver) Interrupt() {
	t.screen.PostEvent(tcell.NewEventInterrupt(nil))
}

func (t *Driver) PollMsg() (gorltk.Msg, bool) {
	for {
		switch tev := t.screen.PollEvent().(type) {
		case *tcell.EventKey:
			ev := gorltk.MsgKeyDown{}
			ev.Time = tev.When()
			switch tev.Key() {
			case tcell.KeyDown:
				ev.Key = gorltk.KeyArrowDown
			case tcell.KeyLeft:
				ev.Key = gorltk.KeyArrowLeft
			case tcell.KeyRight:
				ev.Key = gorltk.KeyArrowRight
			case tcell.KeyUp:
				ev.Key = gorltk.KeyArrowUp
			case tcell.KeyBackspace:
				ev.Key = gorltk.KeyBackspace
			case tcell.KeyDelete:
				ev.Key = gorltk.KeyDelete
			case tcell.KeyEnd:
				ev.Key = gorltk.KeyEnd
			case tcell.KeyEscape:
				ev.Key = gorltk.KeyEscape
			case tcell.KeyEnter:
				ev.Key = gorltk.KeyEnter
			case tcell.KeyHome:
				ev.Key = gorltk.KeyHome
			case tcell.KeyInsert:
				ev.Key = gorltk.KeyInsert
			case tcell.KeyPgUp:
				ev.Key = gorltk.KeyPageUp
			case tcell.KeyPgDn:
				ev.Key = gorltk.KeyPageDown
			case tcell.KeyTab:
				ev.Key = gorltk.KeyTab
			}
			if tev.Rune() != 0 && ev.Key == "" {
				ev.Key = gorltk.Key(tev.Rune())
			}
			return ev, true
		case *tcell.EventMouse:
			x, y := tev.Position()
			switch tev.Buttons() {
			case tcell.Button1:
				ev := gorltk.MsgMouseDown{}
				ev.Time = tev.When()
				ev.MousePos = gorltk.Position{X: x, Y: y}
				ev.Button = gorltk.ButtonMain
				return ev, true
			case tcell.Button3:
				ev := gorltk.MsgMouseDown{}
				ev.Time = tev.When()
				ev.MousePos = gorltk.Position{X: x, Y: y}
				ev.Button = gorltk.ButtonAuxiliary
				return ev, true
			case tcell.Button2:
				ev := gorltk.MsgMouseDown{}
				ev.Time = tev.When()
				ev.MousePos = gorltk.Position{X: x, Y: y}
				ev.Button = gorltk.ButtonSecondary
				return ev, true
			case tcell.ButtonNone:
				ev := gorltk.MsgMouseMove{}
				ev.Time = tev.When()
				ev.MousePos = gorltk.Position{X: x, Y: y}
				return ev, true
			}
		case *tcell.EventInterrupt:
			return nil, false
		case *tcell.EventResize:
			ev := gorltk.MsgScreenSize{}
			ev.Time = tev.When()
			ev.Width, ev.Height = tev.Size()
			return ev, true
		}
	}
}

func (t *Driver) ClearCache() {
	// no special cache
}
