package ui

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	text := "word word word word word"
	stt := NewStyledText(text)
	w, h := stt.Size()
	if w != len(text) || h != 1 {
		t.Errorf("bad text size: %d, %d", w, h)
	}
	stt = stt.Format(9)
	w, h = stt.Size()
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 2 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	if w != 9 || h != 3 {
		t.Errorf("bad formatted text size: %d, %d. Text:\n%s", w, h, stt.Text())
	}
}
