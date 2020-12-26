// This file implements a simple pager using the ui.Pager Model. It reads a
// file given on the command line and split the results into lines to be viewed
// with the pager.

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	"github.com/anaseto/gruid/ui"
	tc "github.com/gdamore/tcell/v2"
)

func main() {
	flag.Parse()
	var lines []string
	args := flag.Args()
	if len(args) > 0 {
		fname := args[0]
		bytes, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Fatal(err)
		}
		s := strings.TrimRight(strings.ReplaceAll(string(bytes), "\t", "    "), "\n")
		lines = strings.Split(s, "\n")
	} else {
		fmt.Println("Usage: go run ./pager.go file")
		os.Exit(1)
	}

	gd := gruid.NewGrid(80, 24)
	st := styler{}
	dr := tcell.NewDriver(tcell.Config{StyleManager: st})
	pager := newPager(gd, lines, args[0])
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dr,
		Model:  pager,
	})
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Successful quit.\n")
}

// Those constants represent the generic colors we use in this example.
const (
	ColorTitle gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorLnum
)

// styler implements the tcell.StyleManager interface.
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorTitle:
		ts = ts.Foreground(tc.ColorOlive)
	case ColorLnum:
		ts = ts.Foreground(tc.ColorNavy)
	}
	return ts
}

// newPager returns an ui.Pager with the given lines as content and filename as
// title. The result can be used as the main model of the application.
func newPager(grid gruid.Grid, lines []string, fname string) *ui.Pager {
	st := gruid.Style{}
	style := ui.PagerStyle{
		LineNum: st.WithFg(ColorLnum),
	}
	pager := ui.NewPager(ui.PagerConfig{
		Grid:       grid,
		StyledText: ui.StyledText{},
		Lines:      lines,
		Box:        &ui.Box{Title: ui.NewStyledText(fname).WithStyle(st.WithFg(ColorTitle))},
		Style:      style,
	})
	return pager
}
