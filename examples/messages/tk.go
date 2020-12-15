// +build tk

package main

import (
	"log"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tk"
)

var driver gruid.Driver

func init() {
	t, err := getTileDrawer()
	if err != nil {
		log.Fatal(err)
	}
	driver = &tk.Driver{
		TileManager: t,
		Width:       80,
		Height:      24,
	}
}
