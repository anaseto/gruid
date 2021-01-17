package ui

import (
	"fmt"

	"github.com/anaseto/gruid"
)

// PagerConfig describes configuration options for creating a pager.
type PagerConfig struct {
	Grid  gruid.Grid   // grid slice where the viewable content is drawn
	Lines []StyledText // content lines to be read
	Box   *Box         // draw optional box around the  label
	Keys  PagerKeys    // optional custom key bindings for the pager
	Style PagerStyle
}

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

// Pager represents a pager widget for viewing a long list of lines.
//
// Pager implements gruid.Model and can be used as main model of an
// application.
type Pager struct {
	grid   gruid.Grid
	box    *Box
	lines  []StyledText
	style  PagerStyle
	index  int // current index
	x      int // x position
	action PagerAction
	init   bool // Update received MsgInit
	keys   PagerKeys
	dirty  bool       // state changed in Update and Draw was still not called
	drawn  gruid.Grid // last drawn grid slice
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

// NewPager returns a new pager with given configuration options.
func NewPager(cfg PagerConfig) *Pager {
	pg := &Pager{
		grid:  cfg.Grid,
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
	pg.dirty = true
	return pg
}

// SetBox updates the pager surrounding box.
func (pg *Pager) SetBox(b *Box) {
	pg.box = b
	pg.dirty = true
}

// SetLines updates the pager text lines.
func (pg *Pager) SetLines(lines []StyledText) {
	nlines := pg.nlines()
	pg.lines = lines
	if pg.index+nlines-1 >= len(pg.lines) {
		pg.index = len(pg.lines) - nlines
		if pg.index <= 0 {
			pg.index = 0
		}
	}
	pg.dirty = true
}

func (pg *Pager) nlines() int {
	nlines := pg.grid.Size().Y
	if pg.box != nil {
		nlines -= 2
	}
	return nlines
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
	nlines := pg.nlines()
	if pg.index+nlines+shift-1 >= len(pg.lines) {
		shift = len(pg.lines) - pg.index - nlines
	}
	if shift > 0 {
		pg.action = PagerMove
		pg.index += shift
	}
}

func (pg *Pager) up(shift int) {
	if pg.index-shift < 0 {
		shift = pg.index
	}
	if shift > 0 {
		pg.action = PagerMove
		pg.index -= shift
	}
}

func (pg *Pager) right() {
	pg.action = PagerMove
	pg.x += 8
}

func (pg *Pager) left() {
	if pg.x > 0 {
		pg.action = PagerMove
		pg.x -= 8
		if pg.x <= 0 {
			pg.x = 0
		}
	}
}

func (pg *Pager) lineStart() {
	if pg.x > 0 {
		pg.action = PagerMove
		pg.x = 0
	}
}

func (pg *Pager) top() {
	if pg.index != 0 {
		pg.index = 0
		pg.action = PagerMove
	}
}

func (pg *Pager) bottom() {
	nlines := pg.nlines()
	if pg.index != len(pg.lines)-nlines {
		pg.index = len(pg.lines) - nlines
		pg.action = PagerMove
	}
}

// Update implements gruid.Model.Update for Pager. It considers mouse message
// coordinates to be absolute in its grid. If a gruid.MsgInit is passed to
// Update, the pager will behave as if it is the main model of an application,
// and send a gruid.Quit() command on PagerQuit action.
func (pg *Pager) Update(msg gruid.Msg) gruid.Effect {
	pg.action = PagerPass
	var eff gruid.Effect
	switch msg := msg.(type) {
	case gruid.MsgInit:
		pg.init = true
	case gruid.MsgKeyDown:
		eff = pg.updateMsgKeyDown(msg)
	case gruid.MsgMouse:
		eff = pg.updateMsgMouse(msg)
	}
	if pg.Action() != PagerPass {
		pg.dirty = true
	}
	return eff
}

func (pg *Pager) updateMsgKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	key := msg.Key
	switch {
	case key.In(pg.keys.Down):
		pg.down(1)
	case key.In(pg.keys.Up):
		pg.up(1)
	case key.In(pg.keys.Left):
		pg.left()
	case key.In(pg.keys.Right):
		pg.right()
	case key.In(pg.keys.Start):
		pg.lineStart()
	case key.In(pg.keys.PageDown), key.In(pg.keys.HalfPageDown):
		shift := pg.nlines() - 1
		if key.In(pg.keys.HalfPageDown) {
			shift /= 2
		}
		pg.down(shift)
	case key.In(pg.keys.PageUp), key.In(pg.keys.HalfPageUp):
		shift := pg.nlines() - 1
		if key.In(pg.keys.HalfPageUp) {
			shift /= 2
		}
		pg.up(shift)
	case key.In(pg.keys.Top):
		pg.top()
	case key.In(pg.keys.Bottom):
		pg.bottom()
	case key.In(pg.keys.Quit):
		pg.action = PagerQuit
		if pg.init {
			return gruid.End()
		}
	}
	return nil
}

func (pg *Pager) updateMsgMouse(msg gruid.MsgMouse) gruid.Effect {
	nlines := pg.grid.Size().Y
	if pg.box != nil {
		nlines -= 2
	}
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
	return nil
}

// Action returns the last action performed with the pager.
func (pg *Pager) Action() PagerAction {
	return pg.action
}

// Draw implements gruid.Model.Draw for Pager. It returns the grid slice that
// was drawn, or the whole grid if it is used as main model.
func (pg *Pager) Draw() gruid.Grid {
	if !pg.dirty {
		if pg.init {
			return pg.drawn.Slice(gruid.Range{})
		}
		return pg.drawn
	}
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
	if pg.init {
		pg.grid.Fill(gruid.Cell{Rune: ' '})
	}
	cgrid := grid
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
		NewStyledText(lnumtext, pg.style.LineNum).Draw(line)
		cgrid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	rg := cgrid.Range()
	for i := 0; i < h-bh; i++ {
		line := cgrid.Slice(rg.Line(i))
		stt := pg.lines[i+pg.index]
		line.Fill(gruid.Cell{Rune: ' ', Style: stt.Style()})
		stt.Iter(func(p gruid.Point, c gruid.Cell) {
			p = p.Shift(-pg.x, 0)
			if p.X >= 0 {
				line.Set(p, c)
			}
		})
	}
	pg.dirty = false
	pg.drawn = grid
	if pg.init {
		pg.drawn = pg.grid
	}
	return pg.drawn
}
