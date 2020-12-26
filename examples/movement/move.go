// This example program shows how to implement movement on a grid either on
// keyboard or mouse input. It implements both single-step movement and
// automatic movement in a direction or path.
package main

import (
	"bytes"
	"context"
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
	pr := paths.NewPathRange(gd.Bounds())
	m := &model{grid: gd, pr: pr}
	framebuf := &bytes.Buffer{} // for compressed recording

	// define new application
	app := gruid.NewApp(gruid.AppConfig{
		Driver:      driver,
		Model:       m,
		FrameWriter: framebuf,
	})

	// start application
	if err := app.Start(context.Background()); err != nil {
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
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successful quit.")
	}
}

// Those constants represent the generic colors we use in this example.
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

// autoMove represents the information for an automatic-movement step.
type autoMove struct {
	// delta represents a position variation such as (0,1), that
	// will be used in position arithmetic to move from one position to an
	// adjacent one in a certain direction.
	delta gruid.Point

	path bool // whether following a path (instead of a simple direction)
}

// msgAutoMove is used to ask Update to move the player's position by delta.
type msgAutoMove struct {
	delta gruid.Point
}

// Update implements gruid.Model.Update. It handles keyboard and mouse input
// messages and updates the model in response to them.
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

		pdelta := gruid.Point{}
		switch msg.Key {
		case gruid.KeyArrowDown, "j", "J":
			pdelta = pdelta.Shift(0, 1)
		case gruid.KeyArrowLeft, "h", "H":
			pdelta = pdelta.Shift(-1, 0)
		case gruid.KeyArrowRight, "l", "L":
			pdelta = pdelta.Shift(1, 0)
		case gruid.KeyArrowUp, "k", "K":
			pdelta = pdelta.Shift(0, -1)
		case "Q", "q", gruid.KeyEscape:
			return gruid.End()
		}
		if pdelta.X != 0 || pdelta.Y != 0 {
			np := m.playerPos.Add(pdelta) //
			if m.grid.Contains(np) {
				m.playerPos = np
				if msg.Mod&gruid.ModShift != 0 || strings.ToUpper(string(msg.Key)) == string(msg.Key) {
					// activate automatic movement in that direction
					m.move.delta = pdelta
					return automoveCmd(m.move.delta)
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
		if m.move.delta != msg.delta {
			break
		}
		if m.move.path {
			if len(m.path) > 1 {
				return m.pathNext()
			}
		} else {
			np := m.playerPos.Add(msg.delta)
			if m.grid.Contains(np) {
				m.path = nil // remove path highlighting if any
				m.playerPos = np
				// continue automatic movement in the same direction
				return automoveCmd(msg.delta)
			}
		}
		m.stopAuto()
	}
	return nil
}

// automoveCmd returns a command that signals automatic movement in a given
// direction.
func automoveCmd(pdelta gruid.Point) gruid.Cmd {
	d := time.Millisecond * 30 // automatic movement time interval
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgAutoMove{pdelta}
	}
}

// autoMove checks whether automatic movement is activated.
func (m *model) autoMove() bool {
	p := gruid.Point{}
	return m.move.delta != p
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

// pathNext moves the player to next position in the path, updates the path
// accordingly, and returns a command that will deliver the message for the
// next automatic movement step along the path.
func (m *model) pathNext() gruid.Cmd {
	p := m.path[1]
	m.path = m.path[1:]
	m.move.path = true
	m.move.delta = p.Sub(m.playerPos)
	m.playerPos = p
	return automoveCmd(m.move.delta)
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
	m.grid.Fill(gruid.Cell{Rune: '.'})
	m.grid.Set(m.playerPos, gruid.Cell{Rune: '@', Style: gruid.Style{}.WithFg(ColorPlayer)})
	for _, p := range m.path {
		c := m.grid.At(p)
		m.grid.Set(p, c.WithStyle(c.Style.WithBg(ColorPath)))
	}
	return m.grid
}
