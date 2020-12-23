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
	dri := sdl.NewDriver(sdl.Config{
		TileManager: t,
	})
	//dri.SetScale(2.0, 2.0)
	dri.PreventQuit()
	driver = dri
}
