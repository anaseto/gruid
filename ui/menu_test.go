package ui

import (
	"fmt"
	"testing"

	"github.com/anaseto/gruid"
)

func TestMenu(t *testing.T) {
	gd := gruid.NewGrid(10, 10)
	entries := []MenuEntry{
		{Text: "one"},
		{Text: "two"},
		{Text: "three"},
		{Text: "four", Disabled: true},
	}
	menu := NewMenu(MenuConfig{
		Grid:    gd,
		Entries: entries,
	})
	keymsg := func(key gruid.Key) gruid.Msg {
		return gruid.MsgKeyDown{Key: key}
	}
	check := func(b bool, s string) {
		if !b {
			t.Errorf("%s", s)
		}
	}
	check(menu.Action() == MenuPass, "pass")
	check(menu.Active() == 0, "active 0")
	menu.Update(keymsg(gruid.KeyArrowUp))
	check(menu.Action() == MenuMove, "move up")
	check(menu.Active() == 2, fmt.Sprintf("active %d", menu.Active()))
	menu.Update(keymsg(gruid.KeyArrowUp))
	check(menu.Action() == MenuMove, "move up 2")
	check(menu.Active() == 1, "active 1")
	menu.Update(keymsg("œ"))
	check(menu.Action() == MenuPass, "pass œ")
	check(menu.Active() == 1, "active 1 (pass œ)")
	menu.Update(keymsg(gruid.KeyEnter))
	check(menu.Action() == MenuInvoke, "invoke")
	check(menu.Active() == 1, "active 1 (invoke)")
	menu.Update(keymsg(gruid.KeyEscape))
	check(menu.Action() == MenuQuit, "quit")
	check(menu.Active() == 1, "active 1 (quit)")
	draw := menu.Draw()
	check(draw.Size().Y == 4, "size")
	menu.SetEntries(append(entries, MenuEntry{Text: "five"}))
	draw = menu.Draw()
	check(draw.Size().Y == 5, "size")
	menu.SetBox(&Box{})
	draw = menu.Draw()
	check(draw.Size().Y == 7, "size")
}

func TestTable(t *testing.T) {
	gd := gruid.NewGrid(10, 10)
	entries := []MenuEntry{
		{Text: "one"},
		{Text: "two"},
		{Text: "three"},
		{Text: "four", Disabled: true},
		{Text: "five"},
	}
	menu := NewMenu(MenuConfig{
		Grid:    gd,
		Entries: entries,
		Style:   MenuStyle{Layout: gruid.Point{2, 2}},
	})
	keymsg := func(key gruid.Key) gruid.Msg {
		return gruid.MsgKeyDown{Key: key}
	}
	check := func(b bool, s string) {
		if !b {
			t.Errorf("%s", s)
		}
	}
	check(menu.Action() == MenuPass, "pass")
	check(menu.Active() == 0, "active 0")
	menu.Update(keymsg(gruid.KeyArrowUp))
	check(menu.Action() == MenuMove, "move up")
	check(menu.Active() == 4, fmt.Sprintf("active %d: expected 4", menu.Active()))
	menu.cursorAtLastChoice()
	check(menu.Active() == 4, fmt.Sprintf("active %d: expected 4", menu.Active()))
	menu.Update(keymsg(gruid.KeyPageUp))
	check(menu.Action() == MenuMove, "move page up")
	check(menu.Active() == 3, fmt.Sprintf("active %d: expected 3", menu.Active()))
	menu.Update(keymsg(gruid.KeyArrowLeft))
	check(menu.Action() == MenuMove, "move left 0")
	check(menu.Active() == 1, "active 0")
	menu.SetEntries(append(entries, MenuEntry{Text: "six"}))
	menu.Update(keymsg(gruid.KeyPageDown))
	check(menu.Action() == MenuMove, "move page down")
	check(len(menu.entries) == 6, "entries")
	check(menu.Active() == 4, "active 4")
	menu.cursorAtLastChoice()
	check(menu.Active() == 5, "active 4")
	menu.cursorAtFirstChoice()
	check(menu.Active() == 0, "active 4")
}

func TestStatus(t *testing.T) {
	gd := gruid.NewGrid(10, 10)
	entries := []MenuEntry{
		{Text: "one"},
		{Text: "two"},
		{Text: "three"},
		{Text: "four", Disabled: true},
	}
	menu := NewMenu(MenuConfig{
		Grid:    gd,
		Entries: entries,
		Style:   MenuStyle{Layout: gruid.Point{2, 1}},
	})
	keymsg := func(key gruid.Key) gruid.Msg {
		return gruid.MsgKeyDown{Key: key}
	}
	check := func(b bool, s string) {
		if !b {
			t.Errorf("%s", s)
		}
	}
	check(menu.Action() == MenuPass, "pass")
	check(menu.Active() == 0, "active 0")
	menu.Update(keymsg(gruid.KeyArrowLeft))
	check(menu.Action() == MenuMove, "move left")
	check(menu.Active() == 2, fmt.Sprintf("active %d", menu.Active()))
	menu.Update(keymsg(gruid.KeyPageUp))
	check(menu.Action() == MenuMove, "move page up")
	check(menu.Active() == 1, fmt.Sprintf("active %d", menu.Active()))
}
