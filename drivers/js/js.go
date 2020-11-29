package js

import (
	"image"
	"time"
	"unicode/utf8"

	"syscall/js"

	"github.com/anaseto/gorltk"
)

type TileManager interface {
	// GetImage returns the image to be used for a given cell style.
	GetImage(gorltk.Cell) *image.RGBA

	// TileSize returns the (width, height) in pixels of the tiles.
	TileSize() (int, int)
}

type Driver struct {
	TileManager TileManager
	Width       int // initial screen width in cells
	Height      int // initial screen height in celles
	display     js.Value
	ctx         js.Value
	cache       map[gorltk.Cell]js.Value
	tw          int
	th          int
	mousepos    gorltk.Position
	grid        *gorltk.Grid
	msgs        chan gorltk.Msg
	interrupt   chan bool
	flushdone   chan bool
}

func (dr *Driver) Init() error {
	dr.msgs = make(chan gorltk.Msg, 5)
	dr.interrupt = make(chan bool)
	dr.flushdone = make(chan bool)
	canvas := js.Global().Get("document").Call("getElementById", "gamecanvas")
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
	dr.cache = make(map[gorltk.Cell]js.Value)

	gamediv := js.Global().Get("document").Call("getElementById", "gamediv")
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
					screenfull.Call("toggle", gamediv)
				}
				return nil
			}
			code := e.Get("code").String()
			if s == "Unidentified" {
				s = code
			}
			if len(dr.msgs) < cap(dr.msgs) {
				if msg, ok := getMsgKeyDown(s, code); ok {
					dr.msgs <- msg
				}
			}
			return nil
		}))
	canvas.Call(
		"addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			e := args[0]
			pos := dr.getMousePos(e)
			if len(dr.msgs) < cap(dr.msgs) {
				dr.msgs <- gorltk.MsgMouseDown{MousePos: pos, Button: gorltk.MouseButton(e.Get("button").Int()), Time: time.Now()}
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
					dr.msgs <- gorltk.MsgMouseMove{MousePos: pos, Time: time.Now()}
				}
			}
			return nil
		}))
	return nil
}

func (dr *Driver) getMousePos(evt js.Value) gorltk.Position {
	canvas := js.Global().Get("document").Call("getElementById", "gamecanvas")
	rect := canvas.Call("getBoundingClientRect")
	scaleX := canvas.Get("width").Float() / rect.Get("width").Float()
	scaleY := canvas.Get("height").Float() / rect.Get("height").Float()
	x := (evt.Get("clientX").Float() - rect.Get("left").Float()) * scaleX
	y := (evt.Get("clientY").Float() - rect.Get("top").Float()) * scaleY
	return gorltk.Position{X: (int(x) - 1) / dr.tw, Y: (int(y) - 1) / dr.th}
}

func getMsgKeyDown(s, code string) (gorltk.Msg, bool) {
	if code == "Numpad5" && s != "5" {
		s = "Enter"
	}
	var key gorltk.Key
	switch s {
	case "ArrowDown":
		key = gorltk.KeyArrowDown
	case "ArrowLeft":
		key = gorltk.KeyArrowLeft
	case "ArrowRight":
		key = gorltk.KeyArrowRight
	case "ArrowUp":
		key = gorltk.KeyArrowUp
	case "BackSpace":
		key = gorltk.KeyBackspace
	case "Delete":
		key = gorltk.KeyDelete
	case "End":
		key = gorltk.KeyEnd
	case "Enter":
		key = gorltk.KeyEnter
	case "Escape":
		key = gorltk.KeyEscape
	case "Home":
		key = gorltk.KeyHome
	case "Insert":
		key = gorltk.KeyInsert
	case "PageUp":
		key = gorltk.KeyPageUp
	case "PageDown":
		key = gorltk.KeyPageDown
	case " ":
		key = gorltk.KeySpace
	case "Tab":
		key = gorltk.KeyTab
	default:
		if utf8.RuneCountInString(s) != 1 {
			return "", false
		}
		key = gorltk.Key(s)
	}
	return gorltk.MsgKeyDown{Key: key, Time: time.Now()}, true
}

func (dr *Driver) PollMsg() (gorltk.Msg, bool) {
	select {
	case msg := <-dr.msgs:
		return msg, true
	case <-dr.interrupt:
		return nil, false
	}
}

func (dr *Driver) Flush(gd *gorltk.Grid) {
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

func (dr *Driver) draw(cell gorltk.Cell, x, y int) {
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

func (dr *Driver) Interrupt() {
	dr.interrupt <- true
}

func (dr *Driver) Close() {
	dr.grid = nil // release grid resource
}

func (dr *Driver) ClearCache() {
	for c, _ := range dr.cache {
		delete(dr.cache, c)
	}
}
