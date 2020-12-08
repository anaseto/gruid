package gruid

import "time"

// Key represents the name of a key press.
type Key string

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
	KeySpace      Key = " "
	KeyTab        Key = "Tab"
)

// MsgKeyDown represents a key press.
type MsgKeyDown struct {
	Key  Key       // name of the key in MsgKeyDown event
	Time time.Time // time when the event was generated

	// Shift reports whether the shift key was pressed when the event
	// occured.  This may be not supported equally well accross all
	// platforms. In particular, terminal drivers may not report it for key
	// presses corresponding to upper case letters. It may conflict in some
	// cases with browser or system shortcuts too.
	Shift bool
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
	MouseRelease                      // button release
	MouseMove                         // mouse motion
)

func (ma MouseAction) String() string {
	var s string
	switch ma {
	case MouseMain:
		s = "button main"
	case MouseAuxiliary:
		s = "button auxiliary"
	case MouseSecondary:
		s = "button secondary"
	case MouseRelease:
		s = "button release"
	case MouseMove:
		s = "move"
	}
	return s
}

// MsgMouse represents a mouse user event.
type MsgMouse struct {
	Action   MouseAction // mouse action (click, release, move)
	MousePos Position    // mouse position in the grid
	Time     time.Time   // time when the event was generated
}

type msgQuit struct{}

// MsgScreenSize is used to report the screen size, when it makes sense.
type MsgScreenSize struct {
	Width  int       // width in cells
	Height int       // height in cells
	Time   time.Time // time when the event was generated
}

// Quit returns a special command that signals the application to exit. Note
// that the application does not wait for pending effects to complete before
// exiting the Start loop, so you have to wait for those command messages
// before using Quit.
func Quit() Cmd {
	return func() Msg {
		return msgQuit{}
	}
}

// msgBatch is an internal message used to perform a bunch of effects. You can
// send a msgBatch with Batch.
type msgBatch []Effect
