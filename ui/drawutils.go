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
	max := crg.Size()
	if b.title.Text() != "" {
		nchars := utf8.RuneCountInString(b.title.Text())
		dist := (max.X - nchars) / 2
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
		b.title.Draw(cgrid.Slice(crg.Line(0).Shift(dist, 0, 0, 0)))
		line = cgrid.Slice(crg.Line(0).Shift(dist+nchars, 0, 0, 0))
		line.Fill(cell)
	} else {
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
	}
	line := cgrid.Slice(crg.Line(max.Y - 1))
	line.Fill(cell)
	max = rg.Size()
	b.grid.Set(rg.Min, cell.WithRune('┌'))
	b.grid.Set(gruid.Point{X: max.X - 1}, cell.WithRune('┐'))
	b.grid.Set(gruid.Point{Y: max.Y - 1}, cell.WithRune('└'))
	b.grid.Set(rg.Max.Shift(-1, -1), cell.WithRune('┘'))
	cell.Rune = '│'
	col := b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(0))
	col.Fill(cell)
	col = b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(max.X - 1))
	col.Fill(cell)
}
