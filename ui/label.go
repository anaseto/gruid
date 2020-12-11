package ui

import (
	"github.com/anaseto/gruid"
)

// LabelStyle describes styling options for a Label.
type LabelStyle struct {
	Boxed       bool        // draw a box around the label
	Box         gruid.Style // box style, if any
	Title       gruid.Style // box title style, if any
	AdjustWidth bool        // reduce the width of the box if possible
}

// LabelConfig describes configuration options for creating a label.
type LabelConfig struct {
	Grid       gruid.Grid // grid slice where the label is drawn
	StyledText StyledText // styled text with initial label text content
	Title      string     // optional title, implies Boxed style
	Style      LabelStyle
}

// Label represents a bunch of text in a grid. It may be boxed and provided
// with a title.
type Label struct {
	grid  gruid.Grid
	stt   StyledText
	title string
	style LabelStyle
}

// NewLabel returns a new label with given configuration options.
func NewLabel(cfg LabelConfig) *Label {
	lb := &Label{
		grid:  cfg.Grid,
		stt:   cfg.StyledText,
		title: cfg.Title,
		style: cfg.Style,
	}
	if lb.title != "" {
		lb.style.Boxed = true
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
	w, h := lb.stt.Size() // text height
	if !lb.style.AdjustWidth {
		w, _ = lb.grid.Range().Size()
	}
	tw, _ := lb.stt.WithText(lb.title).Size()
	if w < tw {
		w = tw
	}
	if lb.style.Boxed {
		h += 2 // borders height
		w += 2
	}
	return lb.grid.Slice(gruid.NewRange(0, 0, w, h))
}

// Draw draws the label into the grid. It returns the grid slice that was
// drawn.
func (lb *Label) Draw() gruid.Grid {
	grid := lb.drawGrid()
	cgrid := grid
	if lb.style.Boxed {
		b := box{
			grid:  grid,
			title: lb.stt.With(lb.title, lb.style.Title),
			style: lb.style.Box,
		}
		b.draw()
		rg := grid.Range().Relative()
		cgrid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	cgrid.Fill(gruid.Cell{Rune: ' ', Style: lb.stt.Style()})
	lb.stt.Draw(cgrid)
	return grid
}
