package gruid

import "time"

// Key represents the name of a key press.
type Key string

// This is the list of the supported non single-character named keys.
const (
	KeyArrowDown  Key = "ArrowDown"
	KeyArrowLeft  Key = "ArrowLeft"
	KeyArrowRight Key = "ArrowRight"
	KeyArrowUp    Key = "ArrowUp"
	KeyBackspace  Key = "Backspace"
	KeyDelete     Key = "Delete"
	KeyEnd        Key = "End"
	KeyEnter      Key = "Enter"
	KeyEscape     Key = "Escape"
	KeyHome       Key = "Home"
	KeyInsert     Key = "Insert"
	KeyPageDown   Key = "PageDown"
	KeyPageUp     Key = "PageUp"
	KeySpace      Key = " "
	KeyTab        Key = "Tab"
)

// MsgKeyDown represents a key press.
type MsgKeyDown struct {
	Key   Key       // name of the key in MsgKeyDown event
	Time  time.Time // time when the event was generated
	shift bool      //
}

// ShiftKey indicates if the shift key was pressed when the event occured. This
// may be not supported equally well accross all platforms.
func (ev MsgKeyDown) ShiftKey() bool {
	return ev.shift
}

// MouseAction represents mouse buttons.
type MouseAction int

// This is the list of supported mouse buttons and actions.
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

// MsgScreenSize is used to report the screen size.
type MsgScreenSize struct {
	Width  int       // width in cells
	Height int       // height in cells
	Time   time.Time // time when the event was generated
}

// Quit is a special command that signals the application to exit.
func Quit() Msg {
	return msgQuit{}
}

// msgBatch is an internal message used to perform a bunch of commands. You can
// send a msgBatch with Batch.
type msgBatch []Cmd
