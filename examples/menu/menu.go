package main

import (
	"fmt"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	"github.com/anaseto/gruid/ui"
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
	fmt.Printf("Successful quit.\n")
}

const (
	ColorHeader gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorSelected
	ColorAltBg
	ColorTitle
)

type styler struct{}

func (sty styler) GetStyle(st gruid.CellStyle) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
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
	grid  gruid.Grid
	menu  *ui.Menu
	label *ui.Label
	init  bool
}

func (m *model) Init() gruid.Cmd {
	entries := []ui.MenuEntry{
		{Header: true, Text: "Header"},
		{Text: "(F)irst", Keys: []gruid.Key{"f", "F"}},
		{Text: "(S)econd", Keys: []gruid.Key{"s", "S"}},
		{Text: "(T)hird", Keys: []gruid.Key{"t", "T"}},
	}
	st := gruid.CellStyle{}
	style := ui.MenuStyle{
		BgAlt:    ColorAltBg,
		Selected: ColorSelected,
		Header:   st.WithFg(ColorHeader),
		Title:    st.WithFg(ColorTitle),
	}
	menu := ui.NewMenu(ui.MenuConfig{
		Grid:    m.grid.Slice(gruid.NewRange(0, 0, 20, len(entries)+2)),
		Entries: entries,
		Title:   "Menu",
		Style:   style,
	})
	m.menu = menu
	label := ui.NewLabel(ui.LabelConfig{
		Grid:       m.grid.Slice(gruid.NewRange(22, 0, 70, 5)),
		Title:      "Menu Last Action",
		StyledText: ui.NewStyledText("Nothing done yet!"),
		Style:      ui.LabelStyle{Title: st.WithFg(ColorHeader), AdjustWidth: true},
	})
	m.label = label
	m.init = true
	return nil
}

func (m *model) Update(msg gruid.Msg) gruid.Cmd {
	m.init = false
	m.menu.Update(msg)
	switch m.menu.Action() {
	case ui.MenuCancel:
		return gruid.Quit
	case ui.MenuMove:
		m.label.SetText(fmt.Sprintf("moved selection to entry number %d", m.menu.Selection()))
	case ui.MenuActivate:
		m.label.SetText(fmt.Sprintf("activated entry number %d", m.menu.Selection()))
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	if m.menu.Action() != ui.MenuPass || m.init {
		m.grid.Iter(func(pos gruid.Position) {
			m.grid.SetCell(pos, gruid.Cell{Rune: ' '})
		})
		m.menu.Draw()
		m.label.Draw()
	}
	return m.grid
}
