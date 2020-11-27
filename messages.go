package gorltk

import "time"

type Key string

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
	shift bool      //
	Time  time.Time // time when the event was generated
}

// ShiftKey indicates if the shift key was pressed when the event occured. This
// may be not supported equally well accross all platforms.
func (ev MsgKeyDown) ShiftKey() bool {
	return ev.shift
}

type MouseButton int

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

// MsgInterrupt represents a wakeup. It can be used to end prematurely a
// PollEvent call, for example to signal the end of an animation.
type MsgInterrupt struct {
	Time time.Time // time when the event was generated
}

type msgQuit struct{}

func Quit() Msg {
	return msgQuit{}
}
