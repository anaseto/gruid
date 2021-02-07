// This file implements line of sight algorithms.

package rl

import (
	"bytes"
	"encoding/gob"

	"github.com/anaseto/gruid"
)

// FOV represents a field of vision. There are two main algorithms available:
// VisionMap and SCCVisionMap, and their multiple-source versions. Both
// algorithms are symmetric (under certain conditions) with expansive walls,
// and fast.
//
// The VisionMap method algorithm is more permissive and only produces
// continuous light rays. Moreover, it allows for non-binary visibility (some
// obstacles may reduce sight range without blocking it completely).
//
// The SCCVisionMap method algorithm is a symmetric shadow casting algorithm
// based on:
//
// 	https://www.albertford.com/shadowcasting/
//
// It offers more euclidian-like geometry, less permissive and with expansive
// shadows, while still being symmetric and fast enough.
//
// Both algorithms can be combined to obtain a less permissive version of
// VisionMap with expansive shadows, while keeping the non-binary visibility
// information.
//
// FOV implements the gob.Decoder and gob.Encoder interfaces for easy
// serialization.
type FOV struct {
	innerFOV
}

type fovNode struct {
	Idx  int // map number (for caching)
	Cost int // ray cost from source to this node
}

type innerFOV struct {
	LMap          []fovNode
	Lighted       []LightNode
	RayCache      []LightNode
	ShadowCasting []bool      // visibility
	Idx           int         // light map number (for caching)
	Rg            gruid.Range // range of valid positions
	Src           gruid.Point
	passable      func(gruid.Point) bool
}

// NewFOV returns new ready to use field of view with a given range of valid
// positions.
func NewFOV(rg gruid.Range) *FOV {
	fov := &FOV{}
	fov.Rg = rg
	fov.LMap = make([]fovNode, fov.Rg.Size().X*fov.Rg.Size().Y)
	return fov
}

// SetRange updates the range used by the field of view. If the size is the
// same, cached structures will be preserved, otherwise they will be
// reinitialized.
func (fov *FOV) SetRange(rg gruid.Range) {
	org := fov.Rg
	fov.Rg = rg
	max := rg.Size()
	omax := org.Size()
	if max == omax {
		return
	}
	*fov = FOV{}
	fov.Rg = rg
}

// Range returns the current FOV's range of positions.
func (fov *FOV) Range() gruid.Range {
	return fov.Rg
}

// GobDecode implements gob.GobDecoder.
func (fov *FOV) GobDecode(bs []byte) error {
	r := bytes.NewReader(bs)
	gd := gob.NewDecoder(r)
	ifov := &innerFOV{}
	err := gd.Decode(ifov)
	if err != nil {
		return err
	}
	fov.innerFOV = *ifov
	return nil
}

// GobEncode implements gob.GobEncoder.
func (fov *FOV) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	ge := gob.NewEncoder(&buf)
	err := ge.Encode(&fov.innerFOV)
	return buf.Bytes(), err
}

// At returns the total ray cost at a given position from the last source(s)
// given to VisionMap or LightMap. It returns a false boolean if the position
// was out of reach.
func (fov *FOV) At(p gruid.Point) (int, bool) {
	if !p.In(fov.Rg) || fov.LMap == nil {
		return 0, false
	}
	node := fov.LMap[fov.idx(p)]
	if node.Idx != fov.Idx {
		return node.Cost, false
	}
	return node.Cost, true
}

// Visible returns true if the given position is visible according to the
// previous SCCVisionMap call.
func (fov *FOV) Visible(p gruid.Point) bool {
	if !p.In(fov.Rg) || fov.ShadowCasting == nil {
		return false
	}
	return fov.ShadowCasting[fov.idx(p)]
}

func (fov *FOV) idx(p gruid.Point) int {
	p = p.Sub(fov.Rg.Min)
	w := fov.Rg.Max.X - fov.Rg.Min.X
	return p.Y*w + p.X
}

// Iter iterates a function on the nodes lighted in the last VisionMap or
// LightMap.
func (fov *FOV) Iter(fn func(LightNode)) {
	for _, n := range fov.Lighted {
		fn(n)
	}
}

func sign(n int) int {
	var i int
	switch {
	case n > 0:
		i = 1
	case n < 0:
		i = -1
	}
	return i
}

func (fov *FOV) octantParents(ps []LightNode, src, p gruid.Point) []LightNode {
	q := src.Sub(p)
	r := gruid.Point{sign(q.X), sign(q.Y)}
	p0 := p.Add(gruid.Point{r.X, r.Y})
	n0 := fov.LMap[fov.idx(p0)]
	if n0.Idx == fov.Idx {
		ps = append(ps, LightNode{P: p0, Cost: n0.Cost})
	}
	switch {
	case q.X == 0 || q.Y == 0 || abs(q.X) == abs(q.Y):
	case abs(q.X) > abs(q.Y):
		p1 := p.Add(gruid.Point{r.X, 0})
		n1 := fov.LMap[fov.idx(p1)]
		if n1.Idx == fov.Idx {
			ps = append(ps, LightNode{P: p1, Cost: n1.Cost})
		}
	default:
		p1 := p.Add(gruid.Point{0, r.Y})
		n1 := fov.LMap[fov.idx(p1)]
		if n1.Idx == fov.Idx {
			ps = append(ps, LightNode{P: p1, Cost: n1.Cost})
		}
	}
	return ps
}

// From returns the previous position in the light ray to the given position,
// as computed in the last VisionMap call.  If none, it returns false.
func (fov *FOV) From(lt Lighter, to gruid.Point) (LightNode, bool) {
	_, ok := fov.At(to)
	if !ok {
		return LightNode{}, false
	}
	ln := fov.from(lt, to)
	if ln.Cost == -1 {
		return LightNode{}, false
	}
	return ln, true
}

func (fov *FOV) from(lt Lighter, to gruid.Point) LightNode {
	var pnodesa [2]LightNode
	pnodes := pnodesa[:0]
	pnodes = fov.octantParents(pnodes, fov.Src, to)
	switch len(pnodes) {
	case 0:
		return LightNode{Cost: -1}
	case 1:
		n := pnodes[0]
		return LightNode{P: n.P, Cost: n.Cost + lt.Cost(fov.Src, n.P, to)}
	default:
		n := pnodes[0]
		m := pnodes[1]
		cost0 := n.Cost + lt.Cost(fov.Src, n.P, to)
		cost1 := m.Cost + lt.Cost(fov.Src, m.P, to)
		if cost0 <= cost1 {
			return LightNode{P: n.P, Cost: cost0}
		}
		return LightNode{P: m.P, Cost: cost1}
	}
}

// Lighter is the interface that captures the requirements for light ray
// propagation.
type Lighter interface {
	// Cost returns the cost of light propagation from a position to
	// an adjacent one given an original source. If you want the resulting
	// FOV to be symmetric, the function should generate symmetric costs
	// for rays in both directions.
	//
	// Note that the FOV algorithm takes care of only providing (from, to)
	// couples that may belong to a same light ray whose source is src,
	// independently of terrain.  This means that the Cost function should
	// essentially take care of terrain considerations, for example giving
	// a cost of 1 if from is a regular ground cell, and a maximal cost if
	// it is a wall, or something in between for fog, bushes or other
	// terrains.
	//
	// As a special case, you normally want Cost(src, src, to) == 1
	// independently of the terrain in src to guarantee symmetry, except
	// for diagonals in certain cases with 4-way movement, because two
	// walls could block vision (for example).
	Cost(src gruid.Point, from gruid.Point, to gruid.Point) int

	// MaxCost indicates the cost limit at which light cannot propagate
	// anymore from the given source. It should normally be equal to
	// maximum sight or light distance.
	MaxCost(src gruid.Point) int
}

// VisionMap builds a field of vision map for a viewer at src. It returns a
// cached slice of lighted nodes. Values can also be consulted individually
// with At.
//
// The algorithm works in a way that can remind of the Dijkstra algorithm, but
// within each cone between a diagonal and an orthogonal axis (an octant), only
// movements along those two directions are allowed. This allows the algorithm
// to be a simple pass on squares around the player, starting from radius 1
// until line of sight range.
//
// Going from a gruid.Point p to a gruid.Point q has a cost, which depends
// essentially on the type of terrain in p, and is determined by a Lighter.
//
// The obtained light rays are lines formed using at most two adjacent
// directions: a diagonal and an orthogonal one (for example north east and
// east).
func (fov *FOV) VisionMap(lt Lighter, src gruid.Point) []LightNode {
	fov.Idx++
	fov.Lighted = fov.Lighted[:0]
	if !src.In(fov.Rg) {
		return fov.Lighted
	}
	fov.Src = src
	fov.LMap[fov.idx(src)] = fovNode{Cost: 0, Idx: fov.Idx}
	fov.Lighted = append(fov.Lighted, LightNode{P: src, Cost: 0})
	for d := 1; d <= lt.MaxCost(src); d++ {
		rg := fov.Rg.Intersect(gruid.NewRange(src.X-d, src.Y-d+1, src.X+d+1, src.Y+d))
		if src.Y+d < fov.Rg.Max.Y {
			for x := rg.Min.X; x < rg.Max.X; x++ {
				fov.visionUpdate(lt, src, gruid.Point{x, src.Y + d})
			}
		}
		if src.Y-d >= fov.Rg.Min.Y {
			for x := rg.Min.X; x < rg.Max.X; x++ {
				fov.visionUpdate(lt, src, gruid.Point{x, src.Y - d})
			}
		}
		if src.X+d < fov.Rg.Max.X {
			for y := rg.Min.Y; y < rg.Max.Y; y++ {
				fov.visionUpdate(lt, src, gruid.Point{src.X + d, y})
			}
		}
		if src.X-d >= fov.Rg.Min.X {
			for y := rg.Min.Y; y < rg.Max.Y; y++ {
				fov.visionUpdate(lt, src, gruid.Point{src.X - d, y})
			}
		}
	}
	fov.checkIdx()
	return fov.Lighted
}

func (fov *FOV) visionUpdate(lt Lighter, src gruid.Point, to gruid.Point) {
	n := fov.from(lt, to)
	if n.Cost >= 0 {
		fov.LMap[fov.idx(to)] = fovNode{Cost: n.Cost, Idx: fov.Idx}
		fov.Lighted = append(fov.Lighted, LightNode{P: to, Cost: n.Cost})
	}
}

// LightMap builds a lighting map with given light sources. It returs a cached
// slice of lighted nodes. Values can also be consulted with At.
func (fov *FOV) LightMap(lt Lighter, srcs []gruid.Point) []LightNode {
	fov.Idx++
	for _, src := range srcs {
		if !src.In(fov.Rg) {
			continue
		}
		fov.Src = src
		fov.LMap[fov.idx(src)] = fovNode{Cost: 0, Idx: fov.Idx}
		for d := 1; d <= lt.MaxCost(src); d++ {
			rg := fov.Rg.Intersect(gruid.NewRange(src.X-d, src.Y-d+1, src.X+d+1, src.Y+d))
			if src.Y+d < fov.Rg.Max.Y {
				for x := rg.Min.X; x < rg.Max.X; x++ {
					fov.lightUpdate(lt, src, gruid.Point{x, src.Y + d})
				}
			}
			if src.Y-d >= fov.Rg.Min.Y {
				for x := rg.Min.X; x < rg.Max.X; x++ {
					fov.lightUpdate(lt, src, gruid.Point{x, src.Y - d})
				}
			}
			if src.X+d < fov.Rg.Max.X {
				for y := rg.Min.Y; y < rg.Max.Y; y++ {
					fov.lightUpdate(lt, src, gruid.Point{src.X + d, y})
				}
			}
			if src.X-d >= fov.Rg.Min.X {
				for y := rg.Min.Y; y < rg.Max.Y; y++ {
					fov.lightUpdate(lt, src, gruid.Point{src.X - d, y})
				}
			}
		}
	}
	fov.checkIdx()
	fov.computeLighted()
	return fov.Lighted
}

func (fov *FOV) lightUpdate(lt Lighter, src gruid.Point, to gruid.Point) {
	n := fov.from(lt, to)
	if n.Cost < 0 {
		return
	}
	m := &fov.LMap[fov.idx(to)]
	if m.Idx == fov.Idx && m.Cost <= n.Cost {
		return
	}
	*m = fovNode{Cost: n.Cost, Idx: fov.Idx}
}

func (fov *FOV) computeLighted() {
	fov.Lighted = fov.Lighted[:0]
	w := fov.Rg.Max.X - fov.Rg.Min.X
	h := len(fov.LMap) / w
	i := 0
	for y := 0; y < h; y = y + 1 {
		for x := 0; x < w; x, i = x+1, i+1 {
			n := fov.LMap[i]
			if n.Idx == fov.Idx {
				fov.Lighted = append(fov.Lighted, LightNode{P: gruid.Point{x, y}.Add(fov.Rg.Min), Cost: n.Cost})
			}
		}
	}
}

func (fov *FOV) checkIdx() {
	if fov.Idx+1 > 0 {
		return
	}
	for i, n := range fov.LMap {
		idx := 0
		if n.Idx == fov.Idx {
			idx = 1
		}
		fov.LMap[i] = fovNode{Cost: n.Cost, Idx: idx}
	}
	fov.Idx = 1
}

// LightNode represents the information attached to a given position in a light
// map.
type LightNode struct {
	P    gruid.Point // position in the light ray
	Cost int         // light cost
}

// Ray returns a single light ray from the source (viewer) position to another.
// It should be preceded by a VisionMap call. If the destination position is
// not within the max distance from the source, a nil slice will be returned.
//
// The returned slice is cached for efficiency, so results will be invalidated
// by future calls.
func (fov *FOV) Ray(lt Lighter, to gruid.Point) []LightNode {
	_, okTo := fov.At(to)
	if !okTo {
		return nil
	}
	fov.RayCache = fov.RayCache[:0]
	var n LightNode
	for to != fov.Src {
		n = fov.from(lt, to)
		fov.RayCache = append(fov.RayCache, LightNode{P: to, Cost: n.Cost})
		to = n.P
	}
	fov.RayCache = append(fov.RayCache, LightNode{P: fov.Src, Cost: 0})
	for i := range fov.RayCache[:len(fov.RayCache)/2] {
		fov.RayCache[i], fov.RayCache[len(fov.RayCache)-i-1] = fov.RayCache[len(fov.RayCache)-i-1], fov.RayCache[i]
	}
	return fov.RayCache
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type row struct {
	depth      int
	slopeStart gruid.Point // fractional number
	slopeEnd   gruid.Point
}

func (r row) tiles(ts []gruid.Point, colmin, colmax int) []gruid.Point {
	min := r.depth * r.slopeStart.X
	div, rem := min/r.slopeStart.Y, min%r.slopeStart.Y
	min = div
	switch sign(rem) {
	case 1:
		if 2*rem >= r.slopeStart.Y {
			min = div + 1
		}
	case -1:
		if -2*rem > r.slopeStart.Y {
			min = div - 1
		}
	}
	max := r.depth * r.slopeEnd.X
	div, rem = max/r.slopeEnd.Y, max%r.slopeEnd.Y
	max = div
	switch sign(rem) {
	case 1:
		if 2*rem > r.slopeEnd.Y {
			max = div + 1
		}
	case -1:
		if -rem*2 >= r.slopeEnd.Y {
			max = div - 1
		}
	}
	if min < colmin {
		min = colmin
	}
	if max > colmax {
		max = colmax
	}
	for col := min; col < max+1; col++ {
		ts = append(ts, gruid.Point{r.depth, col})
	}
	return ts
}

func (r row) next() row {
	r.depth++
	return r
}

func (r row) isSymmetric(tile gruid.Point) bool {
	_, col := tile.X, tile.Y
	return col*r.slopeStart.Y >= r.depth*r.slopeStart.X &&
		col*r.slopeEnd.Y <= r.depth*r.slopeEnd.X
}

func slope(tile gruid.Point) gruid.Point {
	depth, col := tile.X, tile.Y
	return gruid.Point{2*col - 1, 2 * depth}
}

type quadDir int

const (
	north quadDir = iota
	east
	south
	west
)

type quadrant struct {
	dir quadDir
	p   gruid.Point
}

func (qt quadrant) transform(tile gruid.Point) gruid.Point {
	switch qt.dir {
	case north:
		return gruid.Point{qt.p.X + tile.Y, qt.p.Y - tile.X}
	case south:
		return gruid.Point{qt.p.X + tile.Y, qt.p.Y + tile.X}
	case east:
		return gruid.Point{qt.p.X + tile.X, qt.p.Y + tile.Y}
	default:
		return gruid.Point{qt.p.X - tile.X, qt.p.Y + tile.Y}
	}
}

func (qt quadrant) maxCols(rg gruid.Range) (int, int) {
	switch qt.dir {
	case north, south:
		deltaX := qt.p.X - rg.Min.X
		deltaY := rg.Max.X - qt.p.X - 1
		return -deltaX, deltaY
	default:
		deltaX := qt.p.Y - rg.Min.Y
		deltaY := rg.Max.Y - qt.p.Y - 1
		return -deltaX, deltaY
	}
}

func (qt quadrant) maxDepth(rg gruid.Range) int {
	switch qt.dir {
	case north:
		delta := qt.p.Y - rg.Min.Y
		return delta
	case south:
		delta := rg.Max.Y - qt.p.Y - 1
		return delta
	case east:
		delta := rg.Max.X - qt.p.X - 1
		return delta
	default:
		delta := qt.p.X - rg.Min.X
		return delta
	}
}

func (fov *FOV) reveal(qt quadrant, tile gruid.Point) {
	p := qt.transform(tile)
	fov.ShadowCasting[fov.idx(p)] = true
}

const unreachable = 999999 // an unreachable coordinate (a bit hacky, but simpler)

func (fov *FOV) isWall(qt quadrant, tile gruid.Point) bool {
	if tile.X == unreachable {
		return false
	}
	p := qt.transform(tile)
	return !fov.passable(p)
}

func (fov *FOV) isFloor(qt quadrant, tile gruid.Point) bool {
	if tile.X == unreachable {
		return false
	}
	p := qt.transform(tile)
	return fov.passable(p)
}

// Symmetric shadow casting algorithm described here:
//
// 	https://www.albertford.com/shadowcasting/
//
// Visibility of positions can then be checked with the Visible method.
// Contrary to VisionMap and LightMap, this algorithm can have some
// discontinuous rays.
func (fov *FOV) SSCVisionMap(src gruid.Point, maxDepth int, passable func(p gruid.Point) bool) {
	if !src.In(fov.Rg) {
		return
	}
	if fov.ShadowCasting == nil {
		fov.ShadowCasting = make([]bool, fov.Rg.Size().X*fov.Rg.Size().Y)
	}
	for i := range fov.ShadowCasting {
		fov.ShadowCasting[i] = false
	}
	tiles := []gruid.Point{}
	fov.sscVisionMap(src, tiles[:0], maxDepth, passable)
}

func (fov *FOV) sscVisionMap(src gruid.Point, tiles []gruid.Point, maxDepth int, passable func(p gruid.Point) bool) {
	fov.passable = passable
	fov.ShadowCasting[fov.idx(src)] = true
	for i := 0; i < 4; i++ {
		qt := quadrant{dir: quadDir(i), p: src}
		colmin, colmax := qt.maxCols(fov.Rg)
		dmax := qt.maxDepth(fov.Rg)
		if dmax > maxDepth {
			dmax = maxDepth
		}
		if dmax == 0 {
			continue
		}
		r := row{
			depth:      1,
			slopeStart: gruid.Point{-1, 1},
			slopeEnd:   gruid.Point{1, 1},
		}
		rows := []row{r}
		for len(rows) > 0 {
			r := rows[len(rows)-1]
			rows = rows[:len(rows)-1]
			ptile := gruid.Point{unreachable, unreachable}
			tiles = r.tiles(tiles[:0], colmin, colmax)
			for _, tile := range tiles {
				if fov.isWall(qt, tile) || r.isSymmetric(tile) {
					fov.reveal(qt, tile)
				}
				if fov.isWall(qt, ptile) && fov.isFloor(qt, tile) {
					r.slopeStart = slope(tile)
				}
				if fov.isFloor(qt, ptile) && fov.isWall(qt, tile) {
					nr := r.next()
					nr.slopeEnd = slope(tile)
					if nr.depth <= dmax {
						rows = append(rows, nr)
					}
				}
				ptile = tile
			}
			if fov.isFloor(qt, ptile) {
				if r.depth < dmax {
					rows = append(rows, r.next())
				}
			}
		}
	}
}

// SSCLightMap is the equivalent of SSCVisionMap with several sources.
func (fov *FOV) SSCLightMap(srcs []gruid.Point, maxDepth int, passable func(p gruid.Point) bool) {
	if fov.ShadowCasting == nil {
		fov.ShadowCasting = make([]bool, fov.Rg.Size().X*fov.Rg.Size().Y)
	}
	for i := range fov.ShadowCasting {
		fov.ShadowCasting[i] = false
	}
	tiles := []gruid.Point{}
	for _, src := range srcs {
		if !src.In(fov.Rg) {
			continue
		}
		fov.sscVisionMap(src, tiles[:0], maxDepth, passable)
	}
}
