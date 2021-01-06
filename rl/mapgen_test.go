package rl

import (
	"strings"
	"testing"
)

const vaultExample = `
#.#...
......
..####`

const vaultExampleRotated180 = `
####..
......
...#.#`

const vaultExampleRotated = `
..#
..#
..#
#.#
...
#..`

const vaultExampleReflected = `
...#.#
......
####..`

func TestVault(t *testing.T) {
	v := &Vault{}
	err := v.Parse(vaultExample)
	if err != nil {
		t.Errorf("Parse: %v", err)
	}
	if v.Size().X != 6 || v.Size().Y != 3 {
		t.Errorf("bad size: %v", v.Size())
	}
	v.Rotate(1)
	if v.Size().X != 3 || v.Size().Y != 6 {
		t.Errorf("bad size: %v", v.Size())
	}
	if v.Content() != strings.TrimSpace(vaultExampleRotated) {
		t.Errorf("bad rotation 1:`%v`", v.Content())
	}
	v.Rotate(-1)
	if v.Content() != strings.TrimSpace(vaultExample) {
		t.Errorf("bad rotation -1:`%v`", v.Content())
	}
	v.Rotate(2)
	if v.Content() != strings.TrimSpace(vaultExampleRotated180) {
		t.Errorf("bad rotation 2:`%v`", v.Content())
	}
	v.Rotate(2)
	v.Reflect()
	if v.Content() != strings.TrimSpace(vaultExampleReflected) {
		t.Errorf("bad reflection:`%v`", v.Content())
	}
}

func TestVaultSetRunes(t *testing.T) {
	v := &Vault{}
	v.SetRunes("@")
	err := v.Parse(vaultExample)
	if err == nil {
		t.Error("incomplete rune check")
	}
	v.SetRunes(".#")
	err = v.Parse(vaultExample)
	if err != nil {
		t.Error("bad rune check")
	}
}
