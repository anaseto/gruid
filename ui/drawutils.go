package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

type box struct {
	grid       gruid.Grid
	title      string
	style      gruid.CellStyle
	titleStyle gruid.CellStyle
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
	if b.title != "" {
		nchars := utf8.RuneCountInString(b.title)
		dist := (crg.Width() - nchars) / 2
		line := cgrid.Slice(crg.Line(0))
		line.Iter(func(pos gruid.Position) {
			line.SetCell(pos, cell)
		})
		stt := NewStyledText(b.title)
		stt.SetStyle(b.titleStyle)
		stt.Draw(cgrid.Slice(crg.Line(0).Shift(dist, 0, 0, 0)))
		line = cgrid.Slice(crg.Line(0).Shift(dist+nchars, 0, 0, 0))
		line.Iter(func(pos gruid.Position) {
			line.SetCell(pos, cell)
		})
	} else {
		line := cgrid.Slice(crg.Line(0))
		line.Iter(func(pos gruid.Position) {
			line.SetCell(pos, cell)
		})
	}
	line := cgrid.Slice(crg.Line(crg.Height() - 1))
	line.Iter(func(pos gruid.Position) {
		line.SetCell(pos, cell)
	})
	cell.Rune = '┌'
	b.grid.SetCell(rg.Min, cell)
	cell.Rune = '┐'
	b.grid.SetCell(gruid.Position{X: rg.Width() - 1}, cell)
	cell.Rune = '└'
	b.grid.SetCell(gruid.Position{Y: rg.Height() - 1}, cell)
	cell.Rune = '┘'
	b.grid.SetCell(rg.Max.Shift(-1, -1), cell)
	cell.Rune = '│'
	col := b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(0))
	col.Iter(func(pos gruid.Position) {
		col.SetCell(pos, cell)
	})
	col = b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(rg.Width() - 1))
	col.Iter(func(pos gruid.Position) {
		col.SetCell(pos, cell)
	})
}
