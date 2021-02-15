// Package js provides a Driver for making browser apps using wasm.
package js

import (
	"context"
	"image"
	"image/draw"
	"log"
	"time"
	"unicode/utf8"

	"syscall/js"

	"github.com/anaseto/gruid"
)

// TileManager manages tiles fetching.
type TileManager interface {
	// GetImage returns the image to be used for a given cell style.
	GetImage(gruid.Cell) image.Image

	// TileSize returns the (width, height) in pixels of the tiles. Both
	// should be positive and non-zero.
	TileSize() gruid.Point
}

// Driver implements gruid.Driver using the syscall/js interface for
// the browser using javascript and wasm.
type Driver struct {
	tm     TileManager
	width  int
	height int

	ctx       js.Value
	cache     map[gruid.Cell]js.Value
	tw        int
	th        int
	mousepos  gruid.Point
	mousedrag int
	listeners listeners
	init      bool
	flushing  bool
	fcs       []gruid.FrameCell
	appcanvas string
	appdiv    string
}

// Config contains configurations options for the driver.
type Config struct {
	TileManager TileManager // for retrieving tiles (required)
	Width       int         // initial screen width in cells (default: 80)
	Height      int         // initial screen height in cells (default: 24)
	AppCanvasId string      // application's canvas id (default: appcanvas)
	AppDivId    string      // application's div containing the canvas id (default: appdiv)
}

// NewDriver returns a new driver with given configuration options.
func NewDriver(cfg Config) *Driver {
	dr := &Driver{}
	dr.width = cfg.Width
	if dr.width <= 0 {
		dr.width = 80
	}
	dr.height = cfg.Height
	if dr.height <= 0 {
		dr.height = 24
	}
	dr.appcanvas = cfg.AppCanvasId
	if dr.appcanvas == "" {
		dr.appcanvas = "appcanvas"
	}
	dr.appdiv = cfg.AppDivId
	if dr.appdiv == "" {
		dr.appdiv = "appdiv"
	}
	canvas := js.Global().Get("document").Call("getElementById", dr.appcanvas)
	canvas.Call("setAttribute", "tabindex", "1")
	dr.ctx = canvas.Call("getContext", "2d")
	dr.ctx.Set("imageSmoothingEnabled", false)
	dr.SetTileManager(cfg.TileManager)
	return dr
}

// SetTileManager allows to change the used tile manager.
func (dr *Driver) SetTileManager(tm TileManager) {
	dr.tm = tm
	p := tm.TileSize()
	dr.tw, dr.th = p.X, p.Y
	if dr.tw <= 0 {
		dr.tw = 1
	}
	if dr.th <= 0 {
		dr.th = 1
	}
	dr.resizeCanvas()
	if dr.init {
		dr.ClearCache()
	}
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
	dr.cache = make(map[gruid.Cell]js.Value)
	if dr.init {
		dr.resizeCanvas()
	}
	dr.init = true
	dr.mousedrag = -1
	dr.flushing = false
	return nil
}

func (dr *Driver) resizeCanvas() {
	canvas := js.Global().Get("document").Call("getElementById", dr.appcanvas)
	canvas.Set("height", dr.th*dr.height)
	canvas.Set("width", dr.tw*dr.width)
}

func (dr *Driver) getMousePos(evt js.Value) gruid.Point {
	canvas := js.Global().Get("document").Call("getElementById", dr.appcanvas)
	rect := canvas.Call("getBoundingClientRect")
	scaleX := canvas.Get("width").Float() / rect.Get("width").Float()
	scaleY := canvas.Get("height").Float() / rect.Get("height").Float()
	x := (evt.Get("clientX").Float() - rect.Get("left").Float()) * scaleX
	y := (evt.Get("clientY").Float() - rect.Get("top").Float()) * scaleY
	return gruid.Point{X: (int(x) - 1) / dr.tw, Y: (int(y) - 1) / dr.th}
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

// PollMsgs implements gruid.Driver.PollMsgs. To avoid conflicts with browser
// or OS shortcuts, keydown events with modifier keys may not be reported. If
// the script screenfull is available, it binds F11 for a portable canvas
// fullscreen.
func (dr *Driver) PollMsgs(ctx context.Context, msgs chan<- gruid.Msg) error {
	send := func(msg gruid.Msg) {
		select {
		case msgs <- msg:
		case <-ctx.Done():
		}
	}
	send(gruid.MsgScreen{Width: dr.width, Height: dr.height})
	canvas := js.Global().Get("document").Call("getElementById", dr.appcanvas)
	dr.listeners.menu = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		e.Call("preventDefault")
		return nil
	})
	canvas.Call("addEventListener", "contextmenu", dr.listeners.menu, false)
	appdiv := js.Global().Get("document").Call("getElementById", dr.appdiv)
	dr.listeners.keydown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		if !e.Get("ctrlKey").Bool() && !e.Get("metaKey").Bool() && !e.Get("altKey").Bool() {
			e.Call("preventDefault")
		} else if e.Get("ctrlKey").Bool() && e.Get("metaKey").Bool() {
			// There are too many conflicts with browser or OS
			// shortcuts for it to be worth the the trouble.
			return nil
		}
		s := e.Get("key").String()
		if s == "F11" {
			// use portable fullscreen provided by screenfull if available.
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
			send(msg)
		}
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "keydown", dr.listeners.keydown)
	dr.listeners.mousedown = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		if e.Get("ctrlKey").Bool() || e.Get("metaKey").Bool() || e.Get("shiftKey").Bool() {
			// do not report mouse combined with special keys
			return nil
		}
		p := dr.getMousePos(e)
		if dr.mousedrag >= 0 {
			return nil
		}
		n := e.Get("button").Int()
		switch n {
		case 0, 1, 2:
			dr.mousedrag = n
			send(gruid.MsgMouse{P: p, Action: gruid.MouseAction(n), Time: time.Now()})
		}
		return nil
	})
	canvas.Call("addEventListener", "mousedown", dr.listeners.mousedown)
	dr.listeners.mouseup = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		if e.Get("ctrlKey").Bool() || e.Get("metaKey").Bool() || e.Get("shiftKey").Bool() {
			// do not report mouse combined with special keys
			return nil
		}
		p := dr.getMousePos(e)
		n := e.Get("button").Int()
		if dr.mousedrag != n {
			return nil
		}
		dr.mousedrag = -1
		send(gruid.MsgMouse{P: p, Action: gruid.MouseRelease, Time: time.Now()})
		return nil
	})
	canvas.Call("addEventListener", "mouseup", dr.listeners.mouseup)
	dr.listeners.mousemove = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		p := dr.getMousePos(e)
		if p.X != dr.mousepos.X || p.Y != dr.mousepos.Y {
			dr.mousepos.X = p.X
			dr.mousepos.Y = p.Y
			send(gruid.MsgMouse{Action: gruid.MouseMove, P: p, Time: time.Now()})
		}
		return nil
	})
	canvas.Call("addEventListener", "mousemove", dr.listeners.mousemove)
	dr.listeners.wheel = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		p := dr.getMousePos(e)
		var action gruid.MouseAction
		delta := e.Get("deltaY").Float()
		if delta > 0 {
			action = gruid.MouseWheelUp
		} else if delta < 0 {
			action = gruid.MouseWheelDown
		} else {
			return nil
		}
		send(gruid.MsgMouse{Action: action, P: p, Time: time.Now()})
		return nil
	})
	canvas.Call("addEventListener", "onwheel", dr.listeners.wheel)
	<-ctx.Done()
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
	return nil
}

// Flush implements gruid.Driver.Flush.
func (dr *Driver) Flush(frame gruid.Frame) {
	var cached bool
	if frame.Width != dr.width || frame.Height != dr.height {
		dr.width = frame.Width
		dr.height = frame.Height
		dr.resizeCanvas()
	}
	if dr.flushing {
		cells := make([]gruid.FrameCell, len(frame.Cells))
		for i, fc := range frame.Cells {
			cells[i] = fc
		}
		frame.Cells = cells
	} else {
		// avoid allocation in the common case
		dr.fcs = dr.fcs[:0]
		for _, fc := range frame.Cells {
			dr.fcs = append(dr.fcs, fc)
		}
		frame.Cells = dr.fcs
		cached = true
		dr.flushing = true
	}
	js.Global().Get("window").Call("requestAnimationFrame",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} { dr.flushCallback(frame, cached); return nil }))
}

func (dr *Driver) flushCallback(frame gruid.Frame, cached bool) {
	for _, fc := range frame.Cells {
		cell := fc.Cell
		dr.draw(cell, fc.P.X, fc.P.Y)
	}
	if cached {
		dr.flushing = false
	}
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
		img := dr.tm.GetImage(cell)
		if img == nil {
			log.Printf("no tile for %+v", cell)
			return
		}
		var rgbaimg *image.RGBA
		switch img := img.(type) {
		case *image.RGBA:
			rgbaimg = img
		default:
			rect := img.Bounds()
			rgbaimg = image.NewRGBA(rect)
			draw.Draw(rgbaimg, rect, img, rect.Min, draw.Src)
		}
		buf := rgbaimg.Pix
		ua := js.Global().Get("Uint8Array").New(js.ValueOf(len(buf)))
		js.CopyBytesToJS(ua, buf)
		ca := js.Global().Get("Uint8ClampedArray").New(ua)
		imgdata := js.Global().Get("ImageData").New(ca, dr.tw, dr.th)
		ctx.Call("putImageData", imgdata, 0, 0)
		dr.cache[cell] = canvas
	}
	dr.ctx.Call("drawImage", canvas, x*dr.tw, dr.th*y)
}

// Close implements gruid.Driver.Close.
func (dr *Driver) Close() {
	if !dr.init {
		return
	}
	dr.cache = nil
	dr.init = false
}

// ClearCache clears the tiles internal cache.
func (dr *Driver) ClearCache() {
	for c := range dr.cache {
		delete(dr.cache, c)
	}
}
