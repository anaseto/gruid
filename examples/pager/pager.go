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
		fmt.Println("Usage: go run . file-to-read")
		os.Exit(1)
	}

	var gd = gruid.NewGrid(gruid.GridConfig{})
	st := styler{}
	var dri = &tcell.Driver{StyleManager: st}
	m := &model{
		grid:  gd,
		lines: lines,
		fname: args[0],
	}
	app := gruid.NewApp(gruid.AppConfig{
		Driver: dri,
		Model:  m,
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

type model struct {
	grid  gruid.Grid
	lines []string
	pager *ui.Pager
	init  bool
	fname string
}

func (m *model) Init() gruid.Effect {
	st := gruid.Style{}
	style := ui.PagerStyle{
		Title:   st.WithFg(ColorTitle),
		LineNum: st.WithFg(ColorLnum),
	}
	pager := ui.NewPager(ui.PagerConfig{
		Grid:       m.grid,
		StyledText: ui.StyledText{},
		Lines:      m.lines,
		Title:      m.fname,
		Style:      style,
	})
	m.init = true
	m.pager = pager
	return nil
}

func (m *model) Update(msg gruid.Msg) gruid.Effect {
	m.init = false
	m.pager.Update(msg)
	switch m.pager.Action() {
	case ui.PagerQuit:
		return gruid.Quit()
	}
	return nil
}

func (m *model) Draw() gruid.Grid {
	if m.pager.Action() != ui.PagerPass || m.init {
		m.grid.Fill(gruid.Cell{Rune: ' '})
		m.pager.Draw()
	}
	return m.grid
}
