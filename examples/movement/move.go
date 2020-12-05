package main

import (
	"strings"
	"time"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	tc "github.com/gdamore/tcell/v2"
)

func main() {

	// use tcell terminal driver
	st := styler{}
	dri := &tcell.Driver{StyleManager: st}

	// our application's state and grid with default config
	gd := gruid.NewGrid(gruid.GridConfig{})
	m := &model{grid: gd}

	// define new application
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  m,
	})

	// start application
	app.Start()
}

const (
	ColorPlayer gruid.Color = 1 + iota // skip special zero value gruid.ColorDefault
)

// type that implements driver's style manager interface
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorPlayer:
		ts = ts.Foreground(tc.ColorNavy) // blue color for the player
	}
	return ts
}

// models represents our main application state.
type model struct {
	grid      gruid.Grid     // drawing grid
	playerPos gruid.Position // tracks player position
	autodir   gruid.Position // automatic movement in a given direction
}

// Init implements gruid.Model.Init. It does nothing.
func (m *model) Init() gruid.Cmd {
	return nil
}

func (m *model) Update(msg gruid.Msg) gruid.Cmd {
	d := time.Millisecond * 30 // automatic movement time interval
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		// posdiff represents a position variation such as (0,1), that
		// will be used in position arithmetic to move from Y to Y+1,
		// for example.
		posdiff := gruid.Position{}

		// cancel automatic movement on any key
		m.autodir = posdiff

		switch msg.Key {
		case gruid.KeyArrowDown, "j", "J":
			posdiff = posdiff.Shift(0, 1)
		case gruid.KeyArrowLeft, "h", "H":
			posdiff = posdiff.Shift(-1, 0)
		case gruid.KeyArrowRight, "l", "L":
			posdiff = posdiff.Shift(1, 0)
		case gruid.KeyArrowUp, "k", "K":
			posdiff = posdiff.Shift(0, -1)
		case "Q", "q", gruid.KeyEscape:
			return gruid.Quit
		}
		if posdiff.X != 0 || posdiff.Y != 0 {
			newpos := m.playerPos.Add(posdiff) //
			if m.grid.Contains(newpos) {
				m.playerPos = newpos
				if msg.Shift || strings.ToUpper(string(msg.Key)) == string(msg.Key) {
					// activate automatic movement in that direction
					m.autodir = posdiff
					return automoveCmd(posdiff, d)
				}
			}
		}
	case msgAutoMove:
		if m.autodir == msg.Diff {
			newpos := m.playerPos.Add(msg.Diff)
			if m.grid.Contains(newpos) {
				m.playerPos = newpos
				return automoveCmd(msg.Diff, d)
			}
		}
	}
	return nil
}

type msgAutoMove struct {
	Diff gruid.Position
}

// automoveCmd returns a command that signals automatic movement in a given
// direction.
func automoveCmd(posdiff gruid.Position, d time.Duration) gruid.Cmd {
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgAutoMove{posdiff}
	}
}

// Draw implements gruid.Model.Draw. It draws a simple map that spans the whole
// grid.
func (m *model) Draw() gruid.Grid {
	c := gruid.Cell{Rune: '.'} // default cell
	m.grid.Iter(func(pos gruid.Position) {
		if pos == m.playerPos {
			m.grid.SetCell(pos, gruid.Cell{Rune: '@', Style: c.Style.WithFg(ColorPlayer)})
		} else {
			m.grid.SetCell(pos, c)
		}
	})
	return m.grid
}
