package ui

import (
	"github.com/anaseto/gruid"
)

// Label represents a bunch of text in a grid. It may be boxed and provided
// with a title.
type Label struct {
	Content     StyledText // label's styled text content
	Box         *Box       // draw optional box around the label
	AdjustWidth bool       // reduce the width of the label if possible
}

// NewLabel returns a new label with given styled text and AdjustWidth set to
// true.
func NewLabel(content StyledText) *Label {
	lb := &Label{
		Content:     content,
		AdjustWidth: true,
	}
	return lb
}

// SetText updates the text for the label's content, using the same styling.
func (lb *Label) SetText(text string) {
	lb.Content = lb.Content.WithText(text)
}

func (lb *Label) drawGrid(gd gruid.Grid) gruid.Grid {
	max := lb.Content.Size()
	w, h := max.X, max.Y
	if lb.Box != nil {
		ts := lb.Box.Title.Size()
		if w < ts.X {
			w = ts.X
		}
	}
	if lb.Box != nil {
		h += 2 // borders height
		w += 2
	}
	if !lb.AdjustWidth {
		w = gd.Size().X
	}
	return gd.Slice(gruid.NewRange(0, 0, w, h))
}

// Draw draws the label into the given grid. It returns the grid slice that was
// drawn.
func (lb *Label) Draw(gd gruid.Grid) gruid.Grid {
	grid := lb.drawGrid(gd)
	cgrid := grid
	if lb.Box != nil {
		lb.Box.Draw(grid)
		rg := grid.Range()
		cgrid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	cgrid.Fill(gruid.Cell{Rune: ' ', Style: lb.Content.Style()})
	lb.Content.Draw(cgrid)
	return grid
}
