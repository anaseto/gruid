package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// TextInputStyle describes styling options for a TextInput.
type TextInputStyle struct {
	Cursor gruid.Style // cursor style
	Prompt gruid.Style // prompt style, if any
}

// TextInputConfig describes configuration options for creating a text input.
type TextInputConfig struct {
	Grid       gruid.Grid    // grid slice where the text input is drawn
	StyledText StyledText    // styled text with initial text input text content
	Prompt     string        // optional prompt text
	Box        *Box          // draw optional box around the text input
	Keys       TextInputKeys // optional custom key bindings for the text input
	Style      TextInputStyle
}

// TextInputKeys contains key bindings configuration for the text input.
type TextInputKeys struct {
	Quit []gruid.Key // quit text input (default: Escape, Tab)
}

// TextInput represents a line entry with text supplied from the user that can
// be validated. It offers only basic editing shortcuts.
type TextInput struct {
	grid      gruid.Grid
	stt       StyledText
	box       *Box
	prompt    string
	content   []rune
	style     TextInputStyle
	cursorMin int
	cursor    int
	action    TextInputAction
	keys      TextInputKeys
	init      bool // received gruid.MsgInit
}

// TextInputAction represents last user action with the text input.
type TextInputAction int

// These constants represent possible actions araising from interaction with
// the text input.
const (
	TextInputPass     TextInputAction = iota // no change in state
	TextInputChange                          // changed content or moved cursor
	TextInputActivate                        // activate/accept content
	TextInputQuit                            // quit/cancel text input
)

// NewTextInput returns a new text input with givent configuration options.
func NewTextInput(cfg TextInputConfig) *TextInput {
	ti := &TextInput{
		grid:   cfg.Grid,
		stt:    cfg.StyledText,
		box:    cfg.Box,
		prompt: cfg.Prompt,
		style:  cfg.Style,
		keys:   cfg.Keys,
	}
	if ti.style.Cursor == ti.stt.Style() {
		// not true reverse with terminal driver, but good enough as a default
		ti.style.Cursor.Bg, ti.style.Cursor.Fg = ti.style.Cursor.Fg, ti.style.Cursor.Bg
	}
	ti.cursorMin = utf8.RuneCountInString(ti.prompt)
	ti.content = []rune(ti.stt.Text())
	ti.cursor = len(ti.content)
	if ti.keys.Quit == nil {
		ti.keys.Quit = []gruid.Key{gruid.KeyEscape, gruid.KeyTab}
	}
	return ti
}

// SetBox updates the text input surrounding box.
func (ti *TextInput) SetBox(b *Box) {
	ti.box = b
}

func (ti *TextInput) cursorMax() int {
	return len(ti.content)
}

// Update implements gruid.Model.Update for TextInput. It considers mouse
// message coordinates to be absolute in its grid.
func (ti *TextInput) Update(msg gruid.Msg) gruid.Effect {
	ti.action = TextInputPass
	switch msg := msg.(type) {
	case gruid.MsgInit:
		ti.init = true
	case gruid.MsgKeyDown:
		if msg.Key.In(ti.keys.Quit) {
			ti.action = TextInputQuit
			if ti.init {
				return gruid.End()
			}
			return nil
		}
		switch msg.Key {
		case gruid.KeyHome:
			if ti.cursor > 0 {
				ti.action = TextInputChange
				ti.cursor = 0
			}
		case gruid.KeyEnd:
			if ti.cursor < ti.cursorMax() {
				ti.action = TextInputChange
				ti.cursor = ti.cursorMax()
			}
		case gruid.KeyArrowLeft:
			if ti.cursor > 0 {
				ti.action = TextInputChange
				ti.cursor--
			}
		case gruid.KeyArrowRight:
			if ti.cursor < ti.cursorMax() {
				ti.action = TextInputChange
				ti.cursor++
			}
		case gruid.KeyBackspace:
			if ti.cursor > 0 {
				ti.content = append(ti.content[:ti.cursor-1], ti.content[ti.cursor:]...)
				ti.cursor--
				ti.action = TextInputChange
			}
		case gruid.KeyEnter:
			ti.action = TextInputActivate
		default:
			if !msg.Key.IsRune() {
				break
			}
			r, _ := utf8.DecodeRuneInString(string(msg.Key))
			var c []rune
			c = append(c, ti.content[:ti.cursor]...)
			c = append(c, r)
			c = append(c, ti.content[ti.cursor:]...)
			ti.content = c
			ti.cursor++
			ti.action = TextInputChange
		}
	case gruid.MsgMouse:
		cgrid := ti.grid
		if ti.box != nil {
			rg := ti.grid.Range().Origin()
			cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
		}
		start := ti.start()
		p := msg.P.Sub(cgrid.Range().Min)
		switch msg.Action {
		case gruid.MouseMain:
			if !msg.P.In(ti.grid.Range()) {
				ti.action = TextInputQuit
				if ti.init {
					return gruid.End()
				}
				return nil
			}
			if !cgrid.Contains(p) {
				break
			}
			ti.cursor = msg.P.X + start - ti.cursorMin - 1
			if ti.cursor > ti.cursorMax() {
				ti.cursor = ti.cursorMax()
			}
			if ti.cursor < 0 {
				ti.cursor = 0
			}
		}
	}
	return nil
}

// Content returns the current content of the text input.
func (ti *TextInput) Content() string {
	return string(ti.content)
}

// Action returns the action performed with the TextInput in the last call to
// Update.
func (ti *TextInput) Action() TextInputAction {
	return ti.action
}

func (ti *TextInput) cursorRune() rune {
	if ti.cursor < len(ti.content) {
		return ti.content[ti.cursor]
	}
	return ' '
}

func (ti *TextInput) start() int {
	cgrid := ti.grid
	if ti.box != nil {
		rg := ti.grid.Range().Origin()
		cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	crg := cgrid.Range().Origin()
	start := 0
	w := crg.Size().X
	w -= ti.cursorMin
	if ti.cursor >= w-1 {
		if w > 2 {
			start = ti.cursor - w + 2
		} else {
			start = ti.cursor - w
		}
	}
	return start
}

// Draw implements gruid.Model.Draw for TextInput.
func (ti *TextInput) Draw() gruid.Grid {
	cgrid := ti.grid
	if ti.box != nil {
		ti.box.Draw(ti.grid)
		rg := ti.grid.Range().Origin()
		cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	cgrid.Fill(gruid.Cell{Rune: ' ', Style: ti.stt.Style()})
	ti.stt.With(ti.prompt, ti.style.Prompt).Draw(cgrid)
	crg := cgrid.Range().Origin()
	start := ti.start()
	ti.stt.WithText(string(ti.content[start:])).Draw(cgrid.Slice(crg.Shift(ti.cursorMin, 0, 0, 0)))
	ti.stt.With(string(ti.cursorRune()), ti.style.Cursor).Draw(cgrid.Slice(crg.Shift(ti.cursorMin+ti.cursor-start, 0, 0, 0)))
	return ti.grid
}
