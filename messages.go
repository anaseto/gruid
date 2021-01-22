package gruid

import (
	"time"
	"unicode/utf8"
)

// Key represents the name of a key press.
type Key string

// In reports whether the key is found among a given list of keys.
func (k Key) In(keys []Key) bool {
	for _, key := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// IsRune reports whether the key is a single-rune string and not a named key.
func (k Key) IsRune() bool {
	return utf8.RuneCountInString(string(k)) == 1
}

// This is the list of the supported non single-character named keys. The
// drivers that support keypad with numlock off may report some KP_* keypad
// keys as one of this list, as specified in the comments.
const (
	KeyArrowDown  Key = "ArrowDown"  // can be KP_2
	KeyArrowLeft  Key = "ArrowLeft"  // can be KP_4
	KeyArrowRight Key = "ArrowRight" // can be KP_6
	KeyArrowUp    Key = "ArrowUp"    // can be KP_8
	KeyBackspace  Key = "Backspace"
	KeyDelete     Key = "Delete"
	KeyEnd        Key = "End"   // can be KP_1
	KeyEnter      Key = "Enter" // can be KP_5 (arbitrary choice)
	KeyEscape     Key = "Escape"
	KeyHome       Key = "Home" // can be KP_7
	KeyInsert     Key = "Insert"
	KeyPageDown   Key = "PageDown" // can be KP_3
	KeyPageUp     Key = "PageUp"   // can be KP_9
	KeySpace      Key = " "        // constant for clarity (single character)
	KeyTab        Key = "Tab"
)

// ModMask is a bit mask of modifier keys.
type ModMask int16

// These values represent modifier keys for a MsgKeyDown message. Those are not
// supported equally well across all platforms and drivers, for both technical
// and simplicity reasons. In particular, terminal drivers may not report shift
// for key presses corresponding to upper case letters. Modifiers may conflict
// in some cases with browser or system shortcuts too.  If you want portability
// across platforms and drivers, your application should not depend on them
// for its core functionality.
const (
	ModShift ModMask = 1 << iota
	ModCtrl
	ModAlt
	ModMeta
	ModNone ModMask = 0
)

func (m ModMask) String() string {
	var s string
	addmod := func(mod string) {
		if s != "" {
			s += "+"
		}
		s += mod
	}
	if m&ModCtrl != 0 {
		addmod("Ctrl")
	}
	if m&ModAlt != 0 {
		addmod("Alt")
	}
	if m&ModMeta != 0 {
		addmod("Meta")
	}
	if m&ModShift != 0 {
		addmod("Shift")
	}
	if s == "" {
		s = "None"
	}
	return s
}

// MsgKeyDown represents a key press.
type MsgKeyDown struct {
	Key Key // name of the key in MsgKeyDown event

	// Mod represents modifier keys. They are not portable across
	// different platforms and drivers. Avoid using them for core
	// functionality in portable applications.
	Mod ModMask

	Time time.Time // time when the event was generated
}

// MouseAction represents mouse buttons.
type MouseAction int

// This is the list of supported mouse buttons and actions. It is intentionally
// short for simplicity and best portability across drivers. Pressing several
// mouse buttons simultaneously is not reported and, in those cases, only one
// release event will be sent.
const (
	MouseMain      MouseAction = iota // left button
	MouseAuxiliary                    // middle button
	MouseSecondary                    // right button
	MouseWheelUp                      // wheel impulse up
	MouseWheelDown                    // wheel impulse down
	MouseRelease                      // button release
	MouseMove                         // mouse motion
)

func (ma MouseAction) String() string {
	var s string
	switch ma {
	case MouseMain:
		s = "MouseMain"
	case MouseAuxiliary:
		s = "MouseAuxiliary"
	case MouseSecondary:
		s = "MouseSecondary"
	case MouseWheelUp:
		s = "MouseWheelUp"
	case MouseWheelDown:
		s = "MouseWheelDown"
	case MouseRelease:
		s = "MouseRelease"
	case MouseMove:
		s = "MouseMove"
	}
	return s
}

// MsgMouse represents a mouse user input event.
type MsgMouse struct {
	Action MouseAction // mouse action (click, release, move)
	P      Point       // mouse position in the grid
	Mod    ModMask     // modifier keys (unequal driver support)
	Time   time.Time   // time when the event was generated
}

// MsgScreen is reported by some drivers when the screen has been exposed in
// some way and a complete redraw is necessary. It may happen for example after
// a resize, or after a change of tile set invalidating current displayed content.
// Note that the application takes care of the redraw, so you may not need to
// handle it in most cases, unless you want to adapt grid size and layout
// in response to a potential screen resize.
type MsgScreen struct {
	Width  int       // screen width in cells
	Height int       // screen height in cells
	Time   time.Time // time when the event was generated
}

// MsgInit is a special message that is always sent first to Update after
// calling Start on the application.
type MsgInit struct{}

// MsgQuit may be reported by some drivers to request termination of the
// application, such as when the main window is closed. It reports the time at
// which the driver's request was received.
type MsgQuit time.Time

// msgEnd is an internal message used to end the application's Start loop. It
// is manually produced by the End() command.
type msgEnd struct{}

// msgBatch is an internal message used to perform a bunch of effects. You can
// send a msgBatch with Batch.
type msgBatch []Effect
