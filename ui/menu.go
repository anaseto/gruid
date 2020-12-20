package ui

import (
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
	Left   []gruid.Key // move left active entry (default: ArrowLeft, h)
	Right  []gruid.Key // move right active entry (default: ArrowRight, l)
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
	Layout   gruid.Point // menu layout in (columns, lines); 0 means any
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
	if m.keys.Left == nil {
		m.keys.Left = []gruid.Key{gruid.KeyArrowLeft, "h"}
	}
	if m.keys.Right == nil {
		m.keys.Right = []gruid.Key{gruid.KeyArrowRight, "l"}
	}
	if m.keys.Quit == nil {
		m.keys.Quit = []gruid.Key{gruid.KeyEscape, "q", "Q"}
	}
	m.computeItems()
	m.cursorAtFirstChoice()
	return m
}

// item represents a visible entry in the menu at a given position and with a
// given slice.
type item struct {
	grid gruid.Grid  // its grid slice (may be empty)
	i    int         // index of corresponding entry in menu entries
	alt  bool        // even position (alternate background)
	page gruid.Point // page number (x,y)
}

// Menu is a widget that displays a list of entries to the user. It allows to
// move the active entry, as well as invoke a particular entry.
type Menu struct {
	grid    gruid.Grid
	entries []MenuEntry
	view    map[gruid.Point]item
	size    gruid.Point // view size (w, h) in cells
	box     *Box
	stt     StyledText
	style   MenuStyle
	active  gruid.Point
	action  MenuAction
	draw    bool
	init    bool // Update received MsgInit
	keys    MenuKeys
	layout  gruid.Point // current menu layout
}

// Active return the index of the currently active entry.
func (m *Menu) Active() int {
	return m.view[m.active].i
}

// Action returns the last action performed in the menu.
func (m *Menu) Action() MenuAction {
	return m.action
}

// SetEntries updates the list of menu entries.
func (m *Menu) SetEntries(entries []MenuEntry) {
	m.entries = entries
	m.computeItems()
	if !m.contains(m.active) {
		m.cursorAtLastChoice()
	}
}

func (m *Menu) contains(p gruid.Point) bool {
	_, ok := m.view[p]
	return ok
}

// SetActive updates the active entry among entries. It may be used
// to launch the menu at a specific default starting index.
func (m *Menu) SetActive(i int) {
	if i < 0 || i >= len(m.entries) {
		return
	}
	if !m.entries[i].Disabled {
		m.active = m.idxToPos(i)
	}
}

func (m *Menu) idxToPos(i int) gruid.Point {
	for p, it := range m.view {
		if it.i == i {
			return p
		}
	}
	return gruid.Point{}
}

func (m *Menu) moveTo(p gruid.Point) {
	// TODO: handle more intuitively some cases, and add support for
	// advancing pages, and show page number somewhere if possible, or at
	// least that there is more.
	q := m.active
	for {
		q = q.Add(p)
		it, ok := m.view[q]
		if !ok {
			break
		}
		if !m.entries[it.i].Disabled {
			break
		}
	}
	if m.contains(q) {
		m.action = MenuMove
		m.active = q
	}
}

// Update implements gruid.Model.Update and updates the menu state in response to
// user input messages.
func (m *Menu) Update(msg gruid.Msg) gruid.Effect {
	grid := m.drawGrid()
	rg := grid.Range()

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
			m.moveTo(gruid.Point{0, 1})
		case msg.Key.In(m.keys.Up):
			m.moveTo(gruid.Point{0, -1})
		case msg.Key.In(m.keys.Left):
			m.moveTo(gruid.Point{-1, 0})
		case msg.Key.In(m.keys.Right):
			m.moveTo(gruid.Point{1, 0})
		case msg.Key.In(m.keys.Invoke) && m.contains(m.active):
			it, ok := m.view[m.active]
			if ok && !m.entries[it.i].Disabled {
				m.action = MenuInvoke
			}
		default:
			for i, e := range m.entries {
				for _, k := range e.Keys {
					if k == msg.Key {
						m.active = m.idxToPos(i)
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
		p := msg.P
		page := gruid.Point{}
		if it, ok := m.view[m.active]; ok {
			page = it.page
		}
		switch msg.Action {
		case gruid.MouseMove:
			if !p.In(crg) {
				break
			}
			for q, it := range m.view {
				if it.page == page && p.In(it.grid.Range()) {
					if q == m.active {
						break
					}
					m.active = q
					m.action = MenuMove
				}
			}
		case gruid.MouseMain:
			if !p.In(rg) {
				m.action = MenuQuit
				if m.init {
					return gruid.End()
				}
				break
			}
			if !p.In(crg) {
				break
			}
			for q, it := range m.view {
				if it.page == page && p.In(it.grid.Range()) {
					m.active = q
					if m.entries[it.i].Disabled {
						m.action = MenuMove
					} else {
						m.action = MenuInvoke
					}
				}
			}
		}
	}
	return nil
}

func (m *Menu) drawGrid() gruid.Grid {
	h := len(m.entries) // menu content height
	layout := m.layout
	if layout.Y > 0 {
		h = layout.Y
	}
	if m.box != nil {
		h += 2 // borders height
	}
	max := m.grid.Size()
	return m.grid.Slice(gruid.NewRange(0, 0, max.X, h))
}

type mlayout int

const (
	table mlayout = iota
	column
	line
)

func (m *Menu) computeItems() {
	m.layout = m.style.Layout
	if m.layout.Y > m.grid.Size().Y {
		m.layout.Y = m.grid.Size().Y
	}
	if m.layout.Y > len(m.entries) {
		m.layout.Y = len(m.entries)
	}
	if m.layout.X > len(m.entries) {
		m.layout.X = len(m.entries)
	}
	grid := m.drawGrid()
	rg := grid.Range()
	if m.box != nil {
		grid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	m.size = grid.Size()
	w, h := m.size.X, m.size.Y
	layout := m.layout
	lines := layout.Y
	if lines <= 0 {
		lines = len(m.entries)
	}
	columns := layout.X
	if columns <= 0 {
		if lines == len(m.entries) {
			columns = 1
		} else {
			columns = len(m.entries)
		}
	}
	if lines*columns > len(m.entries) {
		columns = len(m.entries) / lines
	}
	var ml mlayout
	if columns > 1 && lines > 1 {
		ml = table
		w = w / columns
	} else if columns > 1 {
		ml = line
	} else {
		ml = column
	}
	if m.view == nil {
		m.view = make(map[gruid.Point]item)
	} else {
		for k := range m.view {
			delete(m.view, k)
		}
	}
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	switch ml {
	case column:
		alt := true
		for i, e := range m.entries {
			if e.Disabled {
				alt = false
			}
			p := gruid.Point{0, i}
			m.view[p] = item{
				grid: grid.Slice(gruid.NewRange(0, i%h, w, (i%h)+1)),
				i:    i,
				alt:  alt,
				page: gruid.Point{0, i / h},
			}
			if !e.Disabled {
				alt = !alt
			}
		}
	case line:
		var to, hpage int
		alt := true
		for i, e := range m.entries {
			if e.Disabled {
				alt = false
			}
			from := to
			tw := m.stt.WithText(e.Text).Size().X
			to += tw
			if from > 0 && to > w {
				from = 0
				to = tw
				hpage++
			}
			p := gruid.Point{i, 0}
			m.view[p] = item{
				grid: grid.Slice(gruid.NewRange(from, 0, to, 1)),
				i:    i,
				page: gruid.Point{hpage, 0},
				alt:  alt,
			}
			if !e.Disabled {
				alt = !alt
			}
		}
	case table:
		for i, _ := range m.entries {
			page := i / (columns * h)
			pageidx := i % (columns * h)
			ln := pageidx % h
			col := pageidx / h
			p := gruid.Point{col, ln + page*h}
			m.view[p] = item{
				grid: grid.Slice(gruid.NewRange(col*w, ln%h, (col+1)*w, (ln%h)+1)),
				i:    i,
				page: gruid.Point{0, page},
				alt:  (col+ln)%h == 0,
			}
		}
	}
}

func (m *Menu) cursorAtFirstChoice() {
	j := 0
	for i, c := range m.entries {
		if !c.Disabled {
			j = i
			break
		}
	}
	m.active = m.idxToPos(j)
}

func (m *Menu) cursorAtLastChoice() {
	j := len(m.entries) - 1
	for i, c := range m.entries {
		if !c.Disabled {
			j = i
		}
	}
	m.active = m.idxToPos(j)
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
	activeItem := m.view[m.active]
	for p, it := range m.view {
		if it.page != activeItem.page {
			continue
		}
		i := it.i
		c := m.entries[i]
		if !c.Disabled {
			st := m.stt.Style()
			if it.alt {
				st.Bg = m.style.BgAlt
			}
			if p == m.active {
				st.Fg = m.style.Selected
			}
			cell := gruid.Cell{Rune: ' ', Style: st}
			it.grid.Fill(cell)
			m.stt.With(c.Text, st).Draw(it.grid)
		} else {
			st := m.style.Disabled
			cell := gruid.Cell{Rune: ' ', Style: st}
			it.grid.Fill(cell)
			m.stt.With(c.Text, st).Draw(it.grid)
		}
	}
	return grid
}
