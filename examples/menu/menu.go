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
	ColorUnavailable gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorHeader
	ColorSelected
	ColorAltBg
	ColorTitle
)

type styler struct{}

func (sty styler) GetStyle(st gruid.CellStyle) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorUnavailable:
		ts = ts.Foreground(tc.ColorWhite)
	case ColorHeader:
		ts = ts.Foreground(tc.ColorNavy)
	case ColorSelected, ColorTitle:
		ts = ts.Foreground(tc.ColorOlive)
	}
	switch st.Bg {
	case ColorAltBg:
		ts = ts.Background(tc.ColorBlack)
	}
	return ts
}

type model struct {
	playerPos gruid.Position
	grid      gruid.Grid
	menu      *models.Menu
	label     *models.Label
}

func (m *model) Init() gruid.Cmd {
	entries := []models.MenuEntry{
		{Kind: models.EntryHeader, Text: "Header"},
		{Kind: models.EntryChoice, Text: "(F)irst", Key: "F"},
		{Kind: models.EntryChoice, Text: "(S)econd", Key: "S"},
	}
	st := gruid.CellStyle{}
	style := models.MenuStyle{
		Boxed:       true,
		Content:     st,
		BgAlt:       ColorAltBg,
		Selected:    ColorSelected,
		Unavailable: ColorUnavailable,
		Header:      st.Foreground(ColorHeader),
		Title:       st.Foreground(ColorTitle),
	}
	menu := models.NewMenu(models.MenuConfig{
		Grid:    m.grid.Slice(gruid.NewRange(0, 0, 20, len(entries)+2)),
		Entries: entries,
		Title:   "Menu",
		Style:   style,
	})
	m.menu = menu
	label := &models.Label{
		Boxed: true,
		Grid:  m.grid.Slice(gruid.NewRange(22, 0, 70, 5)),
		Title: "Menu Action",
		Text:  "",
		Style: models.LabelStyle{Content: st, Title: st.Foreground(ColorHeader)},
	}
	m.label = label
	return nil
}

func (m *model) Update(msg gruid.Msg) gruid.Cmd {
	m.menu.Update(msg)
	switch m.menu.Action() {
	case models.MenuCancel:
		return gruid.Quit
	case models.MenuMove:
		m.label.Text = "changed selection"
	case models.MenuAccept:
		m.label.Text = "accepted selection"
	case models.MenuPass:
		m.label.Text = "no action but that is not important the thing is that it does not fit"
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	m.label.Draw()
	return m.menu.Draw()
}
