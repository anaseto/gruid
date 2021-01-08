package rl

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/paths"
)

// MapGen provides some grid-map generation facilities using a given random
// number generator and a destination grid slice.
type MapGen struct {
	// Rand is the random number generator to be used in map generation.
	Rand *rand.Rand

	// Grid is the destination grid slice where generated maps are drawn.
	Grid Grid
}

// WithGrid returns a derived MapGen using the given destination grid slice.
func (mg MapGen) WithGrid(gd Grid) MapGen {
	mg.Grid = gd
	return mg
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
// If more than one walk is done, the result is not guaranteed to be connected
// and has to be made connected later.
func (mg MapGen) RandomWalkCave(walker RandomWalker, c Cell, fillp float64, walks int) int {
	if fillp > 0.9 {
		fillp = 0.9
	}
	if fillp < 0.01 {
		fillp = 0.01
	}
	max := mg.Grid.Size()
	maxdigs := int(float64(max.X*max.Y) * fillp)
	digged := mg.Grid.Count(c)
	digs := digged
	wlkmax := maxdigs - digged
	if walks > 0 {
		wlkmax /= walks
	}
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
		for digs <= maxdigs && wlkdigs <= wlkmax {
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
			if outDigs > wlkmax || outDigs > 150 {
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
func (mg MapGen) CellularAutomataCave(wall, ground Cell, winit float64, rules []CellularAutomataRule) int {
	if winit > 0.9 {
		winit = 0.9
	}
	if winit < 0.1 {
		winit = 0.1
	}
	max := mg.Grid.Size()
	mg.Grid.FillFunc(func() Cell {
		f := mg.Rand.Float64()
		if f < winit {
			return wall
		}
		return ground
	})
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
	count := mg.Grid.Count(ground)
	return count
}

func (mg MapGen) applyRule(wall, ground Cell, bufgd Grid, rule CellularAutomataRule) {
	for i := 0; i < rule.Reps; i++ {
		bufgd.Map(func(p gruid.Point, c Cell) Cell {
			c1 := mg.countWalls(p, wall, 1, rule.WallsOutOfRange)
			c2 := mg.countWalls(p, wall, 2, rule.WallsOutOfRange)
			if c1 >= rule.WCutoff1 || c2 <= rule.WCutoff2 {
				return wall
			}
			return ground
		})
		mg.Grid.Copy(bufgd)
	}
}

func (mg MapGen) applyRuleWithoutW1(wall, ground Cell, bufgd Grid, rule CellularAutomataRule) {
	// optimization equivalent to disabling WCutoff1
	for i := 0; i < rule.Reps; i++ {
		bufgd.Map(func(p gruid.Point, c Cell) Cell {
			c2 := mg.countWalls(p, wall, 2, rule.WallsOutOfRange)
			if c2 <= rule.WCutoff2 {
				return wall
			}
			return ground
		})
		mg.Grid.Copy(bufgd)
	}
}

func (mg MapGen) applyRuleWithoutW2(wall, ground Cell, bufgd Grid, rule CellularAutomataRule) {
	// optimization equivalent to disabling WCutoff2
	for i := 0; i < rule.Reps; i++ {
		bufgd.Map(func(p gruid.Point, c Cell) Cell {
			c1 := mg.countWalls(p, wall, 1, rule.WallsOutOfRange)
			if c1 >= rule.WCutoff1 {
				return wall
			}
			return ground
		})
		mg.Grid.Copy(bufgd)
	}
}

func (mg MapGen) countWalls(p gruid.Point, w Cell, radius int, countOut bool) int {
	count := 0
	rg := gruid.Range{
		gruid.Point{p.X - radius, p.Y - radius},
		gruid.Point{p.X + radius + 1, p.Y + radius + 1},
	}
	if countOut {
		osize := rg.Size()
		rg = rg.Intersect(mg.Grid.Range())
		size := rg.Size()
		count += osize.X*osize.Y - size.X*size.Y
	} else {
		rg = rg.Intersect(mg.Grid.Range())
	}
	gd := mg.Grid.Slice(rg)
	count += gd.Count(w)
	return count
}

// KeepCC puts walls in all the positions unreachable from p according to last
// ComputeCC or ComputeCCAll call on pr. Paths are supposed to be
// bidirectional. It returns the number of cells in the remaining connected
// component.
func (mg MapGen) KeepCC(pr *paths.PathRange, p gruid.Point, wall Cell) int {
	max := mg.Grid.Size()
	count := 0
	id := pr.CCAt(p)
	if id == -1 {
		mg.Grid.Fill(wall)
		return 0
	}
	for y := 0; y < max.Y; y++ {
		for x := 0; x < max.X; x++ {
			q := gruid.Point{x, y}
			if pr.CCAt(q) != id {
				mg.Grid.Set(q, wall)
			} else {
				count++
			}
		}
	}
	return count
}

// Vault represents a prefabricated room or level section built from a textual
// description using Parse.
type Vault struct {
	content string
	size    gruid.Point
	runes   string
}

// Content returns the vault's textual content.
func (v *Vault) Content() string {
	return v.content
}

// Size returns the (width, height) size of the vault in cells.
func (v *Vault) Size() gruid.Point {
	return v.size
}

// SetRunes states that the permitted runes in the textual vault content should
// be among the runes in the given string. If empty, any rune is allowed.
func (v *Vault) SetRunes(s string) {
	v.runes = s
}

// Runes returns a string containing the currently permitted runes in the
// textual vault content.
func (v *Vault) Runes() string {
	return v.runes
}

// Parse updates the vault's textual content. Each line in the string should
// have the same length (leading and trailing spaces are removed
// automatically). Only characters defined by SetRunes are allowed in the
// textual content.
func (v *Vault) Parse(s string) error {
	x, y := 0, 0
	w := -1
	s = strings.TrimSpace(s)
	for _, r := range s {
		if r == '\n' {
			if x > w {
				if w > 0 {
					return fmt.Errorf("vault: inconsistent size:\n%s", s)
				}
				w = x
			}
			x = 0
			y++
			continue
		}
		if v.runes != "" && !strings.ContainsRune(v.runes, r) {
			return fmt.Errorf("vault contains invalid rune “%c” at %v", r, gruid.Point{x, y})
		}
		x++
	}
	if x > w {
		if w > 0 {
			return fmt.Errorf("vault: inconsistent size:\n%s", s)
		}
		w = x
	}
	if w > 0 || y > 0 {
		y++ // at least one line
	}
	v.content = s
	v.size = gruid.Point{x, y}
	return nil
}

// Iter iterates a function for all the vault positions and content runes.
func (v *Vault) Iter(fn func(gruid.Point, rune)) {
	x, y := 0, 0
	for _, r := range v.content {
		if r == '\n' {
			x = 0
			y++
			continue
		}
		fn(gruid.Point{x, y}, r)
		x++
	}
}

// Draw uses a mapping from runes to cells to draw the vault into a grid. It
// returns the grid slice that was drawn.
func (v *Vault) Draw(gd Grid, fn func(rune) Cell) Grid {
	x, y := 0, 0
	for _, r := range v.content {
		if r == '\n' {
			x = 0
			y++
			continue
		}
		gd.Set(gruid.Point{x, y}, fn(r))
		x++
	}
	return gd.Slice(gruid.NewRange(0, 0, v.size.X, v.size.Y))
}

// Reflect changes the content with its reflection with respect to a middle
// vertical axis (order of characters in each line reversed). The result has
// the same size.
func (v *Vault) Reflect() {
	sb := strings.Builder{}
	sb.Grow(len(v.content))
	line := make([]rune, 0, v.Size().X)
	for _, r := range v.content {
		if r == '\n' {
			for i := len(line) - 1; i >= 0; i-- {
				sb.WriteRune(line[i])
			}
			sb.WriteRune('\n')
			line = line[:0]
			continue
		}
		line = append(line, r)
	}
	for i := len(line) - 1; i >= 0; i-- {
		sb.WriteRune(line[i])
	}
	v.content = sb.String()
}

// Rotate rotates the vault content n times by 90 degrees counter-clockwise (or
// clockwise for negative n values).  The result's size has dimensions
// exchanged for odd n.
func (v *Vault) Rotate(n int) {
	n %= 4
	if n < 0 {
		n = 4 + n
	}
	switch n {
	case 1:
		v.rotate90()
	case 2:
		v.rotate180()
	case 3:
		v.rotate180()
		v.rotate90()
	}
}

func (v *Vault) rotate90() {
	lines := strings.Split(v.content, "\n")
	runelines := make([][]rune, len(lines))
	for i, s := range lines {
		runelines[i] = []rune(s)
	}
	sb := strings.Builder{}
	sb.Grow(len(v.content))
	max := v.size
	for x := 0; x < max.X; x++ {
		for y := 0; y < max.Y; y++ {
			sb.WriteRune(runelines[y][max.X-x-1])
		}
		if x < max.X-1 {
			sb.WriteRune('\n')
		}
	}
	v.content = sb.String()
	v.size.X, v.size.Y = v.size.Y, v.size.X
}

// rotate180 rotates the vault by 180 degrees. It can be obtained with two
// rotate90, but it's just a simple string reversal, so we make a special case
// for it.
func (v *Vault) rotate180() {
	runes := []rune(v.content)
	for i := 0; i < len(runes)/2; i++ {
		j := len(runes) - 1 - i
		runes[i], runes[j] = runes[j], runes[i]
	}
	v.content = string(runes)
}
