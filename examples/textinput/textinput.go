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
	var gd = gruid.NewGrid(80, 24)
	st := styler{}
	var dri = tcell.NewDriver(tcell.Config{StyleManager: st})
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
	ColorCursor gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorTitle
	ColorPrompt
)

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
	grid  gruid.Grid
	input *ui.TextInput
	label *ui.Label
}

func NewModel(gd gruid.Grid) *model {
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
	label := ui.NewLabel(ui.LabelConfig{
		Grid:        gruid.NewGrid(30, 10),
		Box:         &ui.Box{Title: ui.NewStyledText("Last Entered Text").WithStyle(st.WithFg(ColorTitle))},
		StyledText:  ui.NewStyledText("Nothing entered yet!"),
		AdjustWidth: true,
	})
	m.label = label
	return m
}

func (m *model) Update(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgInit:
	default:
		m.input.Update(msg)
	}
	switch m.input.Action() {
	case ui.TextInputQuit:
		return gruid.End()
	case ui.TextInputActivate:
		stt := ui.NewStyledText(m.input.Content()).Format(28)
		m.label.SetStyledText(stt)
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	m.grid.Fill(gruid.Cell{Rune: ' '})
	m.grid.Copy(m.input.Draw())
	m.grid.Slice(m.grid.Range().Shift(0, 6, 0, 0)).Copy(m.label.Draw())
	return m.grid
}
