package ui

import (
	"strings"

	"github.com/anaseto/gruid"
)

type LabelStyle struct {
	Content gruid.CellStyle
	Title   gruid.CellStyle
}

type Label struct {
	Boxed bool
	Grid  gruid.Grid
	Title string
	Text  string
	Style LabelStyle
}

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
	st := NewStyledText(l.Text)
	st.SetStyle(l.Style.Content)
	st.Draw(tgrid)
	return gr
}
