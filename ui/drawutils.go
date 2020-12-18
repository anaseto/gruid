package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

type box struct {
	grid  gruid.Grid
	title StyledText  // for the title
	style gruid.Style // for the borders
}

func (b box) draw() {
	rg := b.grid.Range().Origin()
	if rg.Empty() {
		return
	}
	cgrid := b.grid.Slice(rg.Shift(1, 0, -1, 0))
	crg := cgrid.Range().Origin()
	cell := gruid.Cell{Style: b.style}
	cell.Rune = '─'
	w, h := crg.Size()
	if b.title.Text() != "" {
		nchars := utf8.RuneCountInString(b.title.Text())
		dist := (w - nchars) / 2
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
		b.title.Draw(cgrid.Slice(crg.Line(0).Shift(dist, 0, 0, 0)))
		line = cgrid.Slice(crg.Line(0).Shift(dist+nchars, 0, 0, 0))
		line.Fill(cell)
	} else {
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
	}
	line := cgrid.Slice(crg.Line(h - 1))
	line.Fill(cell)
	w, h = rg.Size()
	b.grid.SetCell(rg.Min, cell.WithRune('┌'))
	b.grid.SetCell(gruid.Position{X: w - 1}, cell.WithRune('┐'))
	b.grid.SetCell(gruid.Position{Y: h - 1}, cell.WithRune('└'))
	b.grid.SetCell(rg.Max.Shift(-1, -1), cell.WithRune('┘'))
	cell.Rune = '│'
	col := b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(0))
	col.Fill(cell)
	col = b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(w - 1))
	col.Fill(cell)
}
