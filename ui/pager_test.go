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
