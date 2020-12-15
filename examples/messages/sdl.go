// +build sdl

package main

import (
	"log"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/sdl"
)

var driver gruid.Driver

func init() {
	t, err := getTileDrawer()
	if err != nil {
		log.Fatal(err)
	}
	driver = &sdl.Driver{
		TileManager: t,
		Width:       80,
		Height:      24,
	}
}
