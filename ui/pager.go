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

// PagerKeys contains key bindings configuration for the pager.
type PagerKeys struct {
	Down         []gruid.Key // go one line down
	Up           []gruid.Key // go one line up
	PageDown     []gruid.Key // go one page down
	PageUp       []gruid.Key // go one page up
	HalfPageDown []gruid.Key // go half page down
	HalfPageUp   []gruid.Key // go half page up
	Top          []gruid.Key // go to the top
	Bottom       []gruid.Key // go to the bottom
	Quit         []gruid.Key // quit pager
}

// PagerConfig describes configuration options for creating a pager.
type PagerConfig struct {
	Grid       gruid.Grid // grid slice where the viewable content is drawn
	StyledText StyledText // styled text for markup rendering
	Lines      []string   // content lines to be read
	Title      string     // optional title, implies Boxed style
	Keys       PagerKeys  // optional custom key bindings for the pager
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
	init   bool // Update received MsgInit
	keys   PagerKeys
}

// NewPager returns a new pager with given configuration options.
func NewPager(cfg PagerConfig) *Pager {
	pg := &Pager{
		grid:  cfg.Grid,
		stt:   cfg.StyledText,
		title: cfg.Title,
		lines: cfg.Lines,
		style: cfg.Style,
		keys:  cfg.Keys,
	}
	if pg.keys.Down == nil {
		pg.keys.Down = []gruid.Key{gruid.KeyArrowDown, "j"}
	}
	if pg.keys.Up == nil {
		pg.keys.Up = []gruid.Key{gruid.KeyArrowUp, "k"}
	}
	if pg.keys.PageDown == nil {
		pg.keys.PageDown = []gruid.Key{gruid.KeyPageDown, "f"}
	}
	if pg.keys.PageUp == nil {
		pg.keys.Up = []gruid.Key{gruid.KeyPageUp, "b"}
	}
	if pg.keys.HalfPageDown == nil {
		pg.keys.HalfPageDown = []gruid.Key{gruid.KeyEnter, "d"}
	}
	if pg.keys.HalfPageUp == nil {
		pg.keys.HalfPageUp = []gruid.Key{gruid.KeyBackspace, "u"}
	}
	if pg.keys.Top == nil {
		pg.keys.Top = []gruid.Key{gruid.KeyHome, "g"}
	}
	if pg.keys.Bottom == nil {
		pg.keys.Bottom = []gruid.Key{gruid.KeyEnd, "G"}
	}
	if pg.keys.Quit == nil {
		pg.keys.Quit = []gruid.Key{gruid.KeyEscape, "q", "Q"}
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
		pg.init = true
		return nil
	case gruid.MsgKeyDown:
		key := msg.Key
		switch {
		case key.In(pg.keys.Down):
			if pg.index+nlines < len(pg.lines) {
				pg.action = PagerMove
				pg.index++
			}
		case key.In(pg.keys.Up):
			if pg.index > 0 {
				pg.action = PagerMove
				pg.index--
			}
		case key.In(pg.keys.PageDown), key.In(pg.keys.HalfPageDown):
			shift := nlines - 1
			if key.In(pg.keys.HalfPageDown) {
				shift /= 2
			}
			if pg.index+nlines+shift-1 >= len(pg.lines) {
				shift = len(pg.lines) - pg.index - nlines
			}
			if shift > 0 {
				pg.action = PagerMove
				pg.index += shift
			}
		case key.In(pg.keys.PageUp), key.In(pg.keys.HalfPageUp):
			shift := nlines - 1
			if key.In(pg.keys.HalfPageUp) {
				shift /= 2
			}
			if pg.index-shift < 0 {
				shift = pg.index
			}
			if shift > 0 {
				pg.action = PagerMove
				pg.index -= shift
			}
		case key.In(pg.keys.Top):
			if pg.index != 0 {
				pg.index = 0
				pg.action = PagerMove
			}
		case key.In(pg.keys.Bottom):
			if pg.index != len(pg.lines)-nlines {
				pg.index = len(pg.lines) - nlines
				pg.action = PagerMove
			}
		case key.In(pg.keys.Quit):
			pg.action = PagerQuit
			if pg.init {
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

// Draw implements gruid.Model.Draw for Pager. It returns the grid slice that
// was drawn.
func (pg *Pager) Draw() gruid.Grid {
	if pg.init {
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
