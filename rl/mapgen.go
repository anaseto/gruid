package rl

import (
	"math/rand"

	"github.com/anaseto/gruid"
)

// MapGen provides some grid-map generation facilities using a given math.Rand
// number generator.
type MapGen struct {
	Rand *rand.Rand // random number generator (required)
	Grid Grid       // destination grid slice where generated maps are drawn
}

func (mg MapGen) rand(n int) int {
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
// If more than one walk is done, the result is not guaranteed to be connex and
// has to be made connex later.
func (mp MapGen) RandomWalkCave(walker RandomWalker, c Cell, fillp float64, walks int) int {
	if fillp > 0.9 {
		fillp = 0.9
	}
	if fillp < 0.01 {
		fillp = 0.01
	}
	max := mp.Grid.Size()
	maxdigs := int(float64(max.X*max.Y) * fillp)
	wlkmax := maxdigs
	if walks > 0 {
		wlkmax /= walks
	}
	digged := 0
	mp.Grid.Iter(func(p gruid.Point, cc Cell) {
		// Compute number of cells already equal to c (in case some
		// other map generation occured before).
		if cc == c {
			digged++
		}
	})
	digs := digged
	for digs <= maxdigs {
		p := gruid.Point{mp.rand(max.X), mp.rand(max.Y)}
		sc := mp.Grid.At(p)
		if c == sc {
			continue
		}
		mp.Grid.Set(p, c)
		digs++
		wlkdigs := 1
		outDigs := 0
		lastInRange := p
		for digs < maxdigs && wlkdigs <= wlkmax {
			q := walker.Neighbor(p)
			if !mp.Grid.Contains(p) && mp.Grid.Contains(q) && mp.Grid.At(q) != c {
				p = lastInRange
				continue
			}
			p = q
			if mp.Grid.Contains(p) {
				if mp.Grid.At(p) != c {
					mp.Grid.Set(p, c)
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
