package main

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	tc "github.com/gdamore/tcell/v2"
)

func main() {
	var dodo = gruid.NewGrid(gruid.GridConfig{})
	st := styler{}
	var dododri = &tcell.Driver{StyleManager: st}
	m := &model{grid: dodo}
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dododri,
		Model:  m,
	})
	app.Start()
}

const (
	Black gruid.Color = iota
	White
	Blue
)

type styler struct{}

func (st styler) GetStyle(cell gruid.Cell) tc.Style {
	ts := tc.StyleDefault
	switch cell.Fg {
	case Black:
		ts.Foreground(tc.ColorBlack)
	case White:
		ts.Foreground(tc.ColorWhite)
	case Blue:
		ts.Foreground(tc.ColorBlue)
	}
	switch cell.Bg {
	case Black:
		ts.Background(tc.ColorBlack)
	case White:
		ts.Background(tc.ColorWhite)
	}
	return ts
}

type model struct {
	playerPos gruid.Position
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
	m.grid.Iter(func(pos gruid.Position) {
		if pos == m.playerPos {
			m.grid.SetCell(pos, gruid.Cell{Fg: Blue, Bg: Black, Rune: '@'})
		} else {
			m.grid.SetCell(pos, gruid.Cell{Fg: White, Bg: Black, Rune: '.'})
		}
	})
	return m.grid
}
