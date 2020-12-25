package ui

import (
	"fmt"

	"github.com/anaseto/gruid"
)

// PagerStyle describes styling options for a Pager.
type PagerStyle struct {
	LineNum gruid.Style // line num display style (for boxed pager)
}

// PagerKeys contains key bindings configuration for the pager.
type PagerKeys struct {
	Down         []gruid.Key // go one line down (default: ArrowDown, j)
	Up           []gruid.Key // go one line up (default: ArrowUp, k)
	Left         []gruid.Key // go left (default: ArrowLeft, h)
	Right        []gruid.Key // go right (default: ArrowRight, l)
	Start        []gruid.Key // go to start of line (default: 0, ^)
	PageDown     []gruid.Key // go one page down (default: PageDown, f)
	PageUp       []gruid.Key // go one page up (default: PageUp, b)
	HalfPageDown []gruid.Key // go half page down (default: Enter, d)
	HalfPageUp   []gruid.Key // go half page up (default: Backspace, u)
	Top          []gruid.Key // go to the top (default: Home, g)
	Bottom       []gruid.Key // go to the bottom (default: End, G)
	Quit         []gruid.Key // quit pager (default: Escape, q, Q)
}

// PagerConfig describes configuration options for creating a pager.
type PagerConfig struct {
	Grid       gruid.Grid // grid slice where the viewable content is drawn
	StyledText StyledText // styled text for markup rendering
	Lines      []string   // content lines to be read
	Box        *Box       // draw optional box around the  label
	Keys       PagerKeys  // optional custom key bindings for the pager
	Style      PagerStyle
}

// Pager represents a pager widget for viewing a long list of lines.
type Pager struct {
	grid   gruid.Grid
	stt    StyledText
	box    *Box
	lines  []string
	style  PagerStyle
	index  int // current index
	x      int // x position
	action PagerAction
	init   bool // Update received MsgInit
	keys   PagerKeys
}

// NewPager returns a new pager with given configuration options.
func NewPager(cfg PagerConfig) *Pager {
	pg := &Pager{
		grid:  cfg.Grid,
		stt:   cfg.StyledText,
		box:   cfg.Box,
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
		pg.keys.PageUp = []gruid.Key{gruid.KeyPageUp, "b"}
	}
	if pg.keys.Left == nil {
		pg.keys.Left = []gruid.Key{gruid.KeyArrowLeft, "h"}
	}
	if pg.keys.Right == nil {
		pg.keys.Right = []gruid.Key{gruid.KeyArrowRight, "l"}
	}
	if pg.keys.Start == nil {
		pg.keys.Start = []gruid.Key{"^", "0"}
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

// SetBox updates the pager surrounding box.
func (pg *Pager) SetBox(b *Box) {
	pg.box = b
}

// SetLines updates the pager text lines.
func (pg *Pager) SetLines(lines []string) {
	nlines := pg.grid.Size().Y
	if pg.box != nil {
		nlines -= 2
	}
	pg.lines = lines
	if pg.index+nlines-1 >= len(pg.lines) {
		pg.index = len(pg.lines) - nlines
	}
}

// View returns a range (Min, Max) such that the currently displayed lines are
// the lines whose index is between Min.Y and Max.Y, and the displayed text
// columns are the ones between Min.X and Max.X.
func (pg *Pager) View() gruid.Range {
	size := pg.grid.Size()
	h := size.Y
	bh := 0
	if pg.box != nil {
		bh = 2
	}
	if h > bh+len(pg.lines) {
		h = bh + len(pg.lines)
	}
	if h-bh <= 0 {
		return gruid.Range{}
	}
	return gruid.NewRange(pg.x, pg.index, pg.x+size.X, pg.index+h-bh)
}

func (pg *Pager) down(shift int) {
	nlines := pg.grid.Size().Y
	if pg.box != nil {
		nlines -= 2
	}
	if pg.index+nlines+shift-1 >= len(pg.lines) {
		shift = len(pg.lines) - pg.index - nlines
	}
	if shift > 0 {
		pg.action = PagerMove
		pg.index += shift
	}
}

func (pg *Pager) up(shift int) {
	nlines := pg.grid.Size().Y
	if pg.box != nil {
		nlines -= 2
	}
	if pg.index-shift < 0 {
		shift = pg.index
	}
	if shift > 0 {
		pg.action = PagerMove
		pg.index -= shift
	}
}

// Update implements gruid.Model.Update for Pager. It considers mouse message
// coordinates to be absolute in its grid.
func (pg *Pager) Update(msg gruid.Msg) gruid.Effect {
	nlines := pg.grid.Size().Y
	if pg.box != nil {
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
			pg.down(1)
		case key.In(pg.keys.Up):
			pg.up(1)
		case key.In(pg.keys.Left):
			if pg.x > 0 {
				pg.action = PagerMove
				pg.x -= 8
				if pg.x <= 0 {
					pg.x = 0
				}
			}
		case key.In(pg.keys.Right):
			pg.action = PagerMove
			pg.x += 8
		case key.In(pg.keys.Start):
			if pg.x > 0 {
				pg.action = PagerMove
				pg.x = 0
			}
		case key.In(pg.keys.PageDown), key.In(pg.keys.HalfPageDown):
			shift := nlines - 1
			if key.In(pg.keys.HalfPageDown) {
				shift /= 2
			}
			pg.down(shift)
		case key.In(pg.keys.PageUp), key.In(pg.keys.HalfPageUp):
			shift := nlines - 1
			if key.In(pg.keys.HalfPageUp) {
				shift /= 2
			}
			pg.up(shift)
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
				return gruid.End()
			}
		}
	case gruid.MsgMouse:
		if !msg.P.In(pg.grid.Bounds()) {
			switch msg.Action {
			case gruid.MouseMain:
				pg.action = PagerQuit
			}
			return nil
		}
		switch msg.Action {
		case gruid.MouseMain:
			if msg.P.Sub(pg.grid.Bounds().Min).Y > nlines/2 {
				pg.down(nlines - 1)
			} else {
				pg.up(nlines - 1)
			}
		case gruid.MouseWheelUp:
			pg.up(1)
		case gruid.MouseWheelDown:
			pg.down(1)
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
	grid := pg.grid
	max := grid.Size()
	w, h := max.X, max.Y
	bh := 0
	if pg.box != nil {
		bh = 2
	}
	if h > bh+len(pg.lines) {
		h = bh + len(pg.lines)
		grid = grid.Slice(gruid.NewRange(0, 0, w, h))
	}
	grid.Fill(gruid.Cell{Rune: ' ', Style: pg.stt.Style()})
	if pg.box != nil {
		pg.box.Draw(grid)
		rg := grid.Range()
		line := grid.Slice(rg.Line(h-1).Shift(2, 0, -2, 0))
		var lnumtext string
		if pg.x > 0 {
			lnumtext = fmt.Sprintf("%d-%d/%d+%d", pg.index, pg.index+h-bh-1, len(pg.lines)-1, pg.x)
		} else {
			lnumtext = fmt.Sprintf("%d-%d/%d", pg.index, pg.index+h-bh-1, len(pg.lines)-1)
		}
		pg.stt.With(lnumtext, pg.style.LineNum).Draw(line)
		grid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	rg := grid.Range()
	for i := 0; i < h-bh; i++ {
		line := grid.Slice(rg.Line(i))
		s := pg.lines[i+pg.index]
		count := 0
		vs := s
		for i, _ := range s {
			if count <= pg.x {
				vs = s[i:]
			} else {
				break
			}
			count++
		}
		if count >= pg.x {
			pg.stt.WithText(vs).Draw(line)
		}
	}
	return grid
}
