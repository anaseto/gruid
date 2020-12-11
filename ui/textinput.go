package ui

import (
	"unicode/utf8"

	"github.com/anaseto/gruid"
)

// TextInputStyle describes styling options for a TextInput.
type TextInputStyle struct {
	Boxed  bool        // draw a box around the text input
	Box    gruid.Style // box style, if any
	Cursor gruid.Style // cursor style
	Title  gruid.Style // box title style, if any
	Prompt gruid.Style // prompt style, if any
}

// TextInputConfig describes configuration options for creating a text input.
type TextInputConfig struct {
	Grid       gruid.Grid // grid slice where the text input is drawn
	StyledText StyledText // styled text with initial text input text content
	Title      string     // optional title, implies Boxed style
	Prompt     string     // optional prompt text
	Style      TextInputStyle
}

// TextInput represents an entry with text supplied from the user that can be
// validated.
type TextInput struct {
	grid      gruid.Grid
	stt       StyledText
	title     string
	prompt    string
	content   []rune
	style     TextInputStyle
	cursorMin int
	cursor    int
	action    TextInputAction
	init      bool // received gruid.MsgInit
}

// TextInputAction represents last user action with the text input.
type TextInputAction int

// These constants represent available actions araising from interaction with
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
		title:  cfg.Title,
		prompt: cfg.Prompt,
		style:  cfg.Style,
	}
	if ti.style.Cursor == ti.stt.Style() {
		// not true reverse with terminal driver, but good enough as a default
		ti.style.Cursor.Bg, ti.style.Cursor.Fg = ti.style.Cursor.Fg, ti.style.Cursor.Bg
	}
	ti.cursorMin = utf8.RuneCountInString(ti.prompt)
	ti.content = []rune(ti.stt.Text())
	ti.cursor = len(ti.content)
	if ti.title != "" {
		ti.style.Boxed = true
	}
	return ti
}

func (ti *TextInput) cursorMax() int {
	return len(ti.content)
}

// Update implements gruid.Model.Update.
func (ti *TextInput) Update(msg gruid.Msg) gruid.Effect {
	ti.action = TextInputPass
	switch msg := msg.(type) {
	case gruid.MsgInit:
		ti.init = true
	case gruid.MsgKeyDown:
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
		case gruid.KeyEscape:
			ti.action = TextInputQuit
			if ti.init {
				return gruid.Quit()
			}
		default:
			nchars := utf8.RuneCountInString(string(msg.Key))
			if nchars != 1 {
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
		// TODO
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

// Draw implements gruid.Model.Draw.
func (ti *TextInput) Draw() gruid.Grid {
	cgrid := ti.grid
	if ti.style.Boxed {
		b := box{
			grid:  ti.grid,
			title: ti.stt.With(ti.title, ti.style.Title),
			style: ti.style.Box,
		}
		b.draw()
		rg := ti.grid.Range().Relative()
		cgrid = ti.grid.Slice(rg.Shift(1, 1, -1, -1))
	}
	ti.stt.With(ti.prompt, ti.style.Prompt).Draw(cgrid)
	crg := cgrid.Range().Relative()
	start := 0
	w, _ := crg.Size()
	w -= ti.cursorMin
	if ti.cursor >= w {
		if w > 2 {
			start = ti.cursor - w + 2
		} else {
			start = ti.cursor - w
		}
	}
	ti.stt.WithText(string(ti.content[start:])).Draw(cgrid.Slice(crg.Shift(ti.cursorMin, 0, 0, 0)))
	ti.stt.With(string(ti.cursorRune()), ti.style.Cursor).Draw(cgrid.Slice(crg.Shift(ti.cursorMin+ti.cursor-start, 0, 0, 0)))
	return ti.grid
}
