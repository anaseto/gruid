package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// Box contains information to draw a rectangle using box characters, with an
// optional title.
type Box struct {
	Style gruid.Style // box style
	Title StyledText  // optional title
}

// Draw draws a rectangular box in a grid, taking the whole grid. It does not
// draw anything in the interior region.
func (b Box) Draw(gd gruid.Grid) {
	rg := gd.Range().Origin()
	if rg.Empty() {
		return
	}
	cgrid := gd.Slice(rg.Shift(1, 0, -1, 0))
	crg := cgrid.Range().Origin()
	cell := gruid.Cell{Style: b.Style}
	cell.Rune = '─'
	max := crg.Size()
	if b.Title.Text() != "" {
		nchars := utf8.RuneCountInString(b.Title.Text())
		dist := (max.X - nchars) / 2
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
		b.Title.Draw(cgrid.Slice(crg.Line(0).Shift(dist, 0, 0, 0)))
		line = cgrid.Slice(crg.Line(0).Shift(dist+nchars, 0, 0, 0))
		line.Fill(cell)
	} else {
		line := cgrid.Slice(crg.Line(0))
		line.Fill(cell)
	}
	line := cgrid.Slice(crg.Line(max.Y - 1))
	line.Fill(cell)
	max = rg.Size()
	gd.Set(rg.Min, cell.WithRune('┌'))
	gd.Set(gruid.Point{X: max.X - 1}, cell.WithRune('┐'))
	gd.Set(gruid.Point{Y: max.Y - 1}, cell.WithRune('└'))
	gd.Set(rg.Max.Shift(-1, -1), cell.WithRune('┘'))
	cell.Rune = '│'
	col := gd.Slice(rg.Shift(0, 1, 0, -1).Column(0))
	col.Fill(cell)
	col = gd.Slice(rg.Shift(0, 1, 0, -1).Column(max.X - 1))
	col.Fill(cell)
}
