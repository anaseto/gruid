package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// MenuEntry represents an entry in the menu. It is displayed on one line, and
// for example can be a choice or a header.
type MenuEntry struct {
	Text   string // displayed text on the entry line
	Header bool   // not an activable entry, but a sub-header entry

	// Keys contains entry shortcuts, if any, and only for non-header
	// activable entries. Other menu key bindings take precedence over
	// those.
	Keys []gruid.Key
}

// MenuKeys contains key bindings configuration for the menu.
type MenuKeys struct {
	Up       []gruid.Key // move selection up
	Down     []gruid.Key // move selection down
	Activate []gruid.Key // activate selection
	Quit     []gruid.Key // quit menu
}

// MenuAction represents an user action with the menu.
type MenuAction int

// These constants represent the available actions in a menu.
const (
	// MenuPass reports that the menu state did not change (for example a
	// mouse motion outside the menu, or within a same entry line).
	MenuPass MenuAction = iota

	// MenuMove reports that the user moved the selection cursor by using
	// the arrow keys or a mouse motion.
	MenuMove

	// MenuActivate reports that the user clicked or pressed enter to
	// activate/accept a selected entry, or used a shortcut key to select
	// and activate/accept a specific entry.
	MenuActivate

	// MenuQuit reports that the user clicked outside the menu, or
	// pressed Esc, Space or X.
	MenuQuit
)

// MenuStyle describes styling options for a menu.
type MenuStyle struct {
	BgAlt    gruid.Color // alternate background on even choice lines
	Selected gruid.Color // foreground for selected entry
	Header   gruid.Style // header entry
	Boxed    bool        // draw a box around the menu
	Box      gruid.Style // box style, if any
	Title    gruid.Style // box title style, if any
}

// MenuConfig contains configuration options for creating a menu.
type MenuConfig struct {
	Grid       gruid.Grid  // grid slice where the menu is drawn
	Entries    []MenuEntry // menu entries
	StyledText StyledText  // default styled text formatter for content
	Title      string      // optional title, implies Boxed style
	Keys       MenuKeys    // optional custom key bindings
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
		keys:    cfg.Keys,
		draw:    true,
	}
	if m.title != "" {
		m.style.Boxed = true
	}
	if m.keys.Activate == nil {
		m.keys.Activate = []gruid.Key{gruid.KeyEnter}
	}
	if m.keys.Down == nil {
		m.keys.Down = []gruid.Key{gruid.KeyArrowDown, "j"}
	}
	if m.keys.Up == nil {
		m.keys.Up = []gruid.Key{gruid.KeyArrowUp, "k"}
	}
	if m.keys.Quit == nil {
		m.keys.Quit = []gruid.Key{gruid.KeyEscape, "q", "Q"}
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
	init    bool // Update received MsgInit
	keys    MenuKeys
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
func (m *Menu) Update(msg gruid.Msg) gruid.Effect {
	grid := m.drawGrid()
	rg := grid.Range()

	l := len(m.entries)
	switch msg := msg.(type) {
	case gruid.MsgInit:
		m.init = true
	case gruid.MsgKeyDown:
		switch {
		case msg.Key.In(m.keys.Quit):
			m.action = MenuQuit
			if m.init {
				return gruid.End()
			}
		case msg.Key.In(m.keys.Down):
			m.action = MenuMove
			m.cursor++
			for m.cursor < l && m.entries[m.cursor].Header {
				m.cursor++
			}
			if m.cursor >= l {
				m.cursorAtFirstChoice()
			}
		case msg.Key.In(m.keys.Up):
			m.action = MenuMove
			m.cursor--
			for m.cursor >= 0 && m.entries[m.cursor].Header {
				m.cursor--
			}
			if m.cursor < 0 {
				m.cursor = 0
				m.cursorAtLastChoice()
			}
		case msg.Key.In(m.keys.Activate) && m.cursor < l && !m.entries[m.cursor].Header:
			m.action = MenuActivate
		default:
			nchars := utf8.RuneCountInString(string(msg.Key))
			if nchars != 1 {
				break
			}
			for i, e := range m.entries {
				for _, k := range e.Keys {
					if k == msg.Key {
						m.cursor = i
						m.action = MenuActivate
						break
					}
				}
			}
		}
	case gruid.MsgMouse:
		crg := rg // content range
		if m.style.Boxed {
			crg = crg.Shift(1, 1, -1, -1)
		}
		p := msg.P.Rel(crg)
		switch msg.Action {
		case gruid.MouseMove:
			if !p.In(crg.Origin()) || p.Y == m.cursor || p.Y >= len(m.entries) {
				break
			}
			m.cursor = p.Y
			m.action = MenuMove
		case gruid.MouseMain:
			if !msg.P.In(rg) || !m.style.Boxed && p.Y >= len(m.entries) {
				m.action = MenuQuit
				if m.init {
					return gruid.End()
				}
				break
			}
			if !p.In(crg.Origin()) || p.Y >= len(m.entries) {
				break
			}
			m.cursor = p.Y
			m.action = MenuMove
			if !m.entries[m.cursor].Header {
				m.action = MenuActivate
			}
		}
	}
	return nil
}

func (m *Menu) drawGrid() gruid.Grid {
	h := len(m.entries) // menu content height
	if m.style.Boxed {
		h += 2 // borders height
	}
	max := m.grid.Size()
	return m.grid.Slice(gruid.NewRange(0, 0, max.X, h))
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

// Draw implements gruid.Model.Draw. It returns the grid slice that was drawn.
func (m *Menu) Draw() gruid.Grid {
	if m.init {
		m.grid.Fill(gruid.Cell{Rune: ' '})
	}
	grid := m.drawGrid()

	if m.style.Boxed {
		b := box{
			grid:  grid,
			title: m.stt.With(m.title, m.style.Title),
			style: m.style.Box,
		}
		b.draw()
	}
	alt := false
	rg := grid.Range().Origin()
	cgrid := grid
	if m.style.Boxed {
		cgrid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	crg := cgrid.Range().Origin()
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
			text := m.stt.With(c.Text, st)
			max := text.Size()
			m.stt.With(c.Text, st).Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(max.X, 0, 0, 0))
			line.Fill(cell)
		} else {
			alt = false
			st := m.style.Header
			nchars := utf8.RuneCountInString(c.Text)
			m.stt.With(c.Text, st).Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Fill(cell)
		}
	}
	return grid
}
