package ui

import (
	"fmt"

	"github.com/anaseto/gruid"
)

// MenuConfig contains configuration options for creating a menu.
type MenuConfig struct {
	Grid    gruid.Grid  // grid slice where the menu is drawn
	Entries []MenuEntry // menu entries
	Keys    MenuKeys    // optional custom key bindings
	Box     *Box        // draw optional box around the menu
	Style   MenuStyle
}

// MenuEntry represents an entry in the menu. By default they behave much like
// a button and can be activated and invoked.
type MenuEntry struct {
	// Text is the styled text displayed on the entry line.
	Text StyledText

	// Disabled means that the entry is not invokable. It may represent a
	// header or an unavailable choice, for example.
	Disabled bool

	// Keys contains entry shortcuts, if any, and only for activable
	// entries. Other menu key bindings take precedence over those.
	Keys []gruid.Key
}

// MenuKeys contains key bindings configuration for the menu. One step movement
// keys skip disabled entries.
type MenuKeys struct {
	Up       []gruid.Key // move up active entry (default: ArrowUp, k)
	Down     []gruid.Key // move down active entry (default: ArrowDown, j)
	Left     []gruid.Key // move left active entry (default: ArrowLeft, h)
	Right    []gruid.Key // move right active entry (default: ArrowRight, l)
	PageDown []gruid.Key // go one page down (default: PageDown)
	PageUp   []gruid.Key // go one page up (default: PageUp)
	Invoke   []gruid.Key // invoke selection (default: Enter)
	Quit     []gruid.Key // requist menu quit (default: Escape, q, Q)
}

// MenuStyle describes styling options for a menu.
type MenuStyle struct {
	Layout  gruid.Point // menu layout in (columns, lines); 0 means any
	Active  gruid.Style // specific styling for active entry (no change if default)
	PageNum gruid.Style // page num display style (for boxed menu)
}

// Menu is a widget that displays a list of entries to the user. It allows to
// move the active entry, as well as invoke a particular entry.
//
// Menu implements gruid.Model, but is not suitable for use as main model of an
// application.
type Menu struct {
	grid    gruid.Grid
	entries []MenuEntry
	table   map[gruid.Point]item
	points  []gruid.Point
	pages   gruid.Point
	size    gruid.Point // view size (w, h) in cells
	box     *Box
	style   MenuStyle
	active  gruid.Point
	action  MenuAction
	keys    MenuKeys
	layout  gruid.Point // current menu layout
	dirty   bool        // state changed in Update and Draw was still not called
	drawn   gruid.Grid  // last grid slice that was drawn
}

// item represents a visible entry in the menu at a given position and with a
// given slice.
type item struct {
	grid gruid.Grid  // its grid slice (may be empty)
	i    int         // index of corresponding entry in menu entries
	page gruid.Point // page number (x,y)
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

// NewMenu returns a menu with a given configuration.
func NewMenu(cfg MenuConfig) *Menu {
	m := &Menu{
		grid:    cfg.Grid,
		entries: cfg.Entries,
		box:     cfg.Box,
		style:   cfg.Style,
		keys:    cfg.Keys,
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
	if m.keys.PageDown == nil {
		m.keys.PageDown = []gruid.Key{gruid.KeyPageDown}
	}
	if m.keys.PageUp == nil {
		m.keys.PageUp = []gruid.Key{gruid.KeyPageUp}
	}
	if m.keys.Quit == nil {
		m.keys.Quit = []gruid.Key{gruid.KeyEscape, "q", "Q"}
	}
	m.placeItems()
	m.cursorAtFirstChoice()
	m.dirty = true
	return m
}

// Active return the index of the currently active entry.
func (m *Menu) Active() int {
	return m.table[m.active].i
}

// Action returns the last action performed in the menu.
func (m *Menu) Action() MenuAction {
	return m.action
}

// SetEntries updates the list of menu entries.
func (m *Menu) SetEntries(entries []MenuEntry) {
	m.entries = entries
	m.placeItems()
	if !m.contains(m.active) {
		m.cursorAtLastChoice()
	}
	m.dirty = true
}

// SetBox updates the menu surrounding box.
func (m *Menu) SetBox(b *Box) {
	m.box = b
	m.placeItems()
	m.dirty = true
}

func (m *Menu) contains(p gruid.Point) bool {
	_, ok := m.table[p]
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
	m.dirty = true
}

func (m *Menu) idxToPos(i int) gruid.Point {
	if i >= 0 && i < len(m.points) {
		return m.points[i]
	}
	return gruid.Point{}
}

func (m *Menu) moveTo(p gruid.Point) {
	oactive := m.active
	q := m.active
	for {
		q = q.Add(p)
		it, ok := m.table[q]
		if !ok {
			break
		}
		if !m.entries[it.i].Disabled {
			break
		}
	}
	if m.contains(q) {
		m.active = q
	} else if q, ok := m.nextPage(p); ok {
		m.active = q
	} else {
		switch p {
		case gruid.Point{0, 1}, gruid.Point{1, 0}:
			m.cursorAtFirstChoice()
		case gruid.Point{0, -1}, gruid.Point{-1, 0}:
			m.cursorAtLastChoice()
		}
	}
	if m.active != oactive {
		m.action = MenuMove
	}
}

func (m *Menu) nextPage(p gruid.Point) (gruid.Point, bool) {
	it, ok := m.table[m.active]
	if !ok {
		return gruid.Point{}, false
	}
	for i := it.i + 1; i < len(m.entries); i++ {
		q := m.idxToPos(i)
		switch p {
		case gruid.Point{0, 1}:
			if m.table[q].page.Y > it.page.Y {
				return q, true
			}
		case gruid.Point{1, 0}:
			if m.table[q].page.X > it.page.X {
				return q, true
			}
		}
	}
	for i := it.i - 1; i >= 0; i-- {
		q := m.idxToPos(i)
		switch p {
		case gruid.Point{0, -1}:
			if m.table[q].page.Y < it.page.Y {
				return q, true
			}
		case gruid.Point{-1, 0}:
			if m.table[q].page.X < it.page.X {
				return q, true
			}
		}
	}
	return gruid.Point{}, false
}

// Update implements gruid.Model.Update and updates the menu state in response to
// user input messages. It considers mouse message coordinates to be absolute in
// its grid.
func (m *Menu) Update(msg gruid.Msg) gruid.Effect {
	m.action = MenuPass
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		m.updateKeyDown(msg)
	case gruid.MsgMouse:
		m.updateMouse(msg)
	}
	if m.Action() != MenuPass {
		m.dirty = true
	}
	return nil
}

func (m *Menu) pageDown() {
	p := gruid.Point{0, 1}
	if m.pages.Y == 0 {
		p = gruid.Point{1, 0}
	}
	if p, ok := m.nextPage(p); ok {
		m.action = MenuMove
		m.active = p
	}
}

func (m *Menu) pageUp() {
	p := gruid.Point{0, -1}
	if m.pages.Y == 0 {
		p = gruid.Point{-1, 0}
	}
	if p, ok := m.nextPage(p); ok {
		m.action = MenuMove
		m.active = p
	}
}

func (m *Menu) keyInvoke(key gruid.Key) {
	for i, e := range m.entries {
		for _, k := range e.Keys {
			if k == key {
				m.active = m.idxToPos(i)
				m.action = MenuInvoke
				break
			}
		}
	}
}

func (m *Menu) updateKeyDown(msg gruid.MsgKeyDown) {
	switch {
	case msg.Key.In(m.keys.Quit):
		m.action = MenuQuit
	case msg.Key.In(m.keys.Down):
		m.moveTo(gruid.Point{0, 1})
	case msg.Key.In(m.keys.Up):
		m.moveTo(gruid.Point{0, -1})
	case msg.Key.In(m.keys.Right):
		m.moveTo(gruid.Point{1, 0})
	case msg.Key.In(m.keys.Left):
		m.moveTo(gruid.Point{-1, 0})
	case msg.Key.In(m.keys.PageDown):
		m.pageDown()
	case msg.Key.In(m.keys.PageUp):
		m.pageUp()
	case msg.Key.In(m.keys.Invoke) && m.contains(m.active):
		it, ok := m.table[m.active]
		if ok && !m.entries[it.i].Disabled {
			m.action = MenuInvoke
		}
	default:
		m.keyInvoke(msg.Key)
	}
}

func (m *Menu) updateMouse(msg gruid.MsgMouse) {
	grid := m.pageGrid()
	rg := grid.Bounds()
	crg := rg // content range
	if m.box != nil {
		crg = crg.Shift(1, 1, -1, -1)
	}
	p := msg.P
	switch msg.Action {
	case gruid.MouseMove:
		if !p.In(crg) {
			break
		}
		m.moveToPoint(p)
	case gruid.MouseWheelDown:
		if !p.In(crg) {
			break
		}
		m.pageDown()
	case gruid.MouseWheelUp:
		if !p.In(crg) {
			break
		}
		m.pageUp()
	case gruid.MouseMain:
		if !p.In(rg) {
			m.action = MenuQuit
			break
		}
		if !p.In(crg) {
			break
		}
		m.invokePoint(p)
	}
}

func (m *Menu) moveToPoint(p gruid.Point) {
	page := gruid.Point{}
	if it, ok := m.table[m.active]; ok {
		page = it.page
	}
	for q, it := range m.table {
		if it.page == page && p.In(it.grid.Bounds()) {
			if q == m.active {
				break
			}
			m.active = q
			m.action = MenuMove
		}
	}
}

func (m *Menu) invokePoint(p gruid.Point) {
	page := gruid.Point{}
	if it, ok := m.table[m.active]; ok {
		page = it.page
	}
	for q, it := range m.table {
		if it.page == page && p.In(it.grid.Bounds()) {
			m.active = q
			if m.entries[it.i].Disabled {
				m.action = MenuMove
			} else {
				m.action = MenuInvoke
			}
		}
	}
}

func (m *Menu) pageGrid() gruid.Grid {
	if m.layout.Y > 0 && m.layout.X == 0 {
		return m.drawGrid()
	}
	h := 0
	for p, it := range m.table {
		if p.X > 0 {
			continue
		}
		activeItem := m.table[m.active]
		if it.page != activeItem.page {
			continue
		}
		h++
	}
	if m.box != nil {
		h += 2 // borders height
	}
	max := m.grid.Size()
	return m.grid.Slice(gruid.NewRange(0, 0, max.X, h))
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

func (m *Menu) updateLayout() {
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
}

func (m *Menu) getLayout(w, h int) (ml mlayout, nw, columns int) {
	lines := m.layout.Y
	nw = w
	if lines <= 0 {
		lines = len(m.entries)
	}
	columns = m.layout.X
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
	if columns > 1 && lines > 1 {
		ml = table
		nw = nw / columns
	} else if columns > 1 {
		ml = line
	} else {
		ml = column
	}
	return ml, nw, columns
}

func (m *Menu) resetPositions() {
	if m.table == nil {
		m.table = make(map[gruid.Point]item)
	} else {
		for k := range m.table {
			delete(m.table, k)
		}
	}
	m.points = m.points[:0]
}

func (m *Menu) placeItems() {
	m.updateLayout()
	grid := m.drawGrid()
	rg := grid.Bounds()
	if m.box != nil {
		grid = grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	m.size = grid.Size()
	w, h := m.size.X, m.size.Y
	ml, w, columns := m.getLayout(w, h)
	m.resetPositions()
	if w <= 0 {
		w = 1
	}
	if h <= 0 {
		h = 1
	}
	switch ml {
	case column:
		m.columnArrangement(grid, w, h)
	case line:
		m.lineArrangement(grid, w)
	case table:
		m.tableArrangement(grid, w, h, columns)
	}
	m.updatePages()
}

func (m *Menu) columnArrangement(grid gruid.Grid, w, h int) {
	for i := range m.entries {
		p := gruid.Point{0, i}
		m.table[p] = item{
			grid: grid.Slice(gruid.NewRange(0, i%h, w, (i%h)+1)),
			i:    i,
			page: gruid.Point{0, i / h},
		}
		m.points = append(m.points, p)
	}
}

func (m *Menu) lineArrangement(grid gruid.Grid, w int) {
	var to, hpage int
	for i, e := range m.entries {
		from := to
		tw := e.Text.Size().X
		to += tw
		if from > 0 && to > w {
			from = 0
			to = tw
			hpage++
		}
		p := gruid.Point{i, 0}
		m.table[p] = item{
			grid: grid.Slice(gruid.NewRange(from, 0, to, 1)),
			i:    i,
			page: gruid.Point{hpage, 0},
		}
		m.points = append(m.points, p)
	}
}

func (m *Menu) tableArrangement(grid gruid.Grid, w, h, columns int) {
	for i := range m.entries {
		page := i / (columns * h)
		pageidx := i % (columns * h)
		ln := pageidx % h
		col := pageidx / h
		p := gruid.Point{col, ln + page*h}
		m.table[p] = item{
			grid: grid.Slice(gruid.NewRange(col*w, ln%h, (col+1)*w, (ln%h)+1)),
			i:    i,
			page: gruid.Point{0, page},
		}
		m.points = append(m.points, p)
	}
}

func (m *Menu) updatePages() {
	for _, p := range m.points {
		pg := m.table[p].page
		if pg.X > m.pages.X {
			m.pages.X = pg.X
		}
		if pg.Y > m.pages.Y {
			m.pages.Y = pg.Y
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
	if !m.dirty {
		return m.drawn
	}
	grid := m.pageGrid()
	if m.box != nil {
		pg := m.table[m.active].page
		var lnumtext string
		if m.pages.X == 0 && m.pages.Y == 0 {
		} else if m.pages.X == 0 {
			lnumtext = fmt.Sprintf("%d/%d", pg.Y, m.pages.Y)
		} else if m.pages.Y == 0 {
			lnumtext = fmt.Sprintf("%d/%d", pg.X, m.pages.X)
		} else {
			lnumtext = fmt.Sprintf("%d,%d/%d,%d", pg.X, pg.Y, m.pages.X, m.pages.Y)
		}
		foot := m.box.Footer
		if foot.Text() == "" {
			m.box.Footer = NewStyledText(lnumtext, m.style.PageNum)
		}
		m.box.Draw(grid)
		m.box.Footer = foot
	}
	activeItem := m.table[m.active]
	for p, it := range m.table {
		if it.page != activeItem.page {
			continue
		}
		i := it.i
		c := m.entries[i]
		st := c.Text.Style()
		if !c.Disabled {
			if p == m.active {
				if m.style.Active.Fg != gruid.ColorDefault {
					st.Fg = m.style.Active.Fg
				}
				if m.style.Active.Bg != gruid.ColorDefault {
					st.Bg = m.style.Active.Bg
				}
				if m.style.Active.Attrs != gruid.AttrsDefault {
					st.Attrs = m.style.Active.Attrs
				}
			}
			cell := gruid.Cell{Rune: ' ', Style: st}
			it.grid.Fill(cell)
			c.Text.WithStyle(st).Draw(it.grid)
		} else {
			cell := gruid.Cell{Rune: ' ', Style: st}
			it.grid.Fill(cell)
			c.Text.Draw(it.grid)
		}
	}
	m.dirty = false
	m.drawn = grid
	return m.drawn
}
