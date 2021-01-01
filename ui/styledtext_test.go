package ui

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
)

func TestSize(t *testing.T) {
	text := ""
	stt := NewStyledText(text)
	max := stt.Size()
	if max.X != 0 || max.Y != 0 {
		t.Errorf("bad text size: %v", max)
	}
	stt = stt.WithText("word")
	max = stt.Size()
	if max.X != 4 || max.Y != 1 {
		t.Errorf("bad text size: %v", max)
	}
	stt = stt.WithText("word\nword")
	max = stt.Size()
	if max.X != 4 || max.Y != 2 {
		t.Errorf("bad text size: %v", max)
	}
}

func TestFormat(t *testing.T) {
	text := "word word word word word"
	stt := NewStyledText(text)
	max := stt.Size()
	if max.X != 4*5+4 || max.Y != 1 {
		t.Errorf("bad text size: %v", max)
	}
	stt = stt.Format(9)
	max = stt.Size()
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 2 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	if max.X != 9 || max.Y != 3 {
		t.Errorf("bad formatted text size: %v. Text:\n%s", max, stt.Text())
	}
}

func TestFormat2(t *testing.T) {
	text := "word word word word word"
	stt := NewStyledText(text)
	stt = stt.Format(8)
	max := stt.Size()
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 4 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	if max.X != 4 || max.Y != 5 {
		t.Errorf("bad formatted text size: %v. Text:\n%s", max, stt.Text())
	}
}

func TestFormat3(t *testing.T) {
	text := "word word word word word"
	stt := NewStyledText(text)
	stt = stt.Format(10)
	max := stt.Size()
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 2 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	if max.X != 9 || max.Y != 3 {
		t.Errorf("bad formatted text size: %v. Text:\n%s", max, stt.Text())
	}
}

func TestFormat4(t *testing.T) {
	text := "word word word word word"
	stt := NewStyledText(text)
	stt = stt.Format(1)
	max := stt.Size()
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 4 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	if max.X != 4 || max.Y != 5 {
		t.Errorf("bad formatted text size: %v. Text:\n%s", max, stt.Text())
	}
}

func TestFormatWithMarkup(t *testing.T) {
	text := "word @cword@N word word word"
	st := gruid.Style{}
	stt := NewStyledText(text).WithMarkup('c', st)
	max := stt.Size()
	if max.X != 4*5+4 || max.Y != 1 {
		t.Errorf("bad text size: %v", max)
	}
	stt = stt.Format(9)
	max = stt.Size()
	gd := gruid.NewGrid(80, 20)
	gd = stt.Draw(gd)
	if gd.Size() != max {
		t.Errorf("bad size %v vs grid %v", max, gd.Size())
	}
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 2 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	if max.X != 9 || max.Y != 3 {
		t.Errorf("bad formatted text size: %v. Text:\n%s", max, stt.Text())
	}
}

func TestSizeMarkup(t *testing.T) {
	st := gruid.Style{}
	stt := NewStyledText("@t•@N ").WithMarkup('t', st)
	if stt.Size().X != 2 || stt.Size().Y != 1 {
		t.Errorf("bad size: %v", stt.Size())
	}
	stt = stt.Format(10)
	if stt.Size().X != 1 || stt.Size().Y != 1 {
		t.Errorf("bad size: %v", stt.Size())
	}
}
