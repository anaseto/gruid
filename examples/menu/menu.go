package main

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	"github.com/anaseto/gruid/models"
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
	White
	Navy
	Green
	Gray
)

type styler struct{}

func (st styler) GetStyle(cell gruid.Cell) tc.Style {
	ts := tc.StyleDefault
	switch cell.Fg {
	case Black:
		ts = ts.Foreground(tc.ColorBlack)
	case White:
		ts = ts.Foreground(tc.ColorWhite)
	case Navy:
		ts = ts.Foreground(tc.ColorNavy)
	case Green:
		ts = ts.Foreground(tc.ColorGreen)
	}
	switch cell.Bg {
	case Black:
		ts = ts.Background(tc.ColorBlack)
	case White:
		ts = ts.Background(tc.ColorWhite)
	case Gray:
		ts = ts.Background(tc.ColorGray)
	}
	return ts
}

type model struct {
	playerPos gruid.Position
	grid      gruid.Grid
	menu      *models.Menu
}

func (m *model) Init() gruid.Cmd {
	entries := []models.MenuEntry{}
	entries = append(entries,
		models.MenuEntry{Kind: models.EntryHeader, Text: "Header"},
		models.MenuEntry{Kind: models.EntryChoice, Text: "(F)irst", Key: "F"},
		models.MenuEntry{Kind: models.EntryChoice, Text: "(S)econd", Key: "S"},
	)
	style := models.MenuStyle{
		Boxed:            true,
		ColorBg:          Black,
		ColorBgAlt:       Gray,
		ColorFg:          White,
		ColorAvailable:   White,
		ColorSelected:    Green,
		ColorUnavailable: White,
		ColorHeader:      Navy,
		ColorTitle:       Green,
	}
	menu := models.NewMenu(models.MenuConfig{
		Grid:    m.grid.Slice(gruid.Range{m.grid.Range().Min, m.grid.Range().Min.Shift(20, len(entries)+2)}),
		Entries: entries,
		Title:   "Menu",
		Style:   style,
	})
	m.menu = menu
	return nil
}

func (m *model) Update(msg gruid.Msg) gruid.Cmd {
	m.menu.Update(msg)
	if m.menu.Action() == models.MenuCancel {
		return gruid.Quit
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	return m.menu.Draw()
}
