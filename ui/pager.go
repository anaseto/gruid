package ui

import (
	"fmt"

	"github.com/anaseto/gruid"
)

// PagerStyle describes styling options for a Pager.
type PagerStyle struct {
	Boxed   bool        // draw a box around the pager
	Box     gruid.Style // box style, if any
	Title   gruid.Style // box title style
	LineNum gruid.Style // line num display style
}

// PagerConfig describes configuration options for creating a pager.
type PagerConfig struct {
	Grid       gruid.Grid // grid slice where the viewable content is drawn
	StyledText StyledText // styled text for markup rendering
	Lines      []string   // content lines to be read
	Title      string     // optional title, implies Boxed style
	Style      PagerStyle
}

// Pager represents a pager widget for viewing a long list of lines.
type Pager struct {
	grid   gruid.Grid
	stt    StyledText
	title  string
	lines  []string
	style  PagerStyle
	index  int // current index
	action PagerAction
	app    bool // running as main model
}

// NewPager returns a new pager with given configuration options.
func NewPager(cfg PagerConfig) *Pager {
	pg := &Pager{
		grid:  cfg.Grid,
		stt:   cfg.StyledText,
		title: cfg.Title,
		lines: cfg.Lines,
		style: cfg.Style,
	}
	if pg.title != "" {
		pg.style.Boxed = true
	}
	return pg
}

// PagerAction represents an user action with the pager.
type PagerAction int

const (
	// PagerPass reports that the pager state did not change (for example a
	// mouse motion outside the menu, or within a same entry line).
	PagerPass PagerAction = iota

	// PagerMove reports a scrolling movement.
	PagerMove

	// PagerQuit reports that the user clicked outside the menu, or pressed
	// Esc, Space or X.
	PagerQuit
)

// Update implements gruid.Model.Update for Pager.
func (pg *Pager) Update(msg gruid.Msg) gruid.Effect {
	_, nlines := pg.grid.Size()
	if pg.style.Boxed {
		nlines -= 2
	}
	pg.action = PagerPass
	switch msg := msg.(type) {
	case gruid.MsgInit:
		pg.app = true
		return nil
	case gruid.MsgKeyDown:
		switch msg.Key {
		case gruid.KeyArrowDown, "j":
			if pg.index+nlines < len(pg.lines) {
				pg.action = PagerMove
				pg.index++
			}
		case gruid.KeyArrowUp, "k":
			if pg.index > 0 {
				pg.action = PagerMove
				pg.index--
			}
		case gruid.KeyPageDown, gruid.KeySpace, "f", "d":
			shift := nlines - 1
			if msg.Key == "d" {
				shift /= 2
			}
			if pg.index+nlines+shift-1 >= len(pg.lines) {
				shift = len(pg.lines) - pg.index - nlines
			}
			if shift > 0 {
				pg.action = PagerMove
				pg.index += shift
			}
		case gruid.KeyPageUp, gruid.KeyBackspace, "b", "u":
			shift := nlines - 1
			if msg.Key == "u" {
				shift /= 2
			}
			if pg.index-shift < 0 {
				shift = pg.index
			}
			if shift > 0 {
				pg.action = PagerMove
				pg.index -= shift
			}
		case "g":
			if pg.index != 0 {
				pg.index = 0
				pg.action = PagerMove
			}
		case "G":
			if pg.index != len(pg.lines)-nlines {
				pg.index = len(pg.lines) - nlines
				pg.action = PagerMove
			}
		case gruid.KeyEscape, "q", "Q":
			pg.action = PagerQuit
			if pg.app {
				return gruid.Quit()
			}
		}
	}
	return nil
}

// Action returns the last action performed with the pager.
func (pg *Pager) Action() PagerAction {
	return pg.action
}

// Update implements gruid.Model.Draw for Pager. It returns the grid slice that
// was drawn.
func (pg *Pager) Draw() gruid.Grid {
	if pg.app {
		pg.grid.Fill(gruid.Cell{Rune: ' '})
	}
	grid := pg.grid
	w, h := grid.Size()
	bh := 0
	if pg.style.Boxed {
		bh = 2
	}
	if h > bh+len(pg.lines) {
		h = bh + len(pg.lines)
		grid = grid.Slice(gruid.NewRange(0, 0, w, h))
	}
	if pg.style.Boxed {
		b := box{
			grid:  grid,
			title: pg.stt.With(pg.title, pg.style.Title),
			style: pg.style.Box,
		}
		b.draw()
		rg := grid.Range().Relative()
		line := grid.Slice(rg.Line(h-1).Shift(2, 0, -2, 0))
		lnumtext := fmt.Sprintf("%d-%d/%d", pg.index, pg.index+h-bh-1, len(pg.lines)-1)
		pg.stt.With(lnumtext, pg.style.LineNum).Draw(line)
		grid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	rg := grid.Range().Relative()
	for i := 0; i < h-bh; i++ {
		line := grid.Slice(rg.Line(i))
		pg.stt.WithText(pg.lines[i+pg.index]).Draw(line)
	}
	return grid
}
