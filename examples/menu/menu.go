package main

import (
	"fmt"
	"log"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	"github.com/anaseto/gruid/ui"
	tc "github.com/gdamore/tcell/v2"
)

func main() {
	var gd = gruid.NewGrid(gruid.GridConfig{})
	st := styler{}
	var dri = &tcell.Driver{StyleManager: st}
	m := NewModel(gd)
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  m,
	})
	if err := app.Start(nil); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Successful quit.\n")
	}
}

const (
	ColorHeader gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorSelected
	ColorAltBg
	ColorTitle
	ColorKey
)

type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorHeader:
		ts = ts.Foreground(tc.ColorNavy)
	case ColorKey:
		ts = ts.Foreground(tc.ColorGreen)
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
}

func NewModel(gd gruid.Grid) *model {
	m := &model{grid: gd}
	entries := []ui.MenuEntry{
		{Header: true, Text: "Header"},
		{Text: "(@kF@N)irst", Keys: []gruid.Key{"f", "F"}},
		{Text: "(@kS@N)econd", Keys: []gruid.Key{"s", "S"}},
		{Text: "(@kT@N)hird", Keys: []gruid.Key{"t", "T"}},
	}
	st := gruid.Style{}
	style := ui.MenuStyle{
		BgAlt:    ColorAltBg,
		Selected: ColorSelected,
		Header:   st.WithFg(ColorHeader),
		Title:    st.WithFg(ColorTitle),
	}
	menu := ui.NewMenu(ui.MenuConfig{
		Grid:       m.grid.Slice(gruid.NewRange(0, 0, 20, len(entries)+2)),
		Entries:    entries,
		StyledText: ui.NewStyledText("").WithMarkup('k', st.WithFg(ColorKey)),
		Title:      "Menu",
		Style:      style,
	})
	m.menu = menu
	label := ui.NewLabel(ui.LabelConfig{
		Grid:       m.grid.Slice(gruid.NewRange(22, 0, 70, 5)),
		Title:      "Menu Last Action",
		StyledText: ui.NewStyledText("Nothing done yet!"),
		Style:      ui.LabelStyle{Title: st.WithFg(ColorHeader), AdjustWidth: true},
	})
	m.label = label
	return m
}

func (m *model) Update(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgInit:
	default:
		m.menu.Update(msg)
	}
	switch m.menu.Action() {
	case ui.MenuQuit:
		return gruid.Quit()
	case ui.MenuMove:
		m.label.SetText(fmt.Sprintf("moved selection to entry number %d", m.menu.Selection()))
	case ui.MenuActivate:
		m.label.SetText(fmt.Sprintf("activated entry number %d", m.menu.Selection()))
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	m.grid.Fill(gruid.Cell{Rune: ' '})
	m.menu.Draw()
	m.label.Draw()
	return m.grid
}
