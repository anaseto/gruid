package models

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// EntryKind represents the different kinds of entries.
type EntryKind int

// These constants define the available entry kinds.
const (
	EntryChoice EntryKind = iota
	EntryUnavailable
	EntryHeader
)

// MenuEntry represents an entry in the menu. It is displayed on one line, and
// for example can be a choice or a header.
type MenuEntry struct {
	Kind EntryKind // available choice, header line, etc.
	Text string    // displayed text on the entry line
	Key  gruid.Key // accept entry shortcut, if any
}

// MenuAction represents an user action with the menu.
type MenuAction int

// These constants represent the available actions in a menu.
const (
	MenuSelection MenuAction = iota // changed selection
	MenuAccept                      // accepted current selection
	MenuCancel                      // cancelled selection / close menu
	MenuNone                        // no action
)

// MenuStyle represents menu styles.
type MenuStyle struct {
	Boxed            bool        // draw a box around the menu
	ColorBg          gruid.Color // background color
	ColorBgAlt       gruid.Color // alternate bg on even entry lines
	ColorFg          gruid.Color // foreground color
	ColorAvailable   gruid.Color // normal entry choice
	ColorSelected    gruid.Color // selected entry
	ColorUnavailable gruid.Color // unavailable entry
	ColorHeader      gruid.Color // header entry
	ColorTitle       gruid.Color // box title
}

// MenuConfig contains configuration options for creating a menu.
type MenuConfig struct {
	Grid    gruid.Grid  // grid slice where the menu is drawn
	Entries []MenuEntry // menu entries
	Title   string      // optional title
	Style   MenuStyle
}

// NewMenu returns a menu with a given configuration.
func NewMenu(cfg MenuConfig) *Menu {
	m := &Menu{
		grid:    cfg.Grid,
		entries: cfg.Entries,
		title:   cfg.Title,
		style:   cfg.Style,
		draw:    true,
	}
	m.cursorAtFirstChoice()
	return m
}

// Menu is a widget that asks the user to select an option among a list of
// entries.
type Menu struct {
	grid    gruid.Grid
	entries []MenuEntry
	title   string
	style   MenuStyle
	cursor  int
	action  MenuAction
	draw    bool
}

// Selection return the index of the currently selected entry.
func (m *Menu) Selection() int {
	return m.cursor
}

// Action returns the last action performed in the menu.
func (m *Menu) Action() MenuAction {
	return m.action
}

// SetEntries updates the list of menu entries.
func (m *Menu) SetEntries(entries []MenuEntry) {
	m.entries = entries
	if m.cursor >= len(m.entries) {
		m.cursorAtLastChoice()
	}
}

// Update implements Model.Update
func (m *Menu) Update(msg gruid.Msg) gruid.Cmd {
	l := len(m.entries)
	m.draw = false
	m.action = MenuSelection // default action
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch {
		case msg.Key == gruid.KeyEscape || msg.Key == gruid.KeySpace || msg.Key == "x" || msg.Key == "X":
			m.action = MenuCancel
		case msg.Key == gruid.KeyArrowDown:
			m.cursor++
			for m.cursor < l && m.entries[m.cursor].Kind != EntryChoice {
				m.cursor++
			}
			if m.cursor >= l {
				m.cursorAtFirstChoice()
			}
			m.draw = true
		case msg.Key == gruid.KeyArrowUp:
			m.cursor--
			for m.cursor >= 0 && m.entries[m.cursor].Kind != EntryChoice {
				m.cursor--
			}
			if m.cursor < 0 {
				m.cursorAtLastChoice()
			}
			m.draw = true
		case msg.Key == gruid.KeyEnter && m.entries[m.cursor].Kind == EntryChoice:
			m.action = MenuAccept
			m.draw = true
		default:
			nchars := utf8.RuneCountInString(string(msg.Key))
			if nchars == 1 {
				for i, e := range m.entries {
					if e.Key == msg.Key {
						m.cursor = i
						m.action = MenuAccept
						break
					}
				}
			}
			m.draw = true
		}
	case gruid.MsgMouseMove:
		pos := msg.MousePos.Relative(m.grid.Range())
		if !m.grid.Contains(pos) {
			break
		}
		if m.isEdgePos(pos) {
			break
		}
		if pos.Y-1 == m.cursor {
			break
		}
		m.draw = true
		m.cursor = pos.Y - 1
	case gruid.MsgMouseDown:
		pos := msg.MousePos.Relative(m.grid.Range())
		switch msg.Button {
		case gruid.ButtonMain:
			if !m.grid.Contains(pos) {
				m.action = MenuCancel
				break
			}
			if m.isEdgePos(pos) {
				break
			}
			m.cursor = pos.Y - 1
			if m.entries[m.cursor].Kind == EntryChoice {
				m.action = MenuAccept
			}
			m.draw = true
		}
	}
	return nil
}

func (m *Menu) isEdgePos(pos gruid.Position) bool {
	return pos.X == 0 || pos.X == m.grid.Range().Width()-1 || pos.Y == 0 || pos.Y == m.grid.Range().Height()-1
}

func (m *Menu) cursorAtFirstChoice() {
	for i, c := range m.entries {
		if c.Kind == EntryChoice {
			m.cursor = i
			break
		}
	}
}

func (m *Menu) cursorAtLastChoice() {
	for i, c := range m.entries {
		if c.Kind == EntryChoice {
			m.cursor = i
		}
	}
}

// Draw implements Model.Draw.
func (m *Menu) Draw() gruid.Grid {
	if !m.draw {
		return m.grid
	}
	if m.style.Boxed {
		b := box{
			grid:    m.grid,
			title:   m.title,
			fg:      m.style.ColorFg,
			bg:      m.style.ColorBg,
			fgtitle: m.style.ColorTitle,
		}
		b.draw()
	}
	alt := false
	t := textline{}
	rg := m.grid.Range().Relative()
	cgrid := m.grid.Slice(rg.Shift(1, 1, -1, -1))
	crg := cgrid.Range().Relative()
	for i, c := range m.entries {
		bg := m.style.ColorBg
		if c.Kind != EntryHeader {
			fg := m.style.ColorFg
			if alt {
				bg = m.style.ColorBgAlt
			}
			alt = !alt
			if c.Kind == EntryUnavailable {
				fg = m.style.ColorUnavailable
			}
			if c.Kind == EntryChoice && i == m.cursor {
				fg = m.style.ColorSelected
			}
			nchars := utf8.RuneCountInString(c.Text)
			t.grid = cgrid.Slice(crg.Line(i))
			t.fg = fg
			t.bg = bg
			t.text = c.Text
			t.draw()
			cell := gruid.Cell{Bg: bg, Rune: ' '}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Iter(func(pos gruid.Position) {
				line.SetCell(pos, cell)
			})
		} else {
			alt = false
			fg := m.style.ColorHeader
			nchars := utf8.RuneCountInString(c.Text)
			t.fg = fg
			t.bg = bg
			t.grid = cgrid.Slice(crg.Line(i))
			t.text = c.Text
			t.draw()
			cell := gruid.Cell{Bg: bg, Rune: ' '}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Iter(func(pos gruid.Position) {
				line.SetCell(pos, cell)
			})
		}
	}
	return m.grid
}

type box struct {
	grid    gruid.Grid
	title   string
	fg      gruid.Color
	bg      gruid.Color
	fgtitle gruid.Color
}

func (b box) draw() {
	rg := b.grid.Range().Relative()
	if rg.Empty() {
		return
	}
	cgrid := b.grid.Slice(rg.Shift(1, 0, -1, 0))
	crg := cgrid.Range().Relative()
	t := textline{
		fg: b.fg,
		bg: b.bg,
	}
	cell := gruid.Cell{Fg: b.fg, Bg: b.bg}
	cell.Rune = '─'
	if b.title != "" {
		nchars := utf8.RuneCountInString(b.title)
		dist := (crg.Width() - nchars) / 2
		line := cgrid.Slice(crg.Line(0))
		line.Iter(func(pos gruid.Position) {
			line.SetCell(pos, cell)
		})
		t.fg = b.fgtitle
		t.grid = cgrid.Slice(crg.Line(0).Shift(dist, 0, 0, 0))
		t.text = b.title
		t.draw()
		line = cgrid.Slice(crg.Line(0).Shift(dist+nchars, 0, 0, 0))
		line.Iter(func(pos gruid.Position) {
			line.SetCell(pos, cell)
		})
	} else {
		line := cgrid.Slice(crg.Line(0))
		line.Iter(func(pos gruid.Position) {
			line.SetCell(pos, cell)
		})
	}
	line := cgrid.Slice(crg.Line(crg.Height() - 1))
	line.Iter(func(pos gruid.Position) {
		line.SetCell(pos, cell)
	})
	cell.Rune = '┌'
	b.grid.SetCell(rg.Min, cell)
	cell.Rune = '┐'
	b.grid.SetCell(gruid.Position{X: rg.Width() - 1}, cell)
	cell.Rune = '└'
	b.grid.SetCell(gruid.Position{Y: rg.Height() - 1}, cell)
	cell.Rune = '┘'
	b.grid.SetCell(rg.Max.Shift(-1, -1), cell)
	cell.Rune = '│'
	col := b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(0))
	col.Iter(func(pos gruid.Position) {
		col.SetCell(pos, cell)
	})
	col = b.grid.Slice(rg.Shift(0, 1, 0, -1).Column(rg.Width() - 1))
	col.Iter(func(pos gruid.Position) {
		col.SetCell(pos, cell)
	})
}

type textline struct {
	grid gruid.Grid
	text string
	fg   gruid.Color
	bg   gruid.Color
}

func (t textline) draw() {
	for i, r := range t.text {
		t.grid.SetCell(gruid.Position{X: i}, gruid.Cell{Fg: t.fg, Bg: t.bg, Rune: r})
	}
}
