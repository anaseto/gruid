// +build tk

package main

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tk"
	"github.com/anaseto/gruid/tiles"
)

var driver gruid.Driver

func init() {
	tile := &Tile{}
	var err error
	font, err := opentype.Parse(gomono.TTF)
	if err != nil {
		log.Fatal(err)
	}
	face, err := opentype.NewFace(font, &opentype.FaceOptions{
		Size: 24,
		DPI:  72,
	})
	if err != nil {
		log.Fatal(err)
	}
	tile.drawer, err = tiles.NewDrawer(face)
	if err != nil {
		log.Fatal(err)
	}
	driver = &tk.Driver{
		TileManager: tile,
		Width:       80,
		Height:      24,
	}
}

type Tile struct {
	drawer *tiles.Drawer
}

func (t *Tile) GetImage(c gruid.Cell) *image.RGBA {
	// we use some selenized colors
	fg := image.NewUniform(color.RGBA{0xad, 0xbc, 0xbc, 255})
	switch c.Style.Fg {
	case ColorHeader:
		fg = image.NewUniform(color.RGBA{0xd1, 0xa4, 0x16, 255})
	}
	bg := image.NewUniform(color.RGBA{0x10, 0x3c, 0x48, 255})
	return t.drawer.Draw(c.Rune, fg, bg)
}

func (t *Tile) TileSize() (int, int) {
	return t.drawer.Size()
}
