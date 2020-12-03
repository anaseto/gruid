package main

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	tc "github.com/gdamore/tcell/v2"
)

func main() {
	var gd = gruid.NewGrid(gruid.GridConfig{})
	st := styler{}
	var dri = &tcell.Driver{StyleManager: st}
	m := &model{grid: gd}
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  m,
	})
	app.Start()
}

const (
	Black gruid.Color = iota
	Gray
	White
	Navy
)

type styler struct{}

func (st styler) GetStyle(cell gruid.Cell) tc.Style {
	ts := tc.StyleDefault
	switch cell.Fg {
	case Gray:
		ts = ts.Foreground(tc.ColorGray)
	case Black:
		ts = ts.Foreground(tc.ColorBlack)
	case White:
		ts = ts.Foreground(tc.ColorDefault)
	case Navy:
		ts = ts.Foreground(tc.ColorNavy)
	}
	switch cell.Bg {
	case Black:
		ts = ts.Background(tc.ColorBlack)
	case White:
		ts = ts.Background(tc.ColorWhite)
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
			m.grid.SetCell(pos, gruid.Cell{Fg: Navy, Bg: Gray, Rune: '@'})
		} else {
			m.grid.SetCell(pos, gruid.Cell{Fg: White, Bg: Gray, Rune: '.'})
		}
	})
	return m.grid
}
