package main

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	tc "github.com/gdamore/tcell/v2"
)

func main() {

	// use tcell terminal driver
	st := styler{}
	dri := &tcell.Driver{StyleManager: st}

	// our application's state and grid with default config
	gd := gruid.NewGrid(gruid.GridConfig{})
	m := &model{grid: gd}

	// define new application
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  m,
	})

	// start application
	app.Start()
}

const (
	ColorPlayer gruid.Color = 1 + iota // skip special zero value gruid.ColorDefault
)

// type that implements driver's style manager interface
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorPlayer:
		ts = ts.Foreground(tc.ColorNavy) // blue color for the player
	}
	return ts
}

type model struct {
	playerPos gruid.Position // track player position
	grid      gruid.Grid
}

func (m *model) Init() gruid.Cmd {
	return nil
}

func (m *model) Update(msg gruid.Msg) gruid.Cmd {
	xmax, ymax := m.grid.Size()
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case gruid.KeyArrowDown, "j":
			if m.playerPos.Y < ymax-1 {
				m.playerPos.Y++
			}
		case gruid.KeyArrowLeft, "h":
			if m.playerPos.X > 0 {
				m.playerPos.X--
			}
		case gruid.KeyArrowRight, "l":
			if m.playerPos.X < xmax-1 {
				m.playerPos.X++
			}
		case gruid.KeyArrowUp, "k":
			if m.playerPos.Y > 0 {
				m.playerPos.Y--
			}
		case "Q", "q", gruid.KeyEscape:
			return gruid.Quit
		}
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	c := gruid.Cell{Rune: '.'} // default cell
	m.grid.Iter(func(pos gruid.Position) {
		if pos == m.playerPos {
			m.grid.SetCell(pos, gruid.Cell{Rune: '@', Style: c.Style.WithFg(ColorPlayer)})
		} else {
			m.grid.SetCell(pos, c)
		}
	})
	return m.grid
}
