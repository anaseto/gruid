package js

import (
	"image"
	"time"
	"unicode/utf8"

	"syscall/js"

	"github.com/anaseto/gruid"
)

type TileManager interface {
	// GetImage returns the image to be used for a given cell style.
	GetImage(gruid.Cell) *image.RGBA

	// TileSize returns the (width, height) in pixels of the tiles.
	TileSize() (int, int)
}

type Driver struct {
	TileManager TileManager
	Width       int // initial screen width in cells
	Height      int // initial screen height in celles
	display     js.Value
	ctx         js.Value
	cache       map[gruid.Cell]js.Value
	tw          int
	th          int
	mousepos    gruid.Position
	grid        gruid.Grid
	msgs        chan gruid.Msg
	flushdone   chan bool
	mousedrag   int
}

func (dr *Driver) Init() error {
	dr.mousedrag = -1
	dr.msgs = make(chan gruid.Msg, 5)
	dr.flushdone = make(chan bool)
	canvas := js.Global().Get("document").Call("getElementById", "appcanvas")
	canvas.Call("addEventListener", "contextmenu", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		e.Call("preventDefault")
		return nil
	}), false)
	canvas.Call("setAttribute", "tabindex", "1")
	dr.ctx = canvas.Call("getContext", "2d")
	dr.ctx.Set("imageSmoothingEnabled", false)
	dr.tw, dr.th = dr.TileManager.TileSize()
	canvas.Set("height", dr.th*dr.Height)
	canvas.Set("width", dr.tw*dr.Width)
	dr.cache = make(map[gruid.Cell]js.Value)

	appdiv := js.Global().Get("document").Call("getElementById", "appdiv")
	js.Global().Get("document").Call(
		"addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			e := args[0]
			if !e.Get("ctrlKey").Bool() && !e.Get("metaKey").Bool() {
				e.Call("preventDefault")
			} else {
				return nil
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
						msg.Shift = true
					}
					dr.msgs <- msg
				}
			}
			return nil
		}))
	canvas.Call(
		"addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
		}))
	canvas.Call(
		"addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
		}))
	canvas.Call(
		"addEventListener", "mousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
		}))
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

func (dr *Driver) PollMsg() gruid.Msg {
	msg := <-dr.msgs
	return msg
}

func (dr *Driver) Flush(gd gruid.Grid) {
	dr.grid = gd
	js.Global().Get("window").Call("requestAnimationFrame",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} { dr.flushCallback(); return nil }))
	<-dr.flushdone
}

func (dr *Driver) flushCallback() {
	for _, cdraw := range dr.grid.Frame().Cells {
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

func (dr *Driver) Close() {
	dr.grid = gruid.Grid{} // release grid resource
}

func (dr *Driver) ClearCache() {
	for c, _ := range dr.cache {
		delete(dr.cache, c)
	}
}
