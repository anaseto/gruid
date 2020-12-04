package models

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// EntryKind represents the different kinds of entries.
type EntryKind int

// These constants define the available entry kinds.
const (
	EntryChoice      EntryKind = iota // a choice
	EntryUnavailable                  // an unavailable choice
	EntryHeader                       // a sub-header
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
	MenuPass   MenuAction = iota // no action
	MenuAccept                   // accepted current selection
	MenuCancel                   // cancelled selection / close menu
	MenuMove                     // changed selection
)

// MenuStyle represents menu styles.
type MenuStyle struct {
	Boxed       bool // draw a box around the menu
	Content     gruid.CellStyle
	BgAlt       gruid.Color     // alternate bg on even entry lines
	Selected    gruid.Color     // for selected entry
	Unavailable gruid.Color     // for unavailable entry
	Header      gruid.CellStyle // header entry
	Title       gruid.CellStyle // box title
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
	m.action = MenuPass // no action still
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch {
		case msg.Key == gruid.KeyEscape || msg.Key == gruid.KeySpace || msg.Key == "x" || msg.Key == "X":
			m.action = MenuCancel
		case msg.Key == gruid.KeyArrowDown:
			m.action = MenuMove
			m.cursor++
			for m.cursor < l && m.entries[m.cursor].Kind != EntryChoice {
				m.cursor++
			}
			if m.cursor >= l {
				m.cursorAtFirstChoice()
			}
			m.draw = true
		case msg.Key == gruid.KeyArrowUp:
			m.action = MenuMove
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
		m.action = MenuMove
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
			m.action = MenuMove
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
			grid:       m.grid,
			title:      m.title,
			style:      m.style.Content,
			titleStyle: m.style.Title,
		}
		b.draw()
	}
	alt := false
	rg := m.grid.Range().Relative()
	cgrid := m.grid.Slice(rg.Shift(1, 1, -1, -1))
	crg := cgrid.Range().Relative()
	for i, c := range m.entries {
		if c.Kind != EntryHeader {
			st := m.style.Content
			if alt {
				st.Bg = m.style.BgAlt
			}
			alt = !alt
			if c.Kind == EntryUnavailable {
				st.Fg = m.style.Unavailable
			}
			if c.Kind == EntryChoice && i == m.cursor {
				st.Fg = m.style.Selected
			}
			nchars := utf8.RuneCountInString(c.Text)
			stt := NewStyledText(c.Text)
			stt.SetStyle(st)
			stt.Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Iter(func(pos gruid.Position) {
				line.SetCell(pos, cell)
			})
		} else {
			alt = false
			st := m.style.Header
			nchars := utf8.RuneCountInString(c.Text)
			stt := NewStyledText(c.Text)
			stt.SetStyle(st)
			stt.Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Iter(func(pos gruid.Position) {
				line.SetCell(pos, cell)
			})
		}
	}
	return m.grid
}
