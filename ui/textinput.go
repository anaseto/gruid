package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// TextInputConfig describes configuration options for creating a text input.
type TextInputConfig struct {
	Grid       gruid.Grid    // grid slice where the text input is drawn
	StyledText StyledText    // styled text with initial text input text content
	Prompt     string        // optional prompt text
	Box        *Box          // draw optional box around the text input
	Keys       TextInputKeys // optional custom key bindings for the text input
	Style      TextInputStyle
}

// TextInputStyle describes styling options for a TextInput.
type TextInputStyle struct {
	Cursor gruid.Style // cursor style
	Prompt gruid.Style // prompt style, if any
}

// TextInputKeys contains key bindings configuration for the text input.
type TextInputKeys struct {
	Quit []gruid.Key // quit text input (default: Escape, Tab)
}

// TextInput represents a line entry with text supplied from the user that can
// be validated. It offers only basic editing shortcuts.
//
// TextInput implements gruid.Model, but is not suitable for use as main model
// of an application.
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
	dirty     bool       // state changed in Update and Draw was still not called
	drawn     gruid.Grid // the last grid slice that was drawn
}

// TextInputAction represents last user action with the text input.
type TextInputAction int

// These constants represent possible actions raising from interaction with the
// text input.
const (
	TextInputPass   TextInputAction = iota // no change in state
	TextInputChange                        // changed content or moved cursor
	TextInputInvoke                        // invoke/accept content
	TextInputQuit                          // quit/cancel text input
)

// NewTextInput returns a new text input with given configuration options.
func NewTextInput(cfg TextInputConfig) *TextInput {
	ti := &TextInput{
		grid:   cfg.Grid,
		stt:    cfg.StyledText.WithMarkups(nil),
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
	ti.dirty = true
	return ti
}

// SetBox updates the text input surrounding box.
func (ti *TextInput) SetBox(b *Box) {
	ti.box = b
	ti.dirty = true
}

func (ti *TextInput) cursorMax() int {
	return len(ti.content)
}

// Update implements gruid.Model.Update for TextInput. It considers mouse
// message coordinates to be absolute in its grid.
func (ti *TextInput) Update(msg gruid.Msg) gruid.Effect {
	ti.action = TextInputPass
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		ti.updateMsgKeyDown(msg)
	case gruid.MsgMouse:
		ti.updateMsgMouse(msg)
	}
	if ti.action != TextInputPass {
		ti.dirty = true
	}
	return nil
}

func (ti *TextInput) updateMsgKeyDown(msg gruid.MsgKeyDown) {
	if msg.Key.In(ti.keys.Quit) {
		ti.action = TextInputQuit
		return
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
		ti.action = TextInputInvoke
	default:
		if !msg.Key.IsRune() {
			return
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
}

func (ti *TextInput) updateMsgMouse(msg gruid.MsgMouse) {
	cgrid := ti.grid
	if ti.box != nil {
		rg := ti.grid.Range()
		cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	start := ti.start()
	p := msg.P.Sub(cgrid.Bounds().Min)
	switch msg.Action {
	case gruid.MouseMain:
		if !msg.P.In(ti.grid.Bounds()) {
			ti.action = TextInputQuit
			return
		}
		if !cgrid.Contains(p) {
			return
		}
		ocursor := ti.cursor
		ti.cursor = msg.P.X + start - ti.cursorMin - 1
		if ti.cursor > ti.cursorMax() {
			ti.cursor = ti.cursorMax()
		}
		if ti.cursor < 0 {
			ti.cursor = 0
		}
		if ti.cursor != ocursor {
			ti.action = TextInputChange
		}
	}
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
		rg := ti.grid.Range()
		cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	crg := cgrid.Range()
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
	if !ti.dirty {
		return ti.drawn
	}
	cgrid := ti.grid
	if ti.box != nil {
		ti.box.Draw(ti.grid)
		rg := ti.grid.Range()
		cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	cgrid.Fill(gruid.Cell{Rune: ' ', Style: ti.stt.Style()})
	ti.stt.With(ti.prompt, ti.style.Prompt).Draw(cgrid)
	crg := cgrid.Range()
	start := ti.start()
	ti.stt.WithText(string(ti.content[start:])).Draw(cgrid.Slice(crg.Shift(ti.cursorMin, 0, 0, 0)))
	ti.stt.With(string(ti.cursorRune()), ti.style.Cursor).Draw(cgrid.Slice(crg.Shift(ti.cursorMin+ti.cursor-start, 0, 0, 0)))
	ti.dirty = false
	ti.drawn = ti.grid
	return ti.drawn
}
