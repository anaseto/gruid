// This example program shows how to implement movement on a grid either on
// keyboard or mouse input. It implements both single-step movement and
// automatic movement in a direction or path, and provides some simple map
// generation and field of vision.
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
	ColorLOS
	ColorDark
)

// Those constants represent styling attributes.
const (
	AttrNone gruid.AttrMask = iota
	AttrReverse
)

// Those constants represent the different types of terrains in the map grid.
// We use the second bit for marking a cell explored or not.
const (
	Wall rl.Cell = iota
	Ground
	Explored rl.Cell = 0b10
)

func cell(c rl.Cell) rl.Cell {
	return c &^ Explored
}

func explored(c rl.Cell) bool {
	return c&Explored != 0
}

// models represents our main application state.
type model struct {
	grid      gruid.Grid       // drawing grid
	playerPos gruid.Point      // tracks player position
	move      autoMove         // automatic movement
	pr        *paths.PathRange // path finding in the grid range
	path      []gruid.Point    // current path (reverse highlighting)
	mapgd     rl.Grid          // map grid
	rand      *rand.Rand       // random number generator
	fov       *rl.FOV          // field of vision
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
	mgen := rl.MapGen{Rand: m.rand, Grid: m.mapgd}
	if m.rand.Float64() > 0.5 {
		mgen.RandomWalkCave(wlk, Ground, 0.5, 1)
	} else {
		rules := []rl.CellularAutomataRule{
			{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
			{WCutoff1: 5, WCutoff2: 25, Reps: 3, WallsOutOfRange: true},
		}
		mgen.CellularAutomataCave(Wall, Ground, 0.40, rules)
	}
	max := m.mapgd.Size()
	var p gruid.Point
	for {
		// find an empty starting position for the player
		p = gruid.Point{m.rand.Intn(max.X), m.rand.Intn(max.Y)}
		if cell(m.mapgd.At(p)) != Wall {
			break
		}
	}
	m.fov = rl.NewFOV(gruid.NewRange(-maxLOS, -maxLOS, maxLOS+1, maxLOS+1))
	m.MovePlayer(p)
}

func (m *model) MovePlayer(to gruid.Point) {
	// We shift the FOV's Range so that it will be centered on the new
	// player's position. We could have simply used the whole map for the
	// range, though it would have used a little bit more memory (not
	// important here, just for showing what can be done).
	m.fov.SetRange(m.fov.Range().Add(to.Sub(m.playerPos)))
	m.playerPos = to
	lt := &lighter{mapgd: m.mapgd}
	m.fov.VisionMap(lt, m.playerPos, maxLOS)
	rg := m.fov.Range()
	// We mark cells in field of view as explored.
	m.fov.Iter(func(ln rl.LightNode) {
		if ln.Cost >= maxLOS {
			return
		}
		p := ln.P.Add(rg.Min)
		c := m.mapgd.At(p)
		if !explored(c) {
			m.mapgd.Set(p, c|Explored)
		}
	})
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
		if m.grid.Contains(np) && cell(m.mapgd.At(np)) != Wall {
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
			m.pathSet(msg.P)
			break
		}
		if len(m.path) > 1 {
			return m.pathNext()
		}
	case gruid.MouseMove:
		if m.autoMove() {
			break
		}
		m.pathSet(msg.P)
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
		if m.grid.Contains(np) && cell(m.mapgd.At(np)) != Wall {
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

// pathSet updates the path from player to a new position.
func (m *model) pathSet(p gruid.Point) {
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
		if !pp.mapgd.Contains(q) {
			return false
		}
		c := pp.mapgd.At(q)
		return explored(c) && cell(c) != Wall
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
	rand *rand.Rand
}

// Neighbor returns a random neighbor position, favoring horizontal directions
// (because the maps we use are longer in that direction).
func (w walker) Neighbor(p gruid.Point) gruid.Point {
	switch w.rand.Intn(6) {
	case 0, 1:
		return p.Shift(1, 0)
	case 2, 3:
		return p.Shift(-1, 0)
	case 4:
		return p.Shift(0, 1)
	default:
		return p.Shift(0, -1)
	}
}

// lighter implements rl.Lighter in a simple way.
type lighter struct {
	mapgd rl.Grid
}

const maxLOS = 10

func (lt *lighter) Cost(src, from, to gruid.Point) int {
	if cell(lt.mapgd.At(from)) == Wall || lt.diagonalWalls(from, to) {
		return maxLOS
	}
	if src == from {
		return 0
	}
	if lt.diagonalStep(from, to) {
		return 2
	}
	return 1
}

// diagonalWalls checks whether diagonal light has to be blocked (this would
// not be necessary if 8-way movement geometry is chosen).
func (lt *lighter) diagonalWalls(from, to gruid.Point) bool {
	if !lt.diagonalStep(from, to) {
		return false
	}
	step := to.Sub(from)
	return cell(lt.mapgd.At(from.Add(gruid.Point{X: step.X}))) == Wall &&
		cell(lt.mapgd.At(from.Add(gruid.Point{Y: step.Y}))) == Wall
}

// diagonalStep reports whether it is a diagonal step.
func (lt *lighter) diagonalStep(from, to gruid.Point) bool {
	step := to.Sub(from)
	return step.X != 0 && step.Y != 0
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
	m.mapgd.Iter(func(p gruid.Point, c rl.Cell) {
		st := gruid.Style{}
		if cost, ok := m.fov.At(p); ok && cost < maxLOS {
			st = st.WithFg(ColorLOS)
		} else {
			st = st.WithBg(ColorDark)
		}
		switch {
		case p == m.playerPos:
			m.grid.Set(p, gruid.Cell{Rune: '@', Style: st.WithFg(ColorPlayer)})
		case !explored(c):
			m.grid.Set(p, gruid.Cell{Rune: ' ', Style: st})
		case cell(c) == Wall:
			m.grid.Set(p, gruid.Cell{Rune: '#', Style: st})
		case cell(c) == Ground:
			m.grid.Set(p, gruid.Cell{Rune: '.', Style: st})
		}
	})
	for _, p := range m.path {
		c := m.grid.At(p)
		m.grid.Set(p, c.WithStyle(c.Style.WithAttrs(AttrReverse)))
	}
	return m.grid
}
