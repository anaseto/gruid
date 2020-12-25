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
	dr := tcell.NewDriver(tcell.Config{StyleManager: styler{}})
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

const (
	ColorCursor gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorTitle
	ColorPrompt
)

// styler implements the tcell.StyleManager interface.
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorCursor:
		ts = ts.Foreground(tc.ColorNavy).Reverse(true)
	case ColorTitle:
		ts = ts.Foreground(tc.ColorOlive)
	case ColorPrompt:
		ts = ts.Foreground(tc.ColorRed)
	}
	return ts
}

type model struct {
	grid  gruid.Grid    // the grid where the ui elements are drawn
	input *ui.TextInput // the text input widget
	label *ui.Label     // the label where we put activated text

	// Whether there where any changes in the input or not, so that we know
	// if it's necessary to redraw. It a completely unnecessary
	// optimization in this example, and just to show how such a thing can
	// be done.
	pass bool
}

func newModel(gd gruid.Grid) *model {
	m := &model{grid: gd}
	st := gruid.Style{}
	style := ui.TextInputStyle{
		Cursor: st.WithFg(ColorCursor),
		Prompt: st.WithFg(ColorPrompt),
	}
	input := ui.NewTextInput(ui.TextInputConfig{
		Grid:   gruid.NewGrid(20, 3),
		Box:    &ui.Box{Title: ui.NewStyledText("Text Input").WithStyle(st.WithFg(ColorTitle))},
		Prompt: "> ",
		Style:  style,
	})
	m.input = input
	label := &ui.Label{
		Box:         &ui.Box{Title: ui.NewStyledText("Last Entered Text").WithStyle(st.WithFg(ColorTitle))},
		StyledText:  ui.NewStyledText("Nothing entered yet!"),
		AdjustWidth: true,
	}
	m.label = label
	return m
}

// Update implements gruid.Model.Update.
func (m *model) Update(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgInit:
		return nil
	default:
		m.input.Update(msg)
	}
	m.pass = false
	switch m.input.Action() {
	case ui.TextInputPass:
		m.pass = true
	case ui.TextInputQuit:
		return gruid.End()
	case ui.TextInputInvoke:
		stt := ui.NewStyledText(m.input.Content()).Format(28)
		m.label.StyledText = stt
	}
	return nil
}

// Draw implements gruid.Model.Draw.
func (m *model) Draw() gruid.Grid {
	if !m.pass {
		m.grid.Fill(gruid.Cell{Rune: ' '})
		m.grid.Copy(m.input.Draw())
		m.label.Draw(m.grid.Slice(gruid.NewRange(0, 5, 30, 15)))
	}
	return m.grid
}
