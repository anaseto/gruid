// This file implements a line of sight algorithm.

package rl

import (
	"bytes"
	"encoding/gob"

	"github.com/anaseto/gruid"
)

// FOV represents a field of vision. With a well-defined Lighter, it has the
// following properties: symmetric light rays, expansive walls, permissive with
// blind diagonal corners, simple octant-based geometry, fast computation.
//
// The default algorithm works in a way that can remind of the Dijkstra
// algorithm, but within each cone between a diagonal and an orthogonal axis
// (an octant), only movements along those two directions are allowed. This
// allows the algorithm to be a simple pass on squares around the player,
// starting from radius 1 until line of sight range.
//
// Going from a gruid.Point p to a gruid.Point q has a cost, which depends
// essentially on the type of terrain in p, and is determined by a Lighter.
//
// The obtained light rays are lines formed using at most two adjacent
// directions: a diagonal and an orthogonal one (for example north east and
// east).
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
	LMap     []fovNode
	Lighted  []LightNode
	RayCache []LightNode
	Idx      int         // light map number (for caching)
	Rg       gruid.Range // range of valid positions
	Src      gruid.Point
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

// At returns the total ray cost at a given position from the last source given
// to VisionMap. It returns a false boolean if the position was out of reach.
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

func (fov *FOV) octantParents(ps []gruid.Point, src, p gruid.Point) []gruid.Point {
	q := src.Sub(p)
	r := gruid.Point{sign(q.X), sign(q.Y)}
	ps = append(ps, p.Add(gruid.Point{r.X, r.Y}))
	switch {
	case q.X == 0 || q.Y == 0 || abs(q.X) == abs(q.Y):
	case abs(q.X) > abs(q.Y):
		ps = append(ps, p.Add(gruid.Point{r.X, 0}))
	default:
		ps = append(ps, p.Add(gruid.Point{0, r.Y}))
	}
	return ps
}

func (fov *FOV) bestParent(lt Lighter, src, p gruid.Point) (gruid.Point, int) {
	var psa [2]gruid.Point
	ps := psa[:0]
	ps = fov.octantParents(ps, src, p)
	q := ps[0]
	cost0 := fov.LMap[fov.idx(q)].Cost + lt.Cost(src, q, p)
	if len(ps) < 2 {
		return q, cost0
	}
	cost1 := fov.LMap[fov.idx(ps[1])].Cost + lt.Cost(src, ps[1], p)
	if cost1 < cost0 {
		q = ps[1]
		return q, cost1
	}
	return q, cost0
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
	// anymore. It should normally be equal to maximum sight distance.
	MaxCost(src gruid.Point) int
}

// VisionMap builds a field of vision map for a viewer at src. It returns a
// cached slice of lighted nodes. Values can also be consulted individually
// with At.
func (fov *FOV) VisionMap(lt Lighter, src gruid.Point) []LightNode {
	fov.Idx++
	if !src.In(fov.Rg) {
		return fov.Lighted
	}
	fov.Src = src
	fov.LMap[fov.idx(src)] = fovNode{Cost: 0, Idx: fov.Idx}
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
	fov.computeLighted()
	return fov.Lighted
}

func (fov *FOV) visionUpdate(lt Lighter, src gruid.Point, to gruid.Point) {
	_, c := fov.bestParent(lt, src, to)
	fov.LMap[fov.idx(to)] = fovNode{Cost: c, Idx: fov.Idx}
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
	_, c := fov.bestParent(lt, src, to)
	n := &fov.LMap[fov.idx(to)]
	if n.Idx == fov.Idx && n.Cost <= c {
		return
	}
	*n = fovNode{Cost: c, Idx: fov.Idx}
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
	var c int
	for to != fov.Src {
		oto := to
		to, c = fov.bestParent(lt, fov.Src, oto)
		fov.RayCache = append(fov.RayCache, LightNode{P: oto, Cost: c})
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
