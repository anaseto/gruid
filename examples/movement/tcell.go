// +build !sdl,!js

package main

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	tc "github.com/gdamore/tcell/v2"
)

var driver gruid.Driver

func init() {
	st := styler{}
	dri := tcell.NewDriver(tcell.Config{StyleManager: st})
	dri.PreventQuit()
	driver = dri
}

type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorPlayer:
		ts = ts.Foreground(tc.ColorNavy) // blue color for the player
	}
	switch st.Bg {
	case ColorPath:
		ts = ts.Reverse(true)
	}
	return ts
}
