package ui

import (
	"github.com/anaseto/gruid"
)

// LabelConfig describes configuration options for creating a label.
type LabelConfig struct {
	StyledText  StyledText // styled text with initial label text content
	Box         *Box       // draw optional box around the  label
	AdjustWidth bool       // reduce the width of the label if possible
}

// Label represents a bunch of text in a grid. It may be boxed and provided
// with a title.
type Label struct {
	stt  StyledText
	box  *Box
	adjw bool
}

// NewLabel returns a new label with given configuration options.
func NewLabel(cfg LabelConfig) *Label {
	lb := &Label{
		stt:  cfg.StyledText,
		box:  cfg.Box,
		adjw: cfg.AdjustWidth,
	}
	return lb
}

// SetStyledText updates the styled text for the label.
func (lb *Label) SetStyledText(stt StyledText) {
	lb.stt = stt
}

// SetText updates the text for the label.
func (lb *Label) SetText(text string) {
	lb.stt = lb.stt.WithText(text)
}

// Text returns the text currently used by the label.
func (lb *Label) Text() string {
	return lb.stt.Text()
}

func (lb *Label) drawGrid(gd gruid.Grid) gruid.Grid {
	max := lb.stt.Size()
	w, h := max.X, max.Y
	if lb.box != nil {
		ts := lb.box.Title.Size()
		if w < ts.X {
			w = ts.X
		}
	}
	if lb.box != nil {
		h += 2 // borders height
		w += 2
	}
	if !lb.adjw {
		w = gd.Range().Size().X
	}
	return gd.Slice(gruid.NewRange(0, 0, w, h))
}

// Draw draws the label into the grid. It returns the grid slice that was
// drawn.
func (lb *Label) Draw(gd gruid.Grid) gruid.Grid {
	grid := lb.drawGrid(gd)
	cgrid := grid
	if lb.box != nil {
		lb.box.Draw(grid)
		rg := grid.Range().Origin()
		cgrid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	cgrid.Fill(gruid.Cell{Rune: ' ', Style: lb.stt.Style()})
	lb.stt.Draw(cgrid)
	return grid
}
