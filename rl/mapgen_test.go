package rl

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/paths"
)

const vaultExample = `
#.#...
......
..####`

const vaultExampleRotated180 = `
####..
......
...#.#`

const vaultExampleRotated = `
..#
..#
..#
#.#
...
#..`

const vaultExampleReflected = `
...#.#
......
####..`

func TestVault(t *testing.T) {
	v := &Vault{}
	err := v.Parse(vaultExample)
	if err != nil {
		t.Errorf("Parse: %v", err)
	}
	if v.Size().X != 6 || v.Size().Y != 3 {
		t.Errorf("bad size: %v", v.Size())
	}
	v.Rotate(1)
	if v.Size().X != 3 || v.Size().Y != 6 {
		t.Errorf("bad size: %v", v.Size())
	}
	if v.Content() != strings.TrimSpace(vaultExampleRotated) {
		t.Errorf("bad rotation 1:`%v`", v.Content())
	}
	v.Rotate(-1)
	if v.Content() != strings.TrimSpace(vaultExample) {
		t.Errorf("bad rotation -1:`%v`", v.Content())
	}
	v.Rotate(2)
	if v.Content() != strings.TrimSpace(vaultExampleRotated180) {
		t.Errorf("bad rotation 2:`%v`", v.Content())
	}
	v.Rotate(2)
	v.Reflect()
	if v.Content() != strings.TrimSpace(vaultExampleReflected) {
		t.Errorf("bad reflection:`%v`", v.Content())
	}
}

func TestVaultIter(t *testing.T) {
	v, err := NewVault(vaultExample)
	if err != nil {
		t.Errorf("Parse: %v", err)
	}
	count := 0
	v.Iter(func(p gruid.Point, r rune) {
		if r == '#' {
			count++
		}
	})
	if count != 6 {
		t.Errorf("bad # count: %d (expected 6)", count)
	}
}

func TestVaultSetRunes(t *testing.T) {
	v := &Vault{}
	v.SetRunes("@")
	if v.Runes() != "@" {
		t.Errorf("bad runes: %v", v.Runes())
	}
	err := v.Parse(vaultExample)
	if err == nil {
		t.Error("incomplete rune check")
	}
	v.SetRunes(".#")
	err = v.Parse(vaultExample)
	if err != nil {
		t.Error("bad rune check")
	}
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

// Those constants represent the different types of terrains in the map grid.
const (
	wall Cell = iota
	ground
)

// playerPath implements paths.Astar interface.
type playerPath struct {
	neighbors *paths.Neighbors
	mapgd     Grid
}

func (pp *playerPath) Neighbors(p gruid.Point) []gruid.Point {
	if pp.mapgd.At(p) == wall {
		return nil
	}
	return pp.neighbors.Cardinal(p, func(q gruid.Point) bool {
		if !pp.mapgd.Contains(q) {
			return false
		}
		c := pp.mapgd.At(q)
		return c != wall
	})
}

func TestRandomWalkCave(t *testing.T) {
	mapgd := NewGrid(80, 24)
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	mgen := MapGen{Rand: rd, Grid: mapgd}
	wlk := walker{rand: rd}
	wlk.neighbors = &paths.Neighbors{}
	n := mgen.RandomWalkCave(wlk, ground, 0.5, 1)
	if n != mgen.Grid.Count(ground) {
		t.Errorf("bad count %d vs %d", n, mgen.Grid.Count(ground))
	}
	if n < 80*24/2 || n > 80*24/2+1 {
		t.Errorf("bad fill count: %d", n)
	}
}

func TestCellularAutomataCave(t *testing.T) {
	mapgd := NewGrid(80, 24)
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	mgen := MapGen{Rand: rd, Grid: mapgd}
	rules := []CellularAutomataRule{
		{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
		{WCutoff1: 5, WCutoff2: 25, Reps: 3, WallsOutOfRange: true},
	}
	n := mgen.CellularAutomataCave(wall, ground, 0.40, rules)
	if n != mgen.Grid.Count(ground) {
		t.Errorf("bad count %d vs %d", n, mgen.Grid.Count(ground))
	}
	pr := paths.NewPathRange(mapgd.Range())
	pp := &playerPath{}
	pp.neighbors = &paths.Neighbors{}
	pp.mapgd = mgen.Grid
	pr.CCMapAll(pp)
	it := mgen.Grid.Iterator()
	for it.Next() {
		if it.Cell() == ground {
			m := mgen.KeepCC(pr, it.P(), wall)
			if m > n {
				t.Errorf("bad number of ground cells: %d vs %d (KeepCC from %v)", m, n, it.P())
			}
			break
		}
	}
}

func BenchmarkMapGenRandomWalkCave(b *testing.B) {
	mapgd := NewGrid(80, 24)
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	mgen := MapGen{Rand: rd, Grid: mapgd}
	wlk := walker{rand: rd}
	wlk.neighbors = &paths.Neighbors{}
	for i := 0; i < b.N; i++ {
		mgen.Grid.Fill(Cell(0))
		mgen.RandomWalkCave(wlk, ground, 0.5, 1)
	}
}

func BenchmarkMapGenCellularAutomataCave(b *testing.B) {
	mapgd := NewGrid(80, 24)
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	mgen := MapGen{Rand: rd, Grid: mapgd}
	rules := []CellularAutomataRule{
		{WCutoff1: 5, WCutoff2: 2, Reps: 4, WallsOutOfRange: true},
		{WCutoff1: 5, WCutoff2: 25, Reps: 3, WallsOutOfRange: true},
	}
	for i := 0; i < b.N; i++ {
		mgen.CellularAutomataCave(wall, ground, 0.40, rules)
	}
}
