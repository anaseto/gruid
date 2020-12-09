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
	m := NewModel(gd)
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  m,
	})
	app.Start(nil)
	fmt.Printf("Successful quit.\n")
}

const (
	ColorHeader gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorSelected
	ColorAltBg
	ColorTitle
)

type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
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
	label *ui.Label
	init  bool
}

func NewModel(gd gruid.Grid) *model {
	m := &model{grid: gd}
	st := gruid.Style{}
	m.grid = m.grid.Slice(gruid.NewRange(0, 0, 80, 5))
	label := ui.NewLabel(ui.LabelConfig{
		Grid:       m.grid,
		Title:      "Menu Last Action",
		StyledText: ui.NewStyledText("No input messages yet!"),
		Style:      ui.LabelStyle{Title: st.WithFg(ColorHeader)},
	})
	m.label = label
	m.init = true
	return m
}

func (m *model) Update(msg gruid.Msg) gruid.Effect {
	if _, ok := msg.(gruid.MsgDraw); ok {
		return nil
	}
	if msg, ok := msg.(gruid.MsgKeyDown); ok {
		switch msg.Key {
		case gruid.KeyEscape, "Q", "q":
			return gruid.Quit()
		}
	}
	m.label.SetStyledText(ui.NewStyledText(fmt.Sprintf("%+v", msg)).Format(78))
	return nil
}

func (m *model) Draw() gruid.Grid {
	m.grid.Fill(gruid.Cell{Rune: ' '})
	m.label.Draw()
	return m.grid
}
