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

// MouseButton represents mouse buttons.
type MouseButton int

// This is the list of supported mouse buttons.
const (
	ButtonMain      MouseButton = iota // left button
	ButtonAuxiliary                    // middle button
	ButtonSecondary                    // right button
)

// MsgMouseDown represents a mouse click.
type MsgMouseDown struct {
	Button   MouseButton // which button was pressed
	MousePos Position    // mouse position in the grid
	Time     time.Time   // time when the event was generated
}

// MsgMouseMove represents a mouse motion.
type MsgMouseMove struct {
	MousePos Position  // mouse position in the grid
	Time     time.Time // time when the event was generated
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
