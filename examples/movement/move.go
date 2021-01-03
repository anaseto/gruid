// This example program shows how to implement movement on a grid either on
// keyboard or mouse input. It implements both single-step movement and
// automatic movement in a direction or path.
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/paths"
	"github.com/anaseto/gruid/rl"
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
	ColorLOS
	ColorDark
)

const (
	Wall rl.Cell = iota
	Ground
)

// models represents our main application state.
type model struct {
	grid      gruid.Grid       // drawing grid
	playerPos gruid.Point      // tracks player position
	move      autoMove         // automatic movement
	pr        *paths.PathRange // path finding in the grid range
	path      []gruid.Point    // current path (reverse highlighting)
	mapgd     rl.Grid          // map grid
	rand      *rand.Rand       // random number generator
	fov       *rl.FOV
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
	case gruid.MsgInit:
		m.InitializeMap()
	case gruid.MsgKeyDown:
		return m.updateMsgKeyDown(msg)
	case gruid.MsgMouse:
		return m.updateMsgMouse(msg)
	case msgAutoMove:
		return m.updateMsgAutomove(msg)
	}
	return nil
}

func (m *model) InitializeMap() {
	m.mapgd = rl.NewGrid(80, 24)
	m.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	wlk := walker{rand: m.rand}
	wlk.neighbors = &paths.Neighbors{}
	mgen := rl.MapGen{Rand: m.rand, Grid: m.mapgd}
	mgen.RandomWalkCave(wlk, Ground, 0.5, 4)
	m.fov = rl.NewFOV(m.mapgd.Range())
	max := m.mapgd.Size()
	var p gruid.Point
	for {
		// find an empty starting position for the player
		p = gruid.Point{m.rand.Intn(max.X), m.rand.Intn(max.Y)}
		if m.mapgd.At(p) != Wall {
			break
		}
	}
	m.MovePlayer(p)
}

func (m *model) MovePlayer(to gruid.Point) {
	m.playerPos = to
	lt := &lighter{mapgd: m.mapgd}
	m.fov.VisionMap(lt, m.playerPos, maxLOS)
}

func (m *model) updateMsgKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	// cancel automatic movement on any key
	if m.autoMove() {
		m.stopAuto()
		return nil
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
		if m.grid.Contains(np) && m.mapgd.At(np) != Wall {
			m.MovePlayer(np)
			if msg.Mod&gruid.ModShift != 0 || strings.ToUpper(string(msg.Key)) == string(msg.Key) {
				// activate automatic movement in that direction
				m.move.delta = pdelta
				return automoveCmd(m.move.delta)
			}
		}
	}
	return nil
}

func (m *model) updateMsgMouse(msg gruid.MsgMouse) gruid.Effect {
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
	return nil
}

func (m *model) updateMsgAutomove(msg msgAutoMove) gruid.Effect {
	if m.move.delta != msg.delta {
		return nil
	}
	if m.move.path {
		if len(m.path) > 1 {
			return m.pathNext()
		}
	} else {
		np := m.playerPos.Add(msg.delta)
		if m.grid.Contains(np) && m.mapgd.At(np) != Wall {
			m.path = nil // remove path highlighting if any
			m.MovePlayer(np)
			// continue automatic movement in the same direction
			return automoveCmd(msg.delta)
		}
	}
	m.stopAuto()
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
	pp.mapgd = m.mapgd
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
	m.MovePlayer(p)
	return automoveCmd(m.move.delta)
}

// playerPath implements paths.Astar interface.
type playerPath struct {
	neighbors *paths.Neighbors
	mapgd     rl.Grid
}

func (pp *playerPath) Neighbors(p gruid.Point) []gruid.Point {
	return pp.neighbors.Cardinal(p, func(q gruid.Point) bool {
		return pp.mapgd.Contains(q) && pp.mapgd.At(q) != Wall
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

// walker implements rl.RandomWalker.
type walker struct {
	neighbors *paths.Neighbors
	rand      *rand.Rand
}

func (w walker) Neighbor(p gruid.Point) gruid.Point {
	neighbors := w.neighbors.Cardinal(p, func(q gruid.Point) bool {
		return true
	})
	return neighbors[w.rand.Intn(len(neighbors))]
}

// lighter implements rl.Lighter (in a very simple way).
type lighter struct {
	mapgd rl.Grid
}

const maxLOS = 8

func (lt *lighter) Cost(src, from, to gruid.Point) int {
	if src == from {
		return 0
	}
	if lt.mapgd.At(from) == Wall {
		return maxLOS
	}
	return 1
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
	max := m.grid.Size()
	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			p := gruid.Point{x, y}
			st := gruid.Style{}
			if cost, ok := m.fov.At(p); ok && cost < maxLOS {
				st = st.WithFg(ColorLOS)
			} else {
				st = st.WithBg(ColorDark)
			}
			switch {
			case p == m.playerPos:
				m.grid.Set(p, gruid.Cell{Rune: '@', Style: st.WithFg(ColorPlayer)})
			case m.mapgd.At(p) == Wall:
				m.grid.Set(p, gruid.Cell{Rune: '#', Style: st})
			case m.mapgd.At(p) == Ground:
				m.grid.Set(p, gruid.Cell{Rune: '.', Style: st})
			}
		}
	}
	for _, p := range m.path {
		c := m.grid.At(p)
		m.grid.Set(p, c.WithStyle(c.Style.WithBg(ColorPath)))
	}
	return m.grid
}
