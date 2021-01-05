package rl

import (
	"math/rand"

	"github.com/anaseto/gruid"
)

// MapGen provides some grid-map generation facilities using a given random
// number generator and a destination grid slice.
type MapGen struct {
	Rand *rand.Rand // random number generator (required)
	Grid Grid       // destination grid slice where generated maps are drawn
}

func (mg *MapGen) rand(n int) int {
	if n <= 0 {
		return 0
	}
	x := mg.Rand.Intn(n)
	return x
}

// RandomWalker describes the requirements for a random tunnel generator.
type RandomWalker interface {
	// Neighbor produces a random neighbor position.
	Neighbor(gruid.Point) gruid.Point
}

// RandomWalkCave draws a map in the destination grid using drunk walking. It
// performs a certain number of approximately equal length random walks,
// digging using the given cell, until a certain filling percentage (given by a
// float between 0 and 1) is reached. It returns the number of digged cells.
// If more than one walk is done, the result is not guaranteed to be connected
// and has to be made connected later.
func (mg *MapGen) RandomWalkCave(walker RandomWalker, c Cell, fillp float64, walks int) int {
	if fillp > 0.9 {
		fillp = 0.9
	}
	if fillp < 0.01 {
		fillp = 0.01
	}
	max := mg.Grid.Size()
	maxdigs := int(float64(max.X*max.Y) * fillp)
	wlkmax := maxdigs
	if walks > 0 {
		wlkmax /= walks
	}
	digged := 0
	mg.Grid.Iter(func(p gruid.Point, cc Cell) {
		// Compute number of cells already equal to c (in case some
		// other map generation occurred before).
		if cc == c {
			digged++
		}
	})
	digs := digged
	for digs <= maxdigs {
		p := gruid.Point{mg.rand(max.X), mg.rand(max.Y)}
		sc := mg.Grid.At(p)
		if c == sc {
			continue
		}
		mg.Grid.Set(p, c)
		digs++
		wlkdigs := 1
		outDigs := 0
		lastInRange := p
		for digs < maxdigs && wlkdigs <= wlkmax {
			q := walker.Neighbor(p)
			if !mg.Grid.Contains(p) && mg.Grid.Contains(q) && mg.Grid.At(q) != c {
				p = lastInRange
				continue
			}
			p = q
			if mg.Grid.Contains(p) {
				if mg.Grid.At(p) != c {
					mg.Grid.Set(p, c)
					digs++
					wlkdigs++
				}
				lastInRange = p
			} else {
				outDigs++
			}
			if outDigs > wlkmax {
				outDigs = 0
				p = lastInRange
			}
		}
	}
	return digs - digged
}

// CellularAutomataRule describes the conditions for generating a wall in an
// iteration of the cellular automata algorithm. In the comments, W(n) is the
// number of walls in an n-radius area centered at a certain position being
// processed.
type CellularAutomataRule struct {
	WCutoff1        int  // wall if W(1) >= WCutoff1 (set to 0 to disable)
	WCutoff2        int  // wall if W(2) <= WCutoff2 (set to 25 to disable)
	WallsOutOfRange bool // count out of range positions as walls
	Reps            int  // number of successive iterations of this rule
}

// CellularAutomataCave generates a map using a cellular automata algorithm.
// You can provide the initial wall filling percentage, and then sets of rules
// to be applied in a certain number of iterations. The result is not
// guaranteed to be connected.
//
// The algorithm is based on:
// http://www.roguebasin.com/index.php?title=Cellular_Automata_Method_for_Generating_Random_Cave-Like_Levels
func (mg *MapGen) CellularAutomataCave(wall, ground Cell, winit float64, rules []CellularAutomataRule) int {
	if winit > 0.9 {
		winit = 0.9
	}
	if winit < 0.1 {
		winit = 0.1
	}
	max := mg.Grid.Size()
	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			p := gruid.Point{x, y}
			f := mg.Rand.Float64()
			if f < winit {
				mg.Grid.Set(p, wall)
			} else {
				mg.Grid.Set(p, ground)
			}
		}
	}
	bufgd := NewGrid(max.X, max.Y)
	for _, rule := range rules {
		if rule.WCutoff2 >= 25 {
			mg.applyRuleWithoutW2(wall, ground, bufgd, rule)
		} else if rule.WCutoff1 <= 0 {
			mg.applyRuleWithoutW1(wall, ground, bufgd, rule)
		} else {
			mg.applyRule(wall, ground, bufgd, rule)
		}
	}
	count := 0
	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			p := gruid.Point{x, y}
			if mg.Grid.At(p) == ground {
				count++
			}
		}
	}
	return count
}

func (mg *MapGen) applyRule(wall, ground Cell, bufgd Grid, rule CellularAutomataRule) {
	max := mg.Grid.Size()
	for i := 0; i < rule.Reps; i++ {
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := gruid.Point{x, y}
				c1 := mg.countWalls(p, wall, 1, rule.WallsOutOfRange)
				c2 := mg.countWalls(p, wall, 2, rule.WallsOutOfRange)
				if c1 >= rule.WCutoff1 || c2 <= rule.WCutoff2 {
					bufgd.Set(p, wall)
				} else {
					bufgd.Set(p, ground)
				}
			}
		}
		mg.Grid.Copy(bufgd)
	}
}

func (mg *MapGen) applyRuleWithoutW1(wall, ground Cell, bufgd Grid, rule CellularAutomataRule) {
	max := mg.Grid.Size()
	// optimization equivalent to disabling WCutoff1
	for i := 0; i < rule.Reps; i++ {
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := gruid.Point{x, y}
				c2 := mg.countWalls(p, wall, 2, rule.WallsOutOfRange)
				if c2 <= rule.WCutoff2 {
					bufgd.Set(p, wall)
				} else {
					bufgd.Set(p, ground)
				}
			}
		}
		mg.Grid.Copy(bufgd)
	}
}

func (mg *MapGen) applyRuleWithoutW2(wall, ground Cell, bufgd Grid, rule CellularAutomataRule) {
	max := mg.Grid.Size()
	// optimization equivalent to disabling WCutoff2
	for i := 0; i < rule.Reps; i++ {
		for y := 0; y < max.Y; y++ {
			for x := 0; x < max.X; x++ {
				p := gruid.Point{x, y}
				c1 := mg.countWalls(p, wall, 1, rule.WallsOutOfRange)
				if c1 >= rule.WCutoff1 {
					bufgd.Set(p, wall)
				} else {
					bufgd.Set(p, ground)
				}
			}
		}
		mg.Grid.Copy(bufgd)
	}
}

func (mg *MapGen) countWalls(p gruid.Point, w Cell, radius int, countOut bool) int {
	count := 0
	for y := p.Y - radius; y <= p.Y+radius; y++ {
		for x := p.X - radius; x <= p.X+radius; x++ {
			q := gruid.Point{x, y}
			if !mg.Grid.Contains(q) {
				if countOut {
					count++
				}
				continue
			}
			if mg.Grid.At(q) == w {
				count++
			}
		}
	}
	return count
}
