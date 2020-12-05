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

func (t *Driver) Flush(gd gruid.Grid) {
	for _, cdraw := range gd.Frame().Cells {
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
			ev := gruid.MsgKeyDown{}
			ev.Time = tev.When()
			switch tev.Key() {
			case tcell.KeyDown:
				ev.Key = gruid.KeyArrowDown
			case tcell.KeyLeft:
				ev.Key = gruid.KeyArrowLeft
			case tcell.KeyRight:
				ev.Key = gruid.KeyArrowRight
			case tcell.KeyUp:
				ev.Key = gruid.KeyArrowUp
			case tcell.KeyBackspace:
				ev.Key = gruid.KeyBackspace
			case tcell.KeyDelete:
				ev.Key = gruid.KeyDelete
			case tcell.KeyEnd:
				ev.Key = gruid.KeyEnd
			case tcell.KeyEscape:
				ev.Key = gruid.KeyEscape
			case tcell.KeyEnter:
				ev.Key = gruid.KeyEnter
			case tcell.KeyHome:
				ev.Key = gruid.KeyHome
			case tcell.KeyInsert:
				ev.Key = gruid.KeyInsert
			case tcell.KeyPgUp:
				ev.Key = gruid.KeyPageUp
			case tcell.KeyPgDn:
				ev.Key = gruid.KeyPageDown
			case tcell.KeyTab:
				ev.Key = gruid.KeyTab
			}
			if tev.Rune() != 0 && ev.Key == "" {
				ev.Key = gruid.Key(tev.Rune())
			}
			return ev
		case *tcell.EventMouse:
			x, y := tev.Position()
			switch tev.Buttons() {
			case tcell.Button1:
				ev := gruid.MsgMouseDown{}
				ev.Time = tev.When()
				ev.MousePos = gruid.Position{X: x, Y: y}
				ev.Button = gruid.ButtonMain
				return ev
			case tcell.Button3:
				ev := gruid.MsgMouseDown{}
				ev.Time = tev.When()
				ev.MousePos = gruid.Position{X: x, Y: y}
				ev.Button = gruid.ButtonAuxiliary
				return ev
			case tcell.Button2:
				ev := gruid.MsgMouseDown{}
				ev.Time = tev.When()
				ev.MousePos = gruid.Position{X: x, Y: y}
				ev.Button = gruid.ButtonSecondary
				return ev
			case tcell.ButtonNone:
				ev := gruid.MsgMouseMove{}
				ev.Time = tev.When()
				ev.MousePos = gruid.Position{X: x, Y: y}
				return ev
			}
		case *tcell.EventResize:
			ev := gruid.MsgScreenSize{}
			ev.Time = tev.When()
			ev.Width, ev.Height = tev.Size()
			return ev
		}
	}
}
