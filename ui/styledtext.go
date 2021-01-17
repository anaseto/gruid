package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/anaseto/gruid"
)

// StyledText is a simple text formatter and styler. The zero value can be
// used, but you may prefer using Text and Textf.
type StyledText struct {
	markups map[rune]gruid.Style
	text    string
	style   gruid.Style
}

// Text is a shorthand for StyledText{}.WithText and creates a new styled text
// with the given text and the default style.
func Text(text string) StyledText {
	return StyledText{text: text}
}

// Textf returns a new styled text with the given formatted text, and default style.
func Textf(format string, a ...interface{}) StyledText {
	return StyledText{text: fmt.Sprintf(format, a...)}
}

// NewStyledText returns a new styled text with the given text and style. It is
// a shorthand for StyledText{}.With.
func NewStyledText(text string, st gruid.Style) StyledText {
	return StyledText{text: text, style: st}
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

// WithText returns a derived styled text with updated formatted text.
func (stt StyledText) WithTextf(format string, a ...interface{}) StyledText {
	stt.text = fmt.Sprintf(format, a...)
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

// Markups returns a copy of the markups currently defined for the styled text.
func (stt StyledText) Markups() map[rune]gruid.Style {
	markups := make(map[rune]gruid.Style, len(stt.markups))
	for k, v := range stt.markups {
		markups[k] = v
	}
	return markups
}

// Iter iterates a function for all couples positions and cells representing
// the styled text, and returns the minimum (w, h) size in cells which can fit
// the text.
func (stt StyledText) Iter(fn func(gruid.Point, gruid.Cell)) gruid.Point {
	x, y := 0, 0
	xmax := 0
	c := gruid.Cell{Style: stt.style}
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	for _, r := range stt.text {
		if markup {
			if procm {
				procm = false
				if r != '@' {
					c.Style = stt.markupStyle(r)
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
		if fn != nil {
			fn(gruid.Point{X: x, Y: y}, c)
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

// Size returns the minimum (w, h) size in cells which can fit the text.
func (stt StyledText) Size() gruid.Point {
	return stt.Iter(nil)
}

// Format formats the text so that lines longer than a certain width get
// wrapped at word boundaries, if possible. It preserves spaces at the
// beginning of a line.  It returns the modified style for convenience.
func (stt StyledText) Format(width int) StyledText {
	s := strings.Builder{}
	wordbuf := bytes.Buffer{}
	col := 0                     // current column (without counting @r markups)
	wantspace := false           // whether we expect currently space (start of a new word that is not at line start)
	wlen := 0                    // current word length
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	start := true                // whether at line start
	for _, r := range stt.text {
		if markup {
			if procm {
				procm = false
				switch r {
				case '@', '\n', ' ':
				default:
					if wlen == 0 {
						s.WriteRune(r)
					} else {
						wordbuf.WriteRune(r)
					}
					continue
				}
			} else if r == '@' {
				procm = true
				if wlen == 0 {
					s.WriteRune(r)
				} else {
					wordbuf.WriteRune(r)
				}
				continue
			}
		}
		if r == ' ' {
			switch {
			case start:
				s.WriteRune(' ')
				col++
				continue
			case wlen > 0:
				if col+wlen+1 > width && wantspace {
					s.WriteRune('\n')
					col = 0
				} else if wantspace {
					s.WriteRune(' ')
					col++
				}
				s.Write(wordbuf.Bytes())
				col += wlen
				wordbuf.Reset()
				wlen = 0
				wantspace = true
			}
			continue
		}
		if r == '\n' {
			if wlen > 0 {
				if 1+col+wlen > width && wantspace {
					s.WriteRune('\n')
				} else if wantspace {
					s.WriteRune(' ')
				}
				s.Write(wordbuf.Bytes())
				wordbuf.Reset()
				wlen = 0
			}
			s.WriteRune('\n')
			col = 0
			wantspace = false
			start = true
			continue
		}
		start = false
		wordbuf.WriteRune(r)
		wlen++
	}
	if wlen > 0 {
		if wantspace {
			if wlen+col+1 > width {
				s.WriteRune('\n')
			} else {
				s.WriteRune(' ')
			}
		}
		s.Write(wordbuf.Bytes())
	}
	stt.text = strings.TrimRight(s.String(), " \n")
	return stt
}

func (stt StyledText) markupStyle(r rune) gruid.Style {
	if r == 'N' {
		return stt.style
	}
	st, ok := stt.markups[r]
	if ok {
		return st
	}
	return stt.style
}

// Draw displays the styled text in a given grid. It returns the smallest grid
// slice containing the drawn part. Note that the grid is not cleared with
// spaces beforehand by this function, not even the returned one, you should
// use the styled text with a label for this.
func (stt StyledText) Draw(gd gruid.Grid) gruid.Grid {
	it := gd.Iterator()
	if !it.Next() {
		return gd
	}
	x, y := 0, 0
	xmax := 0
	c := gruid.Cell{Style: stt.style}
	markup := stt.markups != nil // whether markup is activated
	procm := false               // processing markup
	for _, r := range stt.text {
		if markup {
			if procm {
				procm = false
				if r != '@' {
					c.Style = stt.markupStyle(r)
					continue
				}
			} else if r == '@' {
				procm = true
				continue
			}
		}
		p := it.P()
		if r == '\n' {
			if x > xmax {
				xmax = x
			}
			x = 0
			y++
			if p.Y < y {
				it.SetP(gruid.Point{x, y})
				if it.P().Y != y {
					break
				}
			}
			continue
		}
		x++
		if p.Y > y {
			continue
		}
		c.Rune = r
		it.SetCell(c)
		if !it.Next() {
			break
		}
	}
	if x > xmax {
		xmax = x
	}
	if xmax > 0 || y > 0 {
		y++ // at least one line
	}
	max := gruid.Point{X: xmax, Y: y}
	return gd.Slice(gruid.NewRange(0, 0, max.X, max.Y))
}
