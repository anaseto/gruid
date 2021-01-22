package ui

import (
	"github.com/anaseto/gruid"
)

// Alignment represents left, center or right text alignment. It is used to
// draw text with a given alignment within a grid.
type Alignment int16

// Those constants represent the possible text alignment options. The default
// alignment is AlignCenter.
const (
	AlignCenter Alignment = iota
	AlignLeft
	AlignRight
)

// Box contains information to draw a rectangle using box characters, with an
// optional title.
type Box struct {
	Style       gruid.Style // box style
	Title       StyledText  // optional top text
	Footer      StyledText  // optional bottom text
	AlignTitle  Alignment   // title alignment
	AlignFooter Alignment   // footer alignment
}

// Draw draws a rectangular box in a grid, taking the whole grid. It does not
// draw anything in the interior region. It returns the grid slice that was
// drawn, which usually is the whole grid, except if the grid was too small to
// draw a box.
func (b Box) Draw(gd gruid.Grid) gruid.Grid {
	rg := gd.Range()
	max := rg.Size()
	if max.X < 2 || max.Y < 2 {
		return gd.Slice(gruid.Range{})
	}
	cgrid := gd.Slice(rg.Shift(1, 0, -1, 0))
	crg := cgrid.Range()
	cell := gruid.Cell{Style: b.Style}
	cell.Rune = '─'
	max = crg.Size()
	line := cgrid.Slice(crg.Line(0))
	line.Fill(cell)
	if b.Title.Text() != "" {
		b.Title.drawTextLine(line, b.AlignTitle)
	}
	line = cgrid.Slice(crg.Line(max.Y - 1))
	line.Fill(cell)
	if b.Footer.Text() != "" {
		b.Footer.drawTextLine(line, b.AlignFooter)
	}
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
	return gd
}

func (stt StyledText) drawTextLine(gd gruid.Grid, align Alignment) {
	switch align {
	case AlignCenter:
		tw := stt.Size().X
		rg := gd.Range()
		w := rg.Max.X
		dist := (w - tw) / 2
		stt.Draw(gd.Slice(rg.Shift(dist, 0, 0, 0)))
	case AlignLeft:
		stt.Draw(gd)
	case AlignRight:
		tw := stt.Size().X
		rg := gd.Range()
		w := rg.Max.X
		dist := w - tw
		stt.Draw(gd.Slice(rg.Shift(dist, 0, 0, 0)))
	}
}
