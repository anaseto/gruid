package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// MenuEntry represents an entry in the menu. It is displayed on one line, and
// for example can be a choice or a header.
type MenuEntry struct {
	Text   string    // displayed text on the entry line
	Key    gruid.Key // accept entry shortcut, if any and activable
	Header bool      // not an activable entry, but a sub-header entry
}

// MenuAction represents an user action with the menu.
type MenuAction int

// These constants represent the available actions in a menu.
const (
	// MenuPass reports that the menu state did not change (for example a
	// mouse motion outside the menu, or within a same entry line).
	MenuPass MenuAction = iota

	// MenuActivate reports that the user clicked or pressed enter to
	// activate/accept a selected entry, or used a shortcut key to select
	// and activate/accept a specific entry.
	MenuActivate

	// MenuCancel reports that the user clicked outside the menu, or
	// pressed Esc, Space or X.
	MenuCancel

	// MenuMove reports that the user moved the selection cursor by using
	// the arrow keys or a mouse motion.
	MenuMove
)

// MenuStyle represents menu styles.
type MenuStyle struct {
	BgAlt    gruid.Color     // alternate background on even choice lines
	Selected gruid.Color     // foreground for selected entry
	Header   gruid.CellStyle // header entry
	Boxed    bool            // draw a box around the menu
	Box      gruid.CellStyle // box style, if any
	Title    gruid.CellStyle // box title style, if any
}

// MenuConfig contains configuration options for creating a menu.
type MenuConfig struct {
	Grid       gruid.Grid  // grid slice where the menu is drawn
	Entries    []MenuEntry // menu entries
	Title      string      // optional title
	StyledText StyledText  // default styled text formatter for content
	Style      MenuStyle
}

// NewMenu returns a menu with a given configuration.
func NewMenu(cfg MenuConfig) *Menu {
	m := &Menu{
		grid:    cfg.Grid,
		entries: cfg.Entries,
		title:   cfg.Title,
		stt:     cfg.StyledText,
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
	stt     StyledText
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

// SetCursor updates the cursor selection index among entries. It may be used
// to launch the menu at a specific default starting index.
func (m *Menu) SetCursor(n int) {
	if n < 0 || n >= len(m.entries) {
		return
	}
	m.cursor = n
}

// Update implements gruid.Model.Update and updates the menu state in response to
// user input messages.
func (m *Menu) Update(msg gruid.Msg) gruid.Cmd {
	l := len(m.entries)
	m.action = MenuPass // no action still
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch {
		case msg.Key == gruid.KeyEscape || msg.Key == gruid.KeySpace || msg.Key == "x" || msg.Key == "X":
			m.action = MenuCancel
		case msg.Key == gruid.KeyArrowDown:
			m.action = MenuMove
			m.cursor++
			for m.cursor < l && m.entries[m.cursor].Header {
				m.cursor++
			}
			if m.cursor >= l {
				m.cursorAtFirstChoice()
			}
		case msg.Key == gruid.KeyArrowUp:
			m.action = MenuMove
			m.cursor--
			for m.cursor >= 0 && m.entries[m.cursor].Header {
				m.cursor--
			}
			if m.cursor < 0 {
				m.cursor = 0
				m.cursorAtLastChoice()
			}
		case msg.Key == gruid.KeyEnter && m.cursor < l && !m.entries[m.cursor].Header:
			m.action = MenuActivate
		default:
			nchars := utf8.RuneCountInString(string(msg.Key))
			if nchars == 1 {
				for i, e := range m.entries {
					if e.Key == msg.Key {
						m.cursor = i
						m.action = MenuActivate
						break
					}
				}
			}
		}
	case gruid.MsgMouseMove:
		rg := m.grid.Range()
		if m.style.Boxed {
			rg = rg.Shift(1, 1, -1, -1)
		}
		pos := msg.MousePos.Relative(rg)
		if !pos.In(rg.Relative()) || pos.Y >= len(m.entries) {
			break
		}
		if pos.Y == m.cursor {
			break
		}
		m.cursor = pos.Y
		m.action = MenuMove
	case gruid.MsgMouseDown:
		rg := m.grid.Range()
		if m.style.Boxed {
			rg = rg.Shift(1, 1, -1, -1)
		}
		pos := msg.MousePos.Relative(rg)
		switch msg.Button {
		case gruid.ButtonMain:
			if !msg.MousePos.In(m.grid.Range()) || !m.style.Boxed && pos.Y >= len(m.entries) {
				m.action = MenuCancel
				break
			}
			if !pos.In(rg.Relative()) || pos.Y >= len(m.entries) {
				break
			}
			m.cursor = pos.Y
			m.action = MenuMove
			if !m.entries[m.cursor].Header {
				m.action = MenuActivate
			}
		}
	}
	return nil
}

func (m *Menu) cursorAtFirstChoice() {
	m.cursor = 0
	for i, c := range m.entries {
		if !c.Header {
			m.cursor = i
			break
		}
	}
}

func (m *Menu) cursorAtLastChoice() {
	m.cursor = 0
	for i, c := range m.entries {
		if !c.Header {
			m.cursor = i
		}
	}
}

// Draw implements gruid.Model.Draw.
func (m *Menu) Draw() gruid.Grid {
	if m.style.Boxed {
		b := box{
			grid:       m.grid,
			title:      m.title,
			style:      m.style.Box,
			titleStyle: m.style.Title,
		}
		b.draw()
	}
	alt := false
	rg := m.grid.Range().Relative()
	cgrid := m.grid.Slice(rg.Shift(1, 1, -1, -1))
	if !m.style.Boxed {
		cgrid = m.grid
	}
	crg := cgrid.Range().Relative()
	for i, c := range m.entries {
		if !c.Header {
			st := m.stt.Style()
			if alt {
				st.Bg = m.style.BgAlt
			}
			alt = !alt
			if i == m.cursor {
				st.Fg = m.style.Selected
			}
			nchars := utf8.RuneCountInString(c.Text)
			m.stt.With(c.Text, st).Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Iter(func(pos gruid.Position) {
				line.SetCell(pos, cell)
			})
		} else {
			alt = false
			st := m.style.Header
			nchars := utf8.RuneCountInString(c.Text)
			m.stt.With(c.Text, st).Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Iter(func(pos gruid.Position) {
				line.SetCell(pos, cell)
			})
		}
	}
	return m.grid
}
