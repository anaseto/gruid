package js

import (
	"image"
	"time"
	"unicode/utf8"

	"syscall/js"

	"github.com/anaseto/gruid"
)

// TileManager manages tiles fetching.
type TileManager interface {
	// GetImage returns the image to be used for a given cell style.
	GetImage(gruid.Cell) *image.RGBA

	// TileSize returns the (width, height) in pixels of the tiles. Both
	// should be positive and non-zero.
	TileSize() (int, int)
}

// Driver implements gruid.Driver using the syscall/js interface for
// the browser using javascript and wasm.
type Driver struct {
	TileManager TileManager // for retrieving tiles
	Width       int         // initial screen width in cells
	Height      int         // initial screen height in celles

	display   js.Value
	ctx       js.Value
	cache     map[gruid.Cell]js.Value
	tw        int
	th        int
	mousepos  gruid.Position
	frame     gruid.Frame
	msgs      chan gruid.Msg
	flushdone chan bool
	mousedrag int
	listeners listeners
}

type listeners struct {
	keydown   js.Func
	mousemove js.Func
	mousedown js.Func
	mouseup   js.Func
	menu      js.Func
	wheel     js.Func
}

// Init implements gruid.Driver.Init.
func (dr *Driver) Init() error {
	dr.mousedrag = -1
	dr.msgs = make(chan gruid.Msg, 5)
	dr.flushdone = make(chan bool)
	canvas := js.Global().Get("document").Call("getElementById", "appcanvas")

	dr.listeners.menu = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		e.Call("preventDefault")
		return nil
	})
	canvas.Call("addEventListener", "contextmenu", dr.listeners.menu, false)
	canvas.Call("setAttribute", "tabindex", "1")
	dr.ctx = canvas.Call("getContext", "2d")
	dr.ctx.Set("imageSmoothingEnabled", false)
	dr.tw, dr.th = dr.TileManager.TileSize()
	canvas.Set("height", dr.th*dr.Height)
	canvas.Set("width", dr.tw*dr.Width)
	dr.cache = make(map[gruid.Cell]js.Value)

	appdiv := js.Global().Get("document").Call("getElementById", "appdiv")
	dr.listeners.keydown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		if !e.Get("ctrlKey").Bool() && !e.Get("metaKey").Bool() {
			e.Call("preventDefault")
		}
		s := e.Get("key").String()
		if s == "F11" {
			screenfull := js.Global().Get("screenfull")
			if screenfull.Truthy() && screenfull.Get("enabled").Bool() {
				screenfull.Call("toggle", appdiv)
			}
			return nil
		}
		code := e.Get("code").String()
		if s == "Unidentified" {
			s = code
		}
		if len(dr.msgs) < cap(dr.msgs) {
			if msg, ok := getMsgKeyDown(s, code); ok {
				if e.Get("shiftKey").Bool() {
					msg.Mod |= gruid.ModShift
				}
				if e.Get("ctrlKey").Bool() {
					msg.Mod |= gruid.ModCtrl
				}
				if e.Get("altKey").Bool() {
					msg.Mod |= gruid.ModAlt
				}
				if e.Get("metaKey").Bool() {
					msg.Mod |= gruid.ModMeta
				}
				dr.msgs <- msg
			}
		}
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "keydown", dr.listeners.keydown)
	dr.listeners.mousedown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		pos := dr.getMousePos(e)
		if dr.mousedrag >= 0 {
			return nil
		}
		n := e.Get("button").Int()
		if len(dr.msgs) < cap(dr.msgs) {
			switch n {
			case 0, 1, 2:
				dr.mousedrag = n
				dr.msgs <- gruid.MsgMouse{MousePos: pos, Action: gruid.MouseAction(n), Time: time.Now()}
			}
		}
		return nil
	})
	canvas.Call("addEventListener", "mousedown", dr.listeners.mousedown)
	dr.listeners.mouseup = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		pos := dr.getMousePos(e)
		n := e.Get("button").Int()
		if dr.mousedrag != n {
			return nil
		}
		dr.mousedrag = -1
		if len(dr.msgs) < cap(dr.msgs) {
			dr.msgs <- gruid.MsgMouse{MousePos: pos, Action: gruid.MouseAction(n), Time: time.Now()}
		}
		return nil
	})
	canvas.Call("addEventListener", "mouseup", dr.listeners.mouseup)
	dr.listeners.mousemove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		pos := dr.getMousePos(e)
		if pos.X != dr.mousepos.X || pos.Y != dr.mousepos.Y {
			dr.mousepos.X = pos.X
			dr.mousepos.Y = pos.Y
			if len(dr.msgs) < cap(dr.msgs) {
				dr.msgs <- gruid.MsgMouse{Action: gruid.MouseMove, MousePos: pos, Time: time.Now()}
			}
		}
		return nil
	})
	canvas.Call("addEventListener", "mousemove", dr.listeners.mousemove)
	dr.listeners.wheel = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		pos := dr.getMousePos(e)
		var action gruid.MouseAction
		delta := e.Get("deltaY").Float()
		if delta > 0 {
			action = gruid.MouseWheelUp
		} else if delta < 0 {
			action = gruid.MouseWheelDown
		} else {
			return nil
		}
		if len(dr.msgs) < cap(dr.msgs) {
			dr.msgs <- gruid.MsgMouse{Action: action, MousePos: pos, Time: time.Now()}
		}
		return nil
	})
	canvas.Call("addEventListener", "onwheel", dr.listeners.wheel)
	return nil
}

func (dr *Driver) getMousePos(evt js.Value) gruid.Position {
	canvas := js.Global().Get("document").Call("getElementById", "appcanvas")
	rect := canvas.Call("getBoundingClientRect")
	scaleX := canvas.Get("width").Float() / rect.Get("width").Float()
	scaleY := canvas.Get("height").Float() / rect.Get("height").Float()
	x := (evt.Get("clientX").Float() - rect.Get("left").Float()) * scaleX
	y := (evt.Get("clientY").Float() - rect.Get("top").Float()) * scaleY
	return gruid.Position{X: (int(x) - 1) / dr.tw, Y: (int(y) - 1) / dr.th}
}

func getMsgKeyDown(s, code string) (gruid.MsgKeyDown, bool) {
	if code == "Numpad5" && s != "5" {
		s = "Enter"
	}
	var key gruid.Key
	switch s {
	case "ArrowDown":
		key = gruid.KeyArrowDown
	case "ArrowLeft":
		key = gruid.KeyArrowLeft
	case "ArrowRight":
		key = gruid.KeyArrowRight
	case "ArrowUp":
		key = gruid.KeyArrowUp
	case "BackSpace":
		key = gruid.KeyBackspace
	case "Delete":
		key = gruid.KeyDelete
	case "End":
		key = gruid.KeyEnd
	case "Enter":
		key = gruid.KeyEnter
	case "Escape":
		key = gruid.KeyEscape
	case "Home":
		key = gruid.KeyHome
	case "Insert":
		key = gruid.KeyInsert
	case "PageUp":
		key = gruid.KeyPageUp
	case "PageDown":
		key = gruid.KeyPageDown
	case " ":
		key = gruid.KeySpace
	case "Tab":
		key = gruid.KeyTab
	default:
		if utf8.RuneCountInString(s) != 1 {
			return gruid.MsgKeyDown{}, false
		}
		key = gruid.Key(s)
	}
	return gruid.MsgKeyDown{Key: key, Time: time.Now()}, true
}

// PollMsg implements gruid.Driver.PollMsg.
func (dr *Driver) PollMsg() (gruid.Msg, error) {
	msg, ok := <-dr.msgs
	if ok {
		return msg, nil
	}
	return nil, nil
}

// Flush implements gruid.Driver.Flush.
func (dr *Driver) Flush(frame gruid.Frame) {
	dr.frame = frame
	js.Global().Get("window").Call("requestAnimationFrame",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} { dr.flushCallback(); return nil }))
	<-dr.flushdone
}

func (dr *Driver) flushCallback() {
	for _, cdraw := range dr.frame.Cells {
		cell := cdraw.Cell
		dr.draw(cell, cdraw.Pos.X, cdraw.Pos.Y)
	}
	dr.flushdone <- true
}

func (dr *Driver) draw(cell gruid.Cell, x, y int) {
	var canvas js.Value
	if cv, ok := dr.cache[cell]; ok {
		canvas = cv
	} else {
		canvas = js.Global().Get("document").Call("createElement", "canvas")
		canvas.Set("width", dr.tw)
		canvas.Set("height", dr.th)
		ctx := canvas.Call("getContext", "2d")
		ctx.Set("imageSmoothingEnabled", false)
		buf := dr.TileManager.GetImage(cell).Pix // TODO: do something if image is nil?
		ua := js.Global().Get("Uint8Array").New(js.ValueOf(len(buf)))
		js.CopyBytesToJS(ua, buf)
		ca := js.Global().Get("Uint8ClampedArray").New(ua)
		imgdata := js.Global().Get("ImageData").New(ca, dr.tw, dr.th)
		ctx.Call("putImageData", imgdata, 0, 0)
		dr.cache[cell] = canvas
	}
	dr.ctx.Call("drawImage", canvas, x*dr.tw, dr.th*y)
}

// Close implements gruid.Driver.Close. It releases some resources, such as
// event listeners.
func (dr *Driver) Close() {
	canvas := js.Global().Get("document").Call("getElementById", "appcanvas")
	canvas.Call("removeEventListener", "contextmenu", dr.listeners.menu, false)
	js.Global().Get("document").Call("removeEventListener", "keydown", dr.listeners.keydown)
	canvas.Call("removeEventListener", "mousedown", dr.listeners.mousedown)
	canvas.Call("removeEventListener", "mouseup", dr.listeners.mouseup)
	canvas.Call("removeEventListener", "mousemove", dr.listeners.mousemove)
	canvas.Call("removeEventListener", "onwheel", dr.listeners.wheel)
	dr.listeners.menu.Release()
	dr.listeners.keydown.Release()
	dr.listeners.mousedown.Release()
	dr.listeners.mouseup.Release()
	dr.listeners.mousemove.Release()
	dr.listeners.wheel.Release()
	dr.cache = nil
	dr.frame = gruid.Frame{}
	close(dr.msgs)
}

// ClearCache clears the tiles internal cache.
func (dr *Driver) ClearCache() {
	for c, _ := range dr.cache {
		delete(dr.cache, c)
	}
}
