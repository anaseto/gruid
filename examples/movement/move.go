package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/paths"
	"github.com/anaseto/gruid/ui"
)

func main() {

	// our application's state and grid with default config
	gd := gruid.NewGrid(80, 24)
	pr := paths.NewPathRange(gd.Range())
	m := &model{grid: gd, pr: pr}
	framebuf := &bytes.Buffer{} // for compressed recording

	// define new application
	app := gruid.NewApp(gruid.AppConfig{
		Driver:      driver,
		Model:       m,
		FrameWriter: framebuf,
	})

	// start application
	if err := app.Start(nil); err != nil {
		driver.Close()
		log.Fatal(err)
	}

	// launch replay just after the previous session
	fd, err := gruid.NewFrameDecoder(framebuf)
	if err != nil {
		log.Fatal(err)
	}
	gd = gruid.NewGrid(80, 24)
	rep := ui.NewReplay(ui.ReplayConfig{
		Grid:         gd,
		FrameDecoder: fd,
	})
	app = gruid.NewApp(gruid.AppConfig{
		Driver: driver,
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

// models represents our main application state.
type model struct {
	grid      gruid.Grid       // drawing grid
	playerPos gruid.Point      // tracks player position
	move      autoMove         // automatic movement
	pr        *paths.PathRange // path finding in the grid range
	path      []gruid.Point    // current path (reverse highlighting)
}

type autoMove struct {
	// diff represents a position variation such as (0,1), that
	// will be used in position arithmetic to move from one position to an
	// adjacent one in a certain direction.
	diff gruid.Point

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

		pdiff := gruid.Point{}
		switch msg.Key {
		case gruid.KeyArrowDown, "j", "J":
			pdiff = pdiff.Shift(0, 1)
		case gruid.KeyArrowLeft, "h", "H":
			pdiff = pdiff.Shift(-1, 0)
		case gruid.KeyArrowRight, "l", "L":
			pdiff = pdiff.Shift(1, 0)
		case gruid.KeyArrowUp, "k", "K":
			pdiff = pdiff.Shift(0, -1)
		case "Q", "q", gruid.KeyEscape:
			return gruid.End()
		}
		if pdiff.X != 0 || pdiff.Y != 0 {
			np := m.playerPos.Add(pdiff) //
			if m.grid.Contains(np) {
				m.playerPos = np
				if msg.Mod&gruid.ModShift != 0 || strings.ToUpper(string(msg.Key)) == string(msg.Key) {
					// activate automatic movement in that direction
					m.move.diff = pdiff
					return automoveCmd(m.move.diff)
				}
			}
		}
	case gruid.MsgMouse:
		switch msg.Action {
		case gruid.MouseMain:
			if m.autoMove() {
				m.stopAuto()
				m.pathAt(msg.P)
				break
			}
			if len(m.path) > 1 {
				return m.pathNext()
			}
		case gruid.MouseMove:
			if m.autoMove() {
				break
			}
			m.pathAt(msg.P)
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
			np := m.playerPos.Add(msg.diff)
			if m.grid.Contains(np) {
				m.path = nil // remove path highlighting if any
				m.playerPos = np
				return automoveCmd(msg.diff)
			}
		}
		m.stopAuto()
	}
	return nil
}

// autoMove checks whether automatic movement is activated.
func (m *model) autoMove() bool {
	p := gruid.Point{}
	return m.move.diff != p
}

// stopAuto resets automatic movement information.
func (m *model) stopAuto() {
	m.move = autoMove{}
	m.path = nil
}

// pathAt updates the path from player to a new position.
func (m *model) pathAt(p gruid.Point) {
	pp := &playerPath{}
	pp.neighbors = &paths.Neighbors{}
	m.path = m.pr.AstarPath(pp, m.playerPos, p)
}

// pathNext moves the player to next position in the path, and updates the path
// accordingly.
func (m *model) pathNext() gruid.Cmd {
	p := m.path[len(m.path)-2]
	m.path = m.path[:len(m.path)-1]
	m.move.path = true
	m.move.diff = p.Sub(m.playerPos)
	m.playerPos = p
	return automoveCmd(m.move.diff)
}

type msgAutoMove struct {
	diff gruid.Point
}

// automoveCmd returns a command that signals automatic movement in a given
// direction.
func automoveCmd(posdiff gruid.Point) gruid.Cmd {
	d := time.Millisecond * 30 // automatic movement time interval
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgAutoMove{posdiff}
	}
}

// playerPath implements paths.Astar interface.
type playerPath struct {
	neighbors *paths.Neighbors
}

func (pp *playerPath) Neighbors(p gruid.Point) []gruid.Point {
	return pp.neighbors.All(p, func(q gruid.Point) bool {
		// This is were in a real game we would filter non passable
		// neighbors, such as walls. For this example, we return always
		// true, as there are no obstacles.
		return true
	})
}

func (pp *playerPath) Cost(p, q gruid.Point) int {
	return 1
}

func (pp *playerPath) Estimation(p, q gruid.Point) int {
	// The manhattan distance corresponds here to the optimal distance and
	// is hence an acceptable estimation for astar.
	p = p.Sub(q)
	return abs(p.X) + abs(p.Y)
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
	m.grid.Range().Origin().Iter(func(p gruid.Point) {
		if p == m.playerPos {
			m.grid.Set(p, gruid.Cell{Rune: '@', Style: c.Style.WithFg(ColorPlayer)})
		} else {
			m.grid.Set(p, c)
		}
	})
	for _, p := range m.path {
		c := m.grid.At(p)
		m.grid.Set(p, c.WithStyle(c.Style.WithBg(ColorPath)))
	}
	return m.grid
}
