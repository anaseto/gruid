package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	"github.com/anaseto/gruid/paths"
	"github.com/anaseto/gruid/ui"
	tc "github.com/gdamore/tcell/v2"
)

func main() {

	// use tcell terminal driver
	st := styler{}
	dri := &tcell.Driver{StyleManager: st}

	// our application's state and grid with default config
	gd := gruid.NewGrid(gruid.GridConfig{})
	pr := paths.NewPathRange(gd.Range())
	m := &model{grid: gd, pr: pr}
	framebuf := &bytes.Buffer{} // for compressed recording

	// define new application
	app := gruid.NewApp(gruid.AppConfig{
		Driver:      dri,
		Model:       m,
		FrameWriter: framebuf,
	})

	// start application
	if err := app.Start(nil); err != nil {
		log.Fatal(err)
	}

	// launch replay just after the previous session
	fd, err := gruid.NewFrameDecoder(framebuf)
	if err != nil {
		log.Fatal(err)
	}
	gd = gruid.NewGrid(gruid.GridConfig{})
	rep := ui.NewReplay(ui.ReplayConfig{
		Grid:         gd,
		FrameDecoder: fd,
	})
	app = gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  rep,
	})
	if err := app.Start(nil); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successful quit.")
	}
}

const (
	ColorPlayer gruid.Color = 1 + iota // skip special zero value gruid.ColorDefault
	ColorPath
)

// type that implements driver's style manager interface
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorPlayer:
		ts = ts.Foreground(tc.ColorNavy) // blue color for the player
	}
	switch st.Bg {
	case ColorPath:
		ts = ts.Reverse(true)
	}
	return ts
}

// models represents our main application state.
type model struct {
	grid      gruid.Grid       // drawing grid
	playerPos gruid.Position   // tracks player position
	move      autoMove         // automatic movement
	pr        *paths.PathRange // path finding in the grid range
	path      []gruid.Position // current path (reverse highlighting)
}

type autoMove struct {
	// diff represents a position variation such as (0,1), that
	// will be used in position arithmetic to move from one position to an
	// adjacent one in a certain direction.
	diff gruid.Position

	path bool // whether following a path (instead of a simple direction)
}

func (m *model) Update(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		// cancel automatic movement on any key
		if m.autoMove() {
			m.stopAuto()
			break
		}

		// remove mouse path highlighting
		m.path = nil

		posdiff := gruid.Position{}
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
			return gruid.End()
		}
		if posdiff.X != 0 || posdiff.Y != 0 {
			newpos := m.playerPos.Add(posdiff) //
			if m.grid.Contains(newpos) {
				m.playerPos = newpos
				if msg.Mod&gruid.ModShift != 0 || strings.ToUpper(string(msg.Key)) == string(msg.Key) {
					// activate automatic movement in that direction
					m.move.diff = posdiff
					return automoveCmd(m.move.diff)
				}
			}
		}
	case gruid.MsgMouse:
		switch msg.Action {
		case gruid.MouseMain:
			if m.autoMove() {
				m.stopAuto()
				m.pathAt(msg.MousePos)
				break
			}
			if len(m.path) > 1 {
				return m.pathNext()
			}
		case gruid.MouseMove:
			if m.autoMove() {
				break
			}
			m.pathAt(msg.MousePos)
		}
	case msgAutoMove:
		if m.move.diff != msg.diff {
			break
		}
		if m.move.path {
			if len(m.path) > 1 {
				return m.pathNext()
			}
		} else {
			newpos := m.playerPos.Add(msg.diff)
			if m.grid.Contains(newpos) {
				m.path = nil // remove path highlighting if any
				m.playerPos = newpos
				return automoveCmd(msg.diff)
			}
		}
		m.stopAuto()
	}
	return nil
}

// autoMove checks whether automatic movement is activated.
func (m *model) autoMove() bool {
	pos := gruid.Position{}
	return m.move.diff != pos
}

// stopAuto resets automatic movement information.
func (m *model) stopAuto() {
	m.move = autoMove{}
	m.path = nil
}

// pathAt updates the path from player to a new position.
func (m *model) pathAt(pos gruid.Position) {
	p := &pather{}
	p.neighbors = &paths.Neighbors{}
	m.path = m.pr.AstarPath(p, m.playerPos, pos)
}

// pathNext moves the player to next position in the path, and updates the path
// accordingly.
func (m *model) pathNext() gruid.Cmd {
	pos := m.path[len(m.path)-2]
	m.path = m.path[:len(m.path)-1]
	m.move.path = true
	m.move.diff = pos.Sub(m.playerPos)
	m.playerPos = pos
	return automoveCmd(m.move.diff)
}

type msgAutoMove struct {
	diff gruid.Position
}

// automoveCmd returns a command that signals automatic movement in a given
// direction.
func automoveCmd(posdiff gruid.Position) gruid.Cmd {
	d := time.Millisecond * 30 // automatic movement time interval
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgAutoMove{posdiff}
	}
}

// pather implements paths.Astar interface.
type pather struct {
	neighbors *paths.Neighbors
}

func (p *pather) Neighbors(pos gruid.Position) []gruid.Position {
	return p.neighbors.All(pos, func(pos gruid.Position) bool {
		// This is were in a real game we would filter non passable
		// neighbors, such as walls. For this example, we return always
		// true, as there are no obstacles.
		return true
	})
}

func (p *pather) Cost(pos, npos gruid.Position) int {
	return 1
}

func (p *pather) Estimation(pos, npos gruid.Position) int {
	// The manhattan distance corresponds here to the optimal distance and
	// is hence an acceptable estimation for astar.
	pos = pos.Sub(npos)
	return abs(pos.X) + abs(pos.Y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Draw implements gruid.Model.Draw. It draws a simple map that spans the whole
// grid.
func (m *model) Draw() gruid.Grid {
	c := gruid.Cell{Rune: '.'} // default cell
	m.grid.Range().Relative().Iter(func(pos gruid.Position) {
		if pos == m.playerPos {
			m.grid.SetCell(pos, gruid.Cell{Rune: '@', Style: c.Style.WithFg(ColorPlayer)})
		} else {
			m.grid.SetCell(pos, c)
		}
	})
	for _, pos := range m.path {
		c := m.grid.GetCell(pos)
		m.grid.SetCell(pos, c.WithStyle(c.Style.WithBg(ColorPath)))
	}
	return m.grid
}
