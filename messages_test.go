package gruid

import "testing"

func TestKey(t *testing.T) {
	keys := []Key{"a", "b", "c"}
	if !Key("b").In(keys) {
		t.Error("not in keys")
	}
	if !Key("b").IsRune() {
		t.Error("not rune")
	}
	if Key(KeyEscape).IsRune() {
		t.Error("escape is rune")
	}
}

func TestModMask(t *testing.T) {
	mod := ModShift | ModCtrl | ModAlt | ModMeta
	if mod.String() != "Ctrl+Alt+Meta+Shift" {
		t.Errorf("bad mod String: %v", mod.String())
	}
	if ModNone.String() != "None" {
		t.Errorf("bad empty mod String: %v", ModNone.String())
	}
}

func TestRelMsg(t *testing.T) {
	m := MsgMouse{}
	m.P = Point{7, 6}
	rg := NewRange(5, 5, 20, 20)
	nm := rg.RelMsg(m).(MsgMouse)
	p := Point{2, 1}
	if nm.P != p {
		t.Errorf("bad relative position: %v", nm.P)
	}
}
