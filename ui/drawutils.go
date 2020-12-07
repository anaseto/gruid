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
	rg := b.grid.Range().Relative()
	if rg.Empty() {
		return
	}
	cgrid := b.grid.Slice(rg.Shift(1, 0, -1, 0))
	crg := cgrid.Range().Relative()
	cell := gruid.Cell{Style: b.style}
	cell.Rune = '─'
	if b.title.Text() != "" {
		nchars := utf8.RuneCountInString(b.title.Text())
		dist := (crg.Width() - nchars) / 2
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
		b.title.Draw(cgrid.Slice(crg.Line(0).Shift(dist, 0, 0, 0)))
		line = cgrid.Slice(crg.Line(0).Shift(dist+nchars, 0, 0, 0))
		line.Fill(cell)
	} else {
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
	}
	line := cgrid.Slice(crg.Line(crg.Height() - 1))
	line.Fill(cell)
	b.grid.SetCell(rg.Min, cell.WithRune('┌'))
	b.grid.SetCell(gruid.Position{X: rg.Width() - 1}, cell.WithRune('┐'))
	b.grid.SetCell(gruid.Position{Y: rg.Height() - 1}, cell.WithRune('└'))
	b.grid.SetCell(rg.Max.Shift(-1, -1), cell.WithRune('┘'))
	cell.Rune = '│'
	col := b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(0))
	col.Fill(cell)
	col = b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(rg.Width() - 1))
	col.Fill(cell)
}
