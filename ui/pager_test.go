package ui

import (
	"testing"

	"github.com/anaseto/gruid"
)

func TestPager(t *testing.T) {
	gd := gruid.NewGrid(10, 6)
	lines := []StyledText{
		Text("line one"),
		Text("line two"),
		Text("line three"),
		Text("line four"),
	}
	pager := NewPager(PagerConfig{
		Grid:  gd,
		Lines: lines,
	})
	sendKey := func(key gruid.Key) {
		pager.Update(gruid.MsgKeyDown{Key: key})
	}
	check := func(b bool, s string) {
		if !b {
			t.Errorf("%s", s)
		}
	}
	check(pager.Action() == PagerPass, "pass")
	check(pager.View().Size().Y == 4, "size")
	sendKey(gruid.KeyEscape)
	check(pager.Action() == PagerQuit, "quit")
	pager.SetLines(append(lines, lines...))
	check(pager.View().Size().Y == 6, "size")
	check(pager.View().Max.Y == 6, "max")
	sendKey(gruid.KeyArrowDown)
	check(pager.Action() == PagerMove, "move down")
	check(pager.View().Size().Y == 6, "size")
	check(pager.View().Max.Y == 7, "max")
	pager.SetLines([]StyledText{})
	check(pager.View().Size().Y == 0, "size zero")
	sendKey(gruid.KeyArrowDown)
	check(pager.Action() == PagerPass, "pass")
}

func TestPagerSetCursor(t *testing.T) {
	gd := gruid.NewGrid(10, 6)
	var lines []StyledText
	for i := 0; i < 20; i++ {
		lines = append(lines, Textf("%d", i))
	}
	pager := NewPager(PagerConfig{
		Grid:  gd,
		Lines: lines,
	})
	for i := -1; i < 21; i++ {
		pager.SetCursor(gruid.Point{0, i})
		view := pager.View()
		if view.Min.Y < 0 {
			t.Errorf("view min y: %d (%d)", view.Min.Y, i)
		}
		if view.Max.Y > 20 {
			t.Errorf("view max y: %d (%d)", view.Max.Y, i)
		}
		if i >= 0 && i <= 14 && view.Max.Y != i+6 {
			t.Errorf("view max y: %d (%d)", view.Max.Y, i)
		}
		if i == 14 && view.Min.Y != i {
			t.Errorf("view min y: %d (%d)", view.Min.Y, i)
		}
		if i == 15 && view.Min.Y == i {
			t.Errorf("view min y: %d (%d)", view.Min.Y, i)
		}
	}
}
