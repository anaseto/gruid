package ui

import (
	"strings"

	"github.com/anaseto/gruid"
)

// LabelStyle describes styling options for a Label.
type LabelStyle struct {
	Content gruid.CellStyle
	Title   gruid.CellStyle
}

// Label represents a bunch of text in a grid. It may be boxed and provided
// with a title.
type Label struct {
	Boxed bool
	Grid  gruid.Grid
	Title string
	Text  string
	Style LabelStyle
}

// Draw draws the label into the grid.
func (l Label) Draw() gruid.Grid {
	height := strings.Count(l.Text, "\n") + 1
	gr := l.Grid
	rg := gr.Range().Relative()
	lh := height
	if l.Boxed {
		lh += 2
	}
	if rg.Height() > lh {
		gr = gr.Slice(gruid.NewRange(0, 0, rg.Max.X, lh))
	}
	var tgrid gruid.Grid
	if l.Boxed {
		b := box{
			grid:       gr,
			title:      l.Title,
			style:      l.Style.Content,
			titleStyle: l.Style.Title,
		}
		b.draw()
		rg := gr.Range().Relative()
		tgrid = gr.Slice(rg.Shift(1, 1, -1, -1))
	} else {
		tgrid = gr
	}
	tgrid.Iter(func(pos gruid.Position) {
		tgrid.SetCell(pos, gruid.Cell{Rune: ' ', Style: l.Style.Content})
	})
	NewStyledText(l.Text).WithStyle(l.Style.Content).Draw(tgrid)
	return gr
}
