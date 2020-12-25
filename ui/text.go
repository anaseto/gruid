package ui

import (
	"bytes"
	"strings"

	"github.com/anaseto/gruid"
)

// StyledText is a simple text formatter and styler.
type StyledText struct {
	text    string
	style   gruid.Style
	markups map[rune]gruid.Style
}

// NewStyledText returns a new styled text with the default style.
func NewStyledText(text string) StyledText {
	return StyledText{text: text}
}

// Text returns the current styled text as a string.
func (stt StyledText) Text() string {
	return stt.text
}

// WithText returns a derived styled text with updated text.
func (stt StyledText) WithText(text string) StyledText {
	stt.text = text
	return stt
}

// Style returns the text default style.
func (stt StyledText) Style() gruid.Style {
	return stt.style
}

// WithStyle returns a derived styled text with a updated default style.
func (stt StyledText) WithStyle(style gruid.Style) StyledText {
	stt.style = style
	return stt
}

// With returns a derived styled text with new next and style.
func (stt StyledText) With(text string, style gruid.Style) StyledText {
	stt.text = text
	stt.style = style
	return stt
}

// WithMarkup returns a derived styled text with a new markup @r available for
// a given style.  Markup starts by a @ sign, and is followed then by a rune
// indicating the particular style. Default style is marked by @N.
//
// This simple markup is inspired from github/gdamore/tcell, but with somewhat
// different defaults: unless at least one non-default markup is registered,
// markup commands processing is not activated, and @ is treated as any other
// character.
func (stt StyledText) WithMarkup(r rune, style gruid.Style) StyledText {
	if r == ' ' || r == '\n' {
		// avoid strange cases that can conflict with format
		return stt
	}
	if len(stt.markups) == 0 {
		stt.markups = map[rune]gruid.Style{}
	} else {
		omarks := stt.markups
		stt.markups = make(map[rune]gruid.Style, len(omarks))
		for k, v := range omarks {
			stt.markups[k] = v
		}
	}
	if r == 'N' {
		// N has a built-in meaning
		stt.style = style
		return stt
	}
	stt.markups[r] = style
	return stt
}

// WithMarkups is the same as WithMarkup but passing a whole map of rune-style
// associations.
func (stt StyledText) WithMarkups(markups map[rune]gruid.Style) StyledText {
	stt.markups = markups
	return stt
}

// Size returns the minimum (w, h) size in cells which can fit the text.
func (stt StyledText) Size() gruid.Point {
	x := 0
	xmax := 0
	y := 0
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	for _, r := range stt.text {
		if markup {
			if procm {
				procm = false
				if r != '@' {
					continue
				}
			} else if r == '@' {
				procm = true
				continue
			}
		}
		if r == '\n' {
			if x > xmax {
				xmax = x
			}
			x = 0
			y++
			continue
		}
		x++
	}
	if x > xmax {
		xmax = x
	}
	if xmax > 0 || y > 0 {
		y++ // at least one line
	}
	return gruid.Point{X: xmax, Y: y}
}

// Format formats the text so that lines longer than a certain width get
// wrapped at word boundaries, if possible. It preserves spaces at the
// beginning of a line.  It returns the modified style for convenience.
func (stt StyledText) Format(width int) StyledText {
	pbuf := bytes.Buffer{}
	wordbuf := bytes.Buffer{}
	col := 0                     // current column (without counting @r markups)
	wantspace := false           // whether we expect currently space (start of a new word that is not at line start)
	wlen := 0                    // current word length
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	start := true                // whether at line start
	for _, r := range stt.text {
		if r == ' ' {
			switch {
			case start:
				pbuf.WriteRune(' ')
				col++
				continue
			case wlen > 0:
				if col+wlen+1 > width && wantspace {
					pbuf.WriteRune('\n')
					col = 0
				} else if wantspace {
					pbuf.WriteRune(' ')
					col++
				}
				pbuf.Write(wordbuf.Bytes())
				col += wlen
				wordbuf.Reset()
				wlen = 0
				wantspace = true
			}
			continue
		}
		if r == '\n' {
			if wlen > 0 {
				if col+wlen > width && wantspace {
					pbuf.WriteRune('\n')
				}
				pbuf.Write(wordbuf.Bytes())
				wordbuf.Reset()
				wlen = 0
			}
			pbuf.WriteRune('\n')
			col = 0
			wantspace = false
			start = true
			continue
		}
		if markup {
			if procm {
				procm = false
				if r != '@' {
					if wlen == 0 {
						pbuf.WriteRune(r)
					} else {
						wordbuf.WriteRune(r)
					}
					continue
				}
			} else if r == '@' {
				procm = true
				if wlen == 0 {
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
	if wlen > 0 {
		if wantspace {
			if wlen+col+1 > width {
				pbuf.WriteRune('\n')
			} else {
				pbuf.WriteRune(' ')
			}
		}
		pbuf.Write(wordbuf.Bytes())
	}
	stt.text = strings.TrimRight(pbuf.String(), " \n")
	return stt
}

// Draw displays the styled text in a given grid. It returns the smallest grid
// slice containing the drawn part. Note that the grid is not cleared with
// spaces beforehand by this function, not even the returned one, you should
// use the styled text with a label for this.
func (stt StyledText) Draw(gd gruid.Grid) gruid.Grid {
	x, y := 0, 0
	c := gruid.Cell{Style: stt.style}
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	xmax := 0
	for _, r := range stt.text {
		if markup {
			if procm {
				procm = false
				if r != '@' {
					if r == 'N' {
						c.Style = stt.style
						continue
					}
					st, ok := stt.markups[r]
					if ok {
						c.Style = st
					}
					continue
				}
			} else if r == '@' {
				procm = true
				continue
			}
		}
		if r == '\n' {
			if x > xmax {
				xmax = x
			}
			x = 0
			y++
			continue
		}
		c.Rune = r
		gd.Set(gruid.Point{X: x, Y: y}, c)
		x++
	}
	if x > xmax {
		xmax = x
	}
	if xmax > 0 || y > 0 {
		y++ // at least one line
	}
	return gd.Slice(gruid.NewRange(0, 0, xmax, y))
}
