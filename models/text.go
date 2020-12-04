package models

import (
	"bytes"
	"strings"

	"github.com/anaseto/gruid"
)

// StyledText is a simple text formatter and styler.
type StyledText struct {
	text    string
	style   gruid.CellStyle
	markups map[rune]gruid.CellStyle
}

// NewStyledText returns a new styled text with the default style.
func NewStyledText(text string) *StyledText {
	return &StyledText{text: text}
}

// SetStyle changes the default content style.
func (stt *StyledText) SetStyle(style gruid.CellStyle) {
	stt.style = style
}

// RegisterMarkup defines simple markup for the text. Markup starts by a %
// sign, and is followed then by a rune indicating a particular style. Default
// style is marked by %N.
//
// This simple markup is inspired from github/gdamore/tcell, but with somewhat
// different defaults: unless at least one markup is registered, markup
// commands processing is not activated, and % is treated as any other
// character.
func (stt *StyledText) RegisterMarkup(r rune, style gruid.CellStyle) {
	if len(stt.markups) == 0 {
		stt.markups = map[rune]gruid.CellStyle{}
		stt.markups['N'] = gruid.CellStyle{}
	}
	if r == 'N' {
		// N has a built-in meaning
		stt.SetStyle(style)
		return
	}
	if r == ' ' || r == '\n' {
		// avoid strange cases that can conflict with format
		return
	}
	stt.markups[r] = style
}

// Format formats the text so that lines longer than a certain width get
// wrapped at word boundaries. It preserves spaces at the beginning of a line.
func (stt *StyledText) Format(width int) {
	pbuf := bytes.Buffer{}
	wordbuf := bytes.Buffer{}
	col := 0                       // current column (without counting %r markups)
	wantspace := false             // whether we expect currently space (start of a new word that is not at line start)
	wlen := 0                      // current word length
	markup := len(stt.markups) > 0 // whether markup is activated
	proc := false                  // processing markup
	start := true                  // whether at line start
	for _, r := range stt.text {
		if r == ' ' || r == '\n' {
			if wlen == 0 && r == ' ' && !start {
				continue
			}
			if col+wlen > width {
				if wantspace {
					pbuf.WriteRune('\n')
					col = 0
				}
			} else if wantspace || start {
				pbuf.WriteRune(' ')
				col++
			}
			if wlen > 0 {
				pbuf.Write(wordbuf.Bytes())
				col += wlen
				wordbuf.Reset()
				wlen = 0
			}
			if r == '\n' {
				pbuf.WriteRune('\n')
				col = 0
				wantspace = false
				start = true
			} else if !start {
				wantspace = true
			}
			continue
		}
		if markup {
			if proc {
				proc = false
				if wordbuf.Len() == 0 {
					pbuf.WriteRune(r)
				} else {
					wordbuf.WriteRune(r)
				}
				continue
			} else if r == '%' {
				proc = true
				if wordbuf.Len() == 0 {
					pbuf.WriteRune(r)
				} else {
					wordbuf.WriteRune(r)
				}
				continue
			}
		}
		start = false
		wordbuf.WriteRune(r)
		wlen++
	}
	if wordbuf.Len() > 0 {
		if wantspace {
			if wlen+col > width {
				pbuf.WriteRune('\n')
			} else {
				pbuf.WriteRune(' ')
			}
		}
		pbuf.Write(wordbuf.Bytes())
	}
	stt.text = strings.TrimRight(pbuf.String(), " \n")
}

// Draw displays the styled text in a given grid.
func (stt *StyledText) Draw(gd gruid.Grid) {
	x, y := 0, 0
	c := gruid.Cell{Style: stt.style}
	markup := len(stt.markups) > 0 // whether markup is activated
	proc := false                  // processing markup
	for _, r := range stt.text {
		if markup {
			if proc {
				proc = false
				if r != '%' {
					st, ok := stt.markups[r]
					if ok {
						c.Style = st
					}
					continue
				}
			} else if r == '%' {
				proc = true
			}
		}
		if r == '\n' {
			x = 0
			y++
		}
		c.Rune = r
		gd.SetCell(gruid.Position{X: x, Y: y}, c)
		x++
	}
}
