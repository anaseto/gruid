package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// MenuEntry represents an entry in the menu. By default they behave much like
// a button and can be activated and invoked.
type MenuEntry struct {
	// Text is the text displayed on the entry line.
	Text string

	// Disabled means that the entry is not activable nor invokable. It may
	// represent a header or an unavailable choice, for example.
	Disabled bool

	// Keys contains entry shortcuts, if any, and only for activable
	// entries. Other menu key bindings take precedence over those.
	Keys []gruid.Key
}

// MenuKeys contains key bindings configuration for the menu.
type MenuKeys struct {
	Up     []gruid.Key // move up active entry (default: ArrowUp, k)
	Down   []gruid.Key // move down active entry (default: ArrowDown, j)
	Invoke []gruid.Key // invoke selection (default: Enter)
	Quit   []gruid.Key // requist menu quit (default: Escape, q, Q)
}

// MenuAction represents an user action with the menu.
type MenuAction int

// These constants represent the available actions in a menu.
const (
	// MenuPass reports that the menu state did not change (for example a
	// mouse motion outside the menu, or within a same entry line).
	MenuPass MenuAction = iota

	// MenuMove reports that the user moved the active entry. This happens
	// by default when using the arrow keys or a mouse motion.
	MenuMove

	// MenuInvoke reports that the user clicked or pressed enter to invoke
	// an already active entry, or used a shortcut key to both activate and
	// invoke a specific entry.
	MenuInvoke

	// MenuQuit reports that the user requested to quit the menu, either by
	// clicking outside the menu, or by using a key shortcut.
	MenuQuit
)

// MenuStyle describes styling options for a menu.
type MenuStyle struct {
	BgAlt    gruid.Color // alternate background on even choice lines
	Selected gruid.Color // foreground for selected entry
	Disabled gruid.Style // disabled entry style
}

// MenuConfig contains configuration options for creating a menu.
type MenuConfig struct {
	Grid       gruid.Grid  // grid slice where the menu is drawn
	Entries    []MenuEntry // menu entries
	StyledText StyledText  // default styled text formatter for content
	Keys       MenuKeys    // optional custom key bindings
	Box        *Box        // draw optional box around the menu
	Style      MenuStyle
}

// NewMenu returns a menu with a given configuration.
func NewMenu(cfg MenuConfig) *Menu {
	m := &Menu{
		grid:    cfg.Grid,
		entries: cfg.Entries,
		box:     cfg.Box,
		stt:     cfg.StyledText,
		style:   cfg.Style,
		keys:    cfg.Keys,
		draw:    true,
	}
	if m.keys.Invoke == nil {
		m.keys.Invoke = []gruid.Key{gruid.KeyEnter}
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

// Menu is a widget that displays a list of entries to the user. It allows to
// move the active entry, as well as invoke a particular entry.
type Menu struct {
	grid    gruid.Grid
	entries []MenuEntry
	box     *Box
	stt     StyledText
	style   MenuStyle
	cursor  int
	action  MenuAction
	draw    bool
	init    bool // Update received MsgInit
	keys    MenuKeys
}

// Active return the index of the currently active entry.
func (m *Menu) Active() int {
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
			for m.cursor < l && m.entries[m.cursor].Disabled {
				m.cursor++
			}
			if m.cursor >= l {
				m.cursorAtFirstChoice()
			}
		case msg.Key.In(m.keys.Up):
			m.action = MenuMove
			m.cursor--
			for m.cursor >= 0 && m.entries[m.cursor].Disabled {
				m.cursor--
			}
			if m.cursor < 0 {
				m.cursor = 0
				m.cursorAtLastChoice()
			}
		case msg.Key.In(m.keys.Invoke) && m.cursor < l && !m.entries[m.cursor].Disabled:
			m.action = MenuInvoke
		default:
			nchars := utf8.RuneCountInString(string(msg.Key))
			if nchars != 1 {
				break
			}
			for i, e := range m.entries {
				for _, k := range e.Keys {
					if k == msg.Key {
						m.cursor = i
						m.action = MenuInvoke
						break
					}
				}
			}
		}
	case gruid.MsgMouse:
		crg := rg // content range
		if m.box != nil {
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
			if !msg.P.In(rg) || m.box == nil && p.Y >= len(m.entries) {
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
			if !m.entries[m.cursor].Disabled {
				m.action = MenuInvoke
			}
		}
	}
	return nil
}

func (m *Menu) drawGrid() gruid.Grid {
	h := len(m.entries) // menu content height
	if m.box != nil {
		h += 2 // borders height
	}
	max := m.grid.Size()
	return m.grid.Slice(gruid.NewRange(0, 0, max.X, h))
}

func (m *Menu) cursorAtFirstChoice() {
	m.cursor = 0
	for i, c := range m.entries {
		if !c.Disabled {
			m.cursor = i
			break
		}
	}
}

func (m *Menu) cursorAtLastChoice() {
	m.cursor = 0
	for i, c := range m.entries {
		if !c.Disabled {
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

	if m.box != nil {
		m.box.Draw(grid)
	}
	alt := false
	rg := grid.Range().Origin()
	cgrid := grid
	if m.box != nil {
		cgrid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	crg := cgrid.Range().Origin()
	for i, c := range m.entries {
		if !c.Disabled {
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
			st := m.style.Disabled
			nchars := utf8.RuneCountInString(c.Text)
			m.stt.With(c.Text, st).Draw(cgrid.Slice(crg.Line(i)))
			cell := gruid.Cell{Rune: ' ', Style: st}
			line := cgrid.Slice(crg.Line(i).Shift(nchars, 0, 0, 0))
			line.Fill(cell)
		}
	}
	return grid
}
