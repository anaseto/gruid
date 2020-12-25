// Package tiles provides common utilities for manipulation of graphical tiles,
// such as drawing fonts.
package tiles

import (
	"errors"
	"image"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/anaseto/gruid"
)

// Drawer can be used to draw a text rune on an image of appropriate size to be
// used as a tile. See the example in examples/messages/tiles.go for an example
// of use.
type Drawer struct {
	drawer *font.Drawer
	rect   image.Rectangle
	dot    fixed.Point26_6
}

// NewDrawer returns a new tile Drawer given a monospace font face.
func NewDrawer(face font.Face) (*Drawer, error) {
	d := &Drawer{}
	d.drawer = &font.Drawer{
		Face: face,
	}
	width, ok := face.GlyphAdvance('W')
	if !ok {
		return nil, errors.New("could not get glyph advance")
	}
	metrics := face.Metrics()
	height := metrics.Height
	d.rect = image.Rect(0, 0, width.Ceil(), height.Ceil())
	d.dot = fixed.Point26_6{X: 0, Y: metrics.Ascent}
	return d, nil
}

// Draw draws a rune and returns the produced image with foreground and
// background colors given by images fg and bg.
func (d *Drawer) Draw(r rune, fg, bg image.Image) *image.RGBA {
	d.drawer.Dot = d.dot
	d.drawer.Src = fg
	d.drawer.Dst = image.NewRGBA(d.rect)
	rect := d.drawer.Dst.Bounds()
	draw.Draw(d.drawer.Dst, rect, bg, rect.Min, draw.Src)
	d.drawer.DrawString(string(r))
	img := d.drawer.Dst.(*image.RGBA)
	return img
}

// Size returns the size of drawn tiles, in pixel points.
func (d *Drawer) Size() gruid.Point {
	p := d.rect.Size()
	if p.X <= 0 {
		p.X = 1
	}
	if p.Y <= 0 {
		p.Y = 1
	}
	return gruid.Point{X: p.X, Y: p.Y}
}
