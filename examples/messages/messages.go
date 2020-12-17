package main

import (
	"fmt"
	"log"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/ui"
)

func main() {
	var gd = gruid.NewGrid(gruid.GridConfig{})
	m := NewModel(gd)
	app := gruid.NewApp(gruid.AppConfig{
		Driver: driver,
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
)

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
	if _, ok := msg.(gruid.MsgQuit); ok {
		// The user requested the end of the application (for example
		// by closing the window).
		return gruid.End()
	}
	if msg, ok := msg.(gruid.MsgKeyDown); ok {
		switch msg.Key {
		case gruid.KeyEscape, "Q", "q":
			return gruid.End()
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
