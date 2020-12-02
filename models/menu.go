package models

import (
	"strings"
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
	}
	m.cursorAtFirstChoice()
	return m
}

// Menu is a widget that asks the user to select an option among a list of
// entries.
type Menu struct {
	grid             gruid.Grid
	entries          []MenuEntry
	title            string
	style            MenuStyle
	cursor           int
	action           MenuAction
	nodraw           bool
	colorBg          gruid.Color
	colorBgAlt       gruid.Color
	colorFg          gruid.Color
	colorAvailable   gruid.Color
	colorSelected    gruid.Color
	colorUnavailable gruid.Color
	colorHeader      gruid.Color
	colorTitle       gruid.Color
}

//func (m *Menu) Init() {
//if ui.choiceCursor >= 0 && ui.choiceCursor < len(m.entries) &&
//m.entries[ui.choiceCursor].Kind == EntryChoice {
//m.cursor = ui.choiceCursor
//} else {
//m.cursorAtFirstChoice()
//}
//}

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
	m.nodraw = false
	m.action = MenuSelection // default action
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch {
		case msg.Key == gruid.KeyEscape || msg.Key == gruid.KeySpace || msg.Key == "x" || msg.Key == "X":
			m.action = MenuCancel
		case len(msg.Key) == 1:
		case msg.Key == gruid.KeyArrowDown:
			m.cursor++
			for m.cursor < l && m.entries[m.cursor].Kind != EntryChoice {
				m.cursor++
			}
			if m.cursor >= l {
				m.cursorAtFirstChoice()
			}
		case msg.Key == gruid.KeyArrowUp:
			m.cursor--
			for m.cursor >= 0 && m.entries[m.cursor].Kind != EntryChoice {
				m.cursor--
			}
			if m.cursor < 0 {
				m.cursorAtLastChoice()
			}
		case msg.Key == gruid.KeyEnter && m.entries[m.cursor].Kind == EntryChoice:
			m.action = MenuAccept
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
		}
	case gruid.MsgMouseMove:
		pos := m.grid.Range().Relative(msg.MousePos)
		if !m.grid.Contains(pos) {
			m.nodraw = true
			break
		}
		if m.isEdgePos(pos) {
			m.nodraw = true
			break
		}
		if pos.Y == m.cursor {
			m.nodraw = true
			break
		}
		m.cursor = pos.Y
	case gruid.MsgMouseDown:
		pos := m.grid.Range().Relative(msg.MousePos)
		switch msg.Button {
		case gruid.ButtonMain:
			if !m.grid.Contains(pos) {
				m.action = MenuCancel
				break
			}
			if m.isEdgePos(pos) {
				m.nodraw = true
				break
			}
			m.cursor = pos.Y
			if m.entries[m.cursor].Kind == EntryChoice {
				m.action = MenuAccept
			}
		case gruid.ButtonAuxiliary:
			m.action = MenuCancel
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
	j := 0
	alt := false
	//ui.DrawBox(m.title, m.x0-1, m.y0-1, m.w+2, len(m.entries)+2, m.colorTitle, m.colorBg)
	for i, c := range m.entries {
		bg := m.colorBg
		fg := m.colorFg
		if c.Kind != EntryHeader {
			if alt {
				bg = m.colorBgAlt
			}
			alt = !alt
			if c.Kind == EntryUnavailable {
				fg = m.colorUnavailable
			}
			if c.Kind == EntryChoice && i == m.cursor {
				fg = m.colorSelected
			}
			//ui.ClearWithColor(m.x0, m.y0+i, m.w, bg)
			//if m.letters {
			//ui.DrawColoredTextOnBG(fmt.Sprintf("%c - %s", rune(j+97), c.Text), m.x0, m.y0+i, fg, bg)
			//} else {
			//ui.DrawColoredTextOnBG(fmt.Sprintf(" %s", c.Text), m.x0, m.y0+i, fg, bg)
			//}
			j++
		} else {
			alt = false
			fg = m.colorHeader
			//ui.ClearWithColor(m.x0, m.y0+i, m.w, bg)
			//ui.DrawColoredTextOnBG(c.Text, m.x0, m.y0+i, fg, bg)
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
	rg := b.grid.Range()
	cgrid := b.grid.Slice(rg.Shift(1, 0, -1, 0))
	crg := cgrid.Range()
	t := textline{
		fg:   b.fg,
		bg:   b.bg,
		grid: cgrid.Slice(crg.Line(0)),
	}
	if b.title != "" {
		nchars := utf8.RuneCountInString(b.title)
		dist := (crg.Width() - nchars) / 2
		s := strings.Repeat("─", dist)
		t.text = s
		t.draw()
		t.fg = b.fgtitle
		t.grid = cgrid.Slice(crg.Shift(dist, 0, 0, 0))
	}
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

//func (t textline) draw() {
//nchars := utf8.RuneCountInString(text)
//if t.header {
//dist := (width - nchars) / 2
//for i := 0; i < dist; i++ {
//grid.SetCell(i, y, '─', ColorFg, bg)
//}
//if text != "" {
//ui.DrawColoredTextOnBG(text, x+dist, y, fg, bg)
//}
//for i := x + dist + nchars; i < x+width; i++ {
//grid.SetCell(i, y, '─', ColorFg, bg)
//}
//}

//func (ui *gameui) DrawBox(title string, x, y, width, height int, fg, bg Color) {
//ui.DrawHeader(title, x, y, width, fg, bg)
//ui.SetCell(x, y, '┌', ColorFg, bg)
//ui.SetCell(x+width-1, y, '┐', ColorFg, bg)
//ui.SetCell(x, y+height-1, '└', ColorFg, bg)
//ui.SetCell(x+width-1, y+height-1, '┘', ColorFg, bg)
//for j := y + 1; j < y+height-1; j++ {
//ui.SetCell(x, j, '│', ColorFg, bg)
//ui.SetCell(x+width-1, j, '│', ColorFg, bg)
//}
//for i := x + 1; i < x+width-1; i++ {
//ui.SetCell(i, y+height-1, '─', ColorFg, bg)
//}
//for i := x + 1; i < x+width-1; i++ {
//for j := y + 1; j < y+height-1; j++ {
//ui.SetCell(i, j, ' ', ColorFg, bg)
//}
//}
//}

//func (ui *gameui) DrawHeader(title string, x, y, width int, fg, bg Color) {
//nchars := utf8.RuneCountInString(title)
//dist := (width - nchars) / 2
//for i := x; i < x+dist; i++ {
//ui.SetCell(i, y, '─', ColorFg, bg)
//}
//if title != "" {
//ui.DrawColoredTextOnBG(title, x+dist, y, fg, bg)
//}
//for i := x + dist + nchars; i < x+width; i++ {
//ui.SetCell(i, y, '─', ColorFg, bg)
//}
//}

//func (ui *gameui) DrawColoredTextOnBG(text string, x, y int, fg, bg Color) {
//col := 0
//for _, r := range text {
//if r == '\n' {
//y++
//col = 0
//continue
//}
//if x+col >= UIWidth {
//break
//}
//ui.SetCell(x+col, y, r, fg, bg)
//col++
//}
//}
