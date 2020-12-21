package ui

import (
	"github.com/anaseto/gruid"
)

// Label represents a bunch of text in a grid. It may be boxed and provided
// with a title.
type Label struct {
	StyledText  StyledText // styled text with initial label text content
	Box         *Box       // draw optional box around the  label
	AdjustWidth bool       // reduce the width of the label if possible
}

// NewLabel returns a new label with given styled text and AdjustWidth set to
// true.
func NewLabel(stt StyledText) *Label {
	lb := &Label{
		StyledText:  stt,
		AdjustWidth: true,
	}
	return lb
}

// SetText updates the text for the label.
func (lb *Label) SetText(text string) {
	lb.StyledText = lb.StyledText.WithText(text)
}

func (lb *Label) drawGrid(gd gruid.Grid) gruid.Grid {
	max := lb.StyledText.Size()
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
	cgrid.Fill(gruid.Cell{Rune: ' ', Style: lb.StyledText.Style()})
	lb.StyledText.Draw(cgrid)
	return grid
}
