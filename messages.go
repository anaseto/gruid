package gruid

import "time"

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

// This is the list of the supported non single-character named keys. The
// drivers that support non-numerical keypad may report some KP_* keypad keys
// as one of this list, as specified in the comments.
const (
	KeyArrowDown  Key = "ArrowDown"  // can be KP_2
	KeyArrowLeft  Key = "ArrowLeft"  // can be KP_4
	KeyArrowRight Key = "ArrowRight" // can be KP_6
	KeyArrowUp    Key = "ArrowUp"    // can be KP_8
	KeyBackspace  Key = "Backspace"
	KeyDelete     Key = "Delete"
	KeyEnd        Key = "End"   // can be KP_1
	KeyEnter      Key = "Enter" // can be KP_5
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
// supported equally well accross all platforms and drivers, for both technical
// and simplicity reasons. In particular, terminal drivers may not report shift
// for key presses corresponding to upper case letters. Modifiers may conflict
// in some cases with browser or system shortcuts too.  If you want portability
// accross platforms and drivers, your application should not depend on them
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
	if m&ModCtrl != 0 {
		s += "Ctrl"
	}
	if m&ModAlt != 0 {
		if s != "" {
			s += "+"
		}
		s += "Alt"
	}
	if m&ModMeta != 0 {
		if s != "" {
			s += "+"
		}
		s += "Meta"
	}
	if m&ModShift != 0 {
		if s != "" {
			s += "+"
		}
		s += "Shift"
	}
	if s == "" {
		s = "None"
	}
	return s
}

// MsgKeyDown represents a key press.
type MsgKeyDown struct {
	Key Key // name of the key in MsgKeyDown event

	// Mod represents modifier keys. They are not portable accross
	// different platforms and drivers. Avoid using them for core
	// functionality in portable applications.
	Mod ModMask

	Time time.Time // time when the event was generated
}

// MouseAction represents mouse buttons.
type MouseAction int

// This is the list of supported mouse buttons and actions. It is intentionally
// short for simplicity and best portability accross drivers. Pressing several
// mouse buttons simultaneously is not reported and, in those cases, only one
// release event will be send.
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
	Action   MouseAction // mouse action (click, release, move)
	MousePos Position    // mouse position in the grid
	Mod      ModMask     // modifier keys: not reported in most drivers
	Time     time.Time   // time when the event was generated
}

// MsgScreenSize is used by some drivers to report the screen size, either
// initially or on resize.
type MsgScreenSize struct {
	Width  int       // width in cells
	Height int       // height in cells
	Time   time.Time // time when the event was generated
}

// MsgInit is a special message that is always sent first to Update after
// calling Start on the application.
type MsgInit struct{}

// MsgDraw reports that this Update will be followed by a call to Draw. It is
// regularly reported at the FPS rate.
type MsgDraw time.Time

// MsgQuit may be reported by some drivers to request termination of the
// application, such as when the main window is closed.
type MsgQuit struct{}

// msgEnd is an internal message used to end the application's Start loop. It
// is manually produced by the End() command.
type msgEnd struct{}

// msgBatch is an internal message used to perform a bunch of effects. You can
// send a msgBatch with Batch.
type msgBatch []Effect
