package ui

import (
	"github.com/anaseto/gruid"
)

// LabelConfig describes configuration options for creating a label.
type LabelConfig struct {
	Grid        gruid.Grid // grid slice where the label is drawn // TODO: remove
	StyledText  StyledText // styled text with initial label text content
	Box         *Box       // draw optional box around the  label
	AdjustWidth bool       // reduce the width of the label if possible
}

// Label represents a bunch of text in a grid. It may be boxed and provided
// with a title.
type Label struct {
	grid gruid.Grid
	stt  StyledText
	box  *Box
	adjw bool
}

// NewLabel returns a new label with given configuration options.
func NewLabel(cfg LabelConfig) *Label {
	lb := &Label{
		grid: cfg.Grid,
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

func (lb *Label) drawGrid() gruid.Grid {
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
		w = lb.grid.Range().Size().X
	}
	return lb.grid.Slice(gruid.NewRange(0, 0, w, h))
}

// Draw draws the label into the grid. It returns the grid slice that was
// drawn.
func (lb *Label) Draw() gruid.Grid {
	grid := lb.drawGrid()
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
