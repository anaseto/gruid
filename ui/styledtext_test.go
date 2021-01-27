package ui

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
)

func TestSize(t *testing.T) {
	text := ""
	stt := Text(text)
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
	stt := Text(text)
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
	spaces := strings.Count(stt.Text(), " ")
	if spaces != 2 {
		t.Errorf("bad formatted text number of spaces: %d. Text:\n%s", spaces, stt.Text())
	}
	if max.X != 9 || max.Y != 3 {
		t.Errorf("bad formatted text size: %v. Text:\n%s", max, stt.Text())
	}
}

func TestFormat2(t *testing.T) {
	text := "word word word word word"
	stt := Text(text)
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
	stt := Text(text)
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
	stt := Text(text)
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

func TestFormat5(t *testing.T) {
	text := "word word word word word"
	stt := Text(text)
	stt = stt.Format(20)
	newlines := strings.Count(stt.Text(), "\n")
	if newlines != 1 {
		t.Errorf("bad formatted text number of lines: %d. Text:\n%s", newlines, stt.Text())
	}
	spaces := strings.Count(stt.Text(), " ")
	if spaces != 3 {
		t.Errorf("bad formatted text number of spaces: %d. Text:\n%s", spaces, stt.Text())
	}
}

func TestFormat6(t *testing.T) {
	text := "word word word word word"
	stt := Text(text)
	stt = stt.Format(10)
	if stt.Text() != stt.Format(10).Text() {
		t.Errorf("not idempotent:\n[%v]\n[%v]", stt.Text(), stt.Format(10).Text())
	}
}

func TestFormatWithMarkup(t *testing.T) {
	text := "word @cword@N word word word"
	st := gruid.Style{}
	stt := Text(text).WithMarkup('c', st)
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
	stt := Text("@t•@N ").WithMarkup('t', st)
	if stt.Size().X != 2 || stt.Size().Y != 1 {
		t.Errorf("bad size: %v", stt.Size())
	}
	count := 0
	stt.Iter(func(p gruid.Point, c gruid.Cell) { count++ })
	if count != 2 {
		t.Errorf("bad count: %v", count)
	}
	stt = stt.Format(10)
	if stt.Size().X != 1 || stt.Size().Y != 1 {
		t.Errorf("bad size: %v", stt.Size())
	}
}

func TestMarkupConsecutive(t *testing.T) {
	st := gruid.Style{}
	stt := Text("@N@@@t•@N ").WithMarkup('t', st)
	if stt.Size().X != 3 || stt.Size().Y != 1 {
		t.Errorf("bad size: %v", stt.Size())
	}
	count := 0
	stt.Iter(func(p gruid.Point, c gruid.Cell) { count++ })
	if count != 3 {
		t.Errorf("bad count: %v", count)
	}
}

func TestDrawLine(t *testing.T) {
	gd := gruid.NewGrid(5, 2)
	Text("xxxx").Draw(gd)
	gd.Iter(func(p gruid.Point, c gruid.Cell) {
		if p.Y == 0 && p.X < 4 {
			if c.Rune != 'x' {
				t.Errorf("not x")
			}
		} else if c.Rune == 'x' {
			t.Errorf("unexpected x at %v", p)
		}
	})
	gd.Fill(gruid.Cell{Rune: ' '})
	Text("xxxxx").Draw(gd)
	gd.Iter(func(p gruid.Point, c gruid.Cell) {
		if p.Y == 0 {
			if c.Rune != 'x' {
				t.Errorf("not x")
			}
		} else if c.Rune == 'x' {
			t.Errorf("unexpected x")
		}
	})
}

func TestDrawPar(t *testing.T) {
	gd := gruid.NewGrid(5, 2)
	Text("xxxxxx\nx").Draw(gd)
	gd.Iter(func(p gruid.Point, c gruid.Cell) {
		if p.Y == 0 && c.Rune != 'x' {
			t.Errorf("not x")
		}
		if p.Y == 1 && p.X > 0 && c.Rune == 'x' {
			t.Errorf("unexpected x")
		}
		if p.Y == 1 && p.X == 0 && c.Rune != 'x' {
			t.Errorf("not x")
		}
	})
	Text("xxxxxx\nxxxxxxxxx").Draw(gd)
	gd.Iter(func(p gruid.Point, c gruid.Cell) {
		if c.Rune != 'x' {
			t.Errorf("not x")
		}
	})
}

func BenchmarkTextSize(b *testing.B) {
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20))
	for i := 0; i < b.N; i++ {
		stt.Size()
	}
}

func BenchmarkTextSizeWithMarkup(b *testing.B) {
	st := gruid.Style{}
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20)).WithMarkup('t', st)
	for i := 0; i < b.N; i++ {
		stt.Size()
	}
}

func BenchmarkTextDrawShort(b *testing.B) {
	gd := gruid.NewGrid(80, 24)
	stt := Text("A short sentence\n")
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextDrawMedium(b *testing.B) {
	gd := gruid.NewGrid(80, 24)
	stt := Text("A short sentence but not so short than the other\n")
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextDrawLong(b *testing.B) {
	gd := gruid.NewGrid(80, 24)
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20))
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextDrawLongWithMarkup(b *testing.B) {
	gd := gruid.NewGrid(80, 24)
	st := gruid.Style{}
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20)).WithMarkup('t', st)
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextDrawLongThinGrid(b *testing.B) {
	gd := gruid.NewGrid(10, 24)
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20))
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextDrawLongShortGrid(b *testing.B) {
	gd := gruid.NewGrid(80, 5)
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20))
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextDrawMediumWithMarkup(b *testing.B) {
	gd := gruid.NewGrid(80, 24)
	st := gruid.Style{}
	stt := Text("A short sentence but not so short than the other\n").WithMarkup('t', st)
	for i := 0; i < b.N; i++ {
		stt.Draw(gd)
	}
}

func BenchmarkTextFormat(b *testing.B) {
	stt := Text(strings.Repeat("A test sentence that says nothing interesting\n", 20))
	for i := 0; i < b.N; i++ {
		stt.Format(30)
	}
}
