// +build sdl js

package main

import (
	"image"
	"image/color"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/tiles"
)

func getTileDrawer() (*TileDrawer, error) {
	t := &TileDrawer{}
	var err error
	// We get a monospace font TTF.
	font, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	// We retrieve a font face.
	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size: 24,
		DPI:  72,
	})
	if err != nil {
		return nil, err
	}
	// We create a new drawer for tiles using the previous face. Note that
	// if more than one face is wanted (such as an italic or bold variant),
	// you would have to create drawers for thoses faces too, and then use
	// the relevant one accordingly in the GetImage method.
	t.drawer, err = tiles.NewDrawer(face)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Tile implements TileManager.
type TileDrawer struct {
	drawer *tiles.Drawer
}

func (t *TileDrawer) GetImage(c gruid.Cell) *image.RGBA {
	// we use some selenized colors
	fg := image.NewUniform(color.RGBA{0xad, 0xbc, 0xbc, 255})
	switch c.Style.Fg {
	case ColorLnum:
		fg = image.NewUniform(color.RGBA{0x46, 0x95, 0xf7, 255})
	case ColorTitle:
		fg = image.NewUniform(color.RGBA{0xdb, 0xb3, 0x2d, 255})
	}
	bg := image.NewUniform(color.RGBA{0x10, 0x3c, 0x48, 255})
	return t.drawer.Draw(c.Rune, fg, bg)
}

func (t *TileDrawer) TileSize() gruid.Point {
	return t.drawer.Size()
}
