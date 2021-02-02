package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	"github.com/anaseto/gruid/ui"
	tc "github.com/gdamore/tcell/v2"
)

func main() {
	gd := gruid.NewGrid(80, 24)
	st := styler{}
	dr := tcell.NewDriver(tcell.Config{StyleManager: st})
	m := newModel(gd)
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dr,
		Model:  m,
	})
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Successful quit.\n")
	}
}

// Those constants represent the generic colors we use in this example.
const (
	ColorHeader gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorActive
	ColorAltBg
	ColorTitle
	ColorKey
)

// styler implements the tcell.StyleManager interface.
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorHeader:
		ts = ts.Foreground(tc.ColorNavy)
	case ColorKey:
		ts = ts.Foreground(tc.ColorGreen)
	case ColorActive, ColorTitle:
		ts = ts.Foreground(tc.ColorOlive)
	}
	switch st.Bg {
	case ColorAltBg:
		ts = ts.Background(tc.ColorBlack)
	}
	return ts
}

// model implements gruid.Model.
type model struct {
	grid  gruid.Grid
	menu  *ui.Menu
	label *ui.Label
}

func newModel(gd gruid.Grid) *model {
	m := &model{grid: gd}
	st := gruid.Style{}
	stt := ui.Text("").WithMarkup('k', st.WithFg(ColorKey))
	entries := []ui.MenuEntry{
		{Disabled: true, Text: stt.With("Header", st.WithFg(ColorHeader))},
		{Text: stt.WithText("(@kF@N)irst"), Keys: []gruid.Key{"f", "F"}},
		{Text: stt.WithText("(@kS@N)econd"), Keys: []gruid.Key{"s", "S"}},
		{Text: stt.WithText("(@kT@N)hird"), Keys: []gruid.Key{"t", "T"}},
		{Text: stt.WithText("(@kL@N)ast"), Keys: []gruid.Key{"l", "L"}},
	}
	for i := range entries {
		if i%2 == 1 {
			entries[i].Text = entries[i].Text.WithStyle(st.WithBg(ColorAltBg))
		}
	}
	style := ui.MenuStyle{
		//Layout: gruid.Point{0, 1}, // one-line layout (with two pages)
		//Layout:   gruid.Point{2, 2}, // tabular layout (with two pages)
		Active: st.WithFg(ColorActive),
	}
	menu := ui.NewMenu(ui.MenuConfig{
		Grid:    gruid.NewGrid(30, len(entries)+2),
		Entries: entries,
		Box:     &ui.Box{Title: ui.Text("Menu").WithStyle(st.WithFg(ColorTitle))},
		Style:   style,
	})
	m.menu = menu
	label := &ui.Label{
		Box:         &ui.Box{Title: ui.Text("Menu Last Action").WithStyle(st.WithFg(ColorHeader))},
		Content:  ui.Text("Nothing done yet!"),
		AdjustWidth: true,
	}
	m.label = label
	return m
}

// Update implements gruid.Model.Update.
func (m *model) Update(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgInit:
	default:
		m.menu.Update(msg)
	}
	switch m.menu.Action() {
	case ui.MenuQuit:
		// The user requested to quit the menu, either with a key or by
		// clicking outside the menu.
		return gruid.End()
	case ui.MenuMove:
		m.label.SetText(fmt.Sprintf("activated entry number %d", m.menu.Active()))
	case ui.MenuInvoke:
		m.label.SetText(fmt.Sprintf("invoked entry number %d", m.menu.Active()))
	}
	return nil
}

// Draw implements gruid.Model.Draw.
func (m *model) Draw() gruid.Grid {
	m.grid.Fill(gruid.Cell{Rune: ' '})
	m.grid.Copy(m.menu.Draw())
	m.label.Draw(m.grid.Slice(gruid.NewRange(32, 0, 70, 5)))
	return m.grid
}
