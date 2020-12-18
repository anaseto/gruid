package main

import (
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

	var gd = gruid.NewGrid(80, 24)
	st := styler{}
	var dri = tcell.NewDriver(tcell.Config{StyleManager: st})
	pager := NewPager(gd, lines, args[0])
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  pager,
	})
	app.Start(nil)
	fmt.Printf("Successful quit.\n")
}

const (
	ColorTitle gruid.Color = 1 + iota // skip zero value ColorDefault
	ColorLnum
)

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

func NewPager(grid gruid.Grid, lines []string, fname string) *ui.Pager {
	st := gruid.Style{}
	style := ui.PagerStyle{
		Title:   st.WithFg(ColorTitle),
		LineNum: st.WithFg(ColorLnum),
	}
	pager := ui.NewPager(ui.PagerConfig{
		Grid:       grid,
		StyledText: ui.StyledText{},
		Lines:      lines,
		Title:      fname,
		Style:      style,
	})
	return pager
}
