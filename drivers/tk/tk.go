package tk

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"time"
	"unicode/utf8"

	"github.com/anaseto/gorltk"
	"github.com/nsf/gothic"
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
	Height      int // initial screen height in cells
	ir          *gothic.Interpreter
	cache       map[gorltk.Cell]*image.RGBA
	tw          int
	th          int
	mousepos    gorltk.Position
	canvas      *image.RGBA
	msgs        chan gorltk.Msg
	interrupt   chan bool
}

func (tk *Driver) Init() error {
	tk.msgs = make(chan gorltk.Msg, 5)
	tk.interrupt = make(chan bool)
	tk.tw, tk.th = tk.TileManager.TileSize()
	tk.cache = make(map[gorltk.Cell]*image.RGBA)
	tk.canvas = image.NewRGBA(image.Rect(0, 0, tk.Width*tk.tw, tk.Height*tk.th))
	tk.ir = gothic.NewInterpreter(fmt.Sprintf(`
wm title . "gorltk Tk"
wm resizable . 0 0
set width [expr {%d * %d}]
set height [expr {%d * %d}]
wm geometry . =${width}x$height
set can [canvas .c -width $width -height $height -background #002b36]
grid $can -row 0 -column 0
focus $can
image create photo gamescreen -width $width -height $height -palette 256/256/256
image create photo bufscreen -width $width -height $height -palette 256/256/256
$can create image 0 0 -anchor nw -image gamescreen
`, tk.tw, tk.Width, tk.th, tk.Height))
	tk.ir.RegisterCommand("GetKey", func(c, keysym string) {
		var s string
		if c != "" {
			s = c
		} else {
			s = keysym
		}
		if len(tk.msgs) < cap(tk.msgs) {
			if msg, ok := getMsgKeyDown(s); ok {
				tk.msgs <- msg
			}
		}
	})
	tk.ir.RegisterCommand("MouseDown", func(x, y, n int) {
		if len(tk.msgs) < cap(tk.msgs) {
			tk.msgs <- gorltk.MsgMouseDown{MousePos: gorltk.Position{X: (x - 1) / tk.tw, Y: (y - 1) / tk.th}, Button: gorltk.MouseButton(n - 1), Time: time.Now()}
		}
	})
	tk.ir.RegisterCommand("MouseMotion", func(x, y int) {
		nx := (x - 1) / tk.tw
		ny := (y - 1) / tk.th
		if nx != tk.mousepos.X || ny != tk.mousepos.Y {
			if len(tk.msgs) < cap(tk.msgs) {
				tk.mousepos.X = nx
				tk.mousepos.Y = ny
				tk.msgs <- gorltk.MsgMouseMove{MousePos: gorltk.Position{X: nx, Y: ny}, Time: time.Now()}
			}
		}
	})
	tk.ir.RegisterCommand("OnClosing", func() {
		if len(tk.msgs) < cap(tk.msgs) {
			tk.msgs <- gorltk.Quit()
		}
	})
	tk.ir.Eval(`
bind .c <Key> {
	GetKey %A %K
}
bind .c <Motion> {
	MouseMotion %x %y
}
bind .c <ButtonPress> {
	MouseDown %x %y %b
}
wm protocol . WM_DELETE_WINDOW OnClosing
`)

	return nil
}

func getMsgKeyDown(s string) (gorltk.Msg, bool) {
	var key gorltk.Key
	switch s {
	case "Down", "KP_2":
		key = gorltk.KeyArrowDown
	case "Left", "KP_4":
		key = gorltk.KeyArrowLeft
	case "Right", "KP_6":
		key = gorltk.KeyArrowRight
	case "Up", "KP_8":
		key = gorltk.KeyArrowUp
	case "BackSpace":
		key = gorltk.KeyBackspace
	case "Delete", "KP_7":
		key = gorltk.KeyDelete
	case "End", "KP_1":
		key = gorltk.KeyEnd
	case "KP_Enter", "Return", "KP_5":
		key = gorltk.KeyEnter
	case "Escape":
		key = gorltk.KeyEscape
	case "Home":
		key = gorltk.KeyHome
	case "Insert":
		key = gorltk.KeyInsert
	case "KP_9", "Prior":
		key = gorltk.KeyPageUp
	case "KP_3", "Next":
		key = gorltk.KeyPageDown
	case "space":
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

func (tk *Driver) PollMsg() (gorltk.Msg, bool) {
	select {
	case msg := <-tk.msgs:
		return msg, true
	case <-tk.interrupt:
		return nil, false
	}
}

type rectangle struct {
	xmin, xmax, ymin, ymax int
}

func (tk *Driver) Flush(gd *gorltk.Grid) {
	w, h := gd.Size()
	rects := []rectangle{}
	r := rectangle{w - 1, 0, h - 1, 0}
	n := 0
	for _, cdraw := range gd.Frame().Cells {
		cs := cdraw.Cell
		x, y := cdraw.Pos.X, cdraw.Pos.Y
		tk.draw(gd, cs, x, y)
		if n > 0 && y-r.ymax > 2 {
			rects = append(rects, r)
			n = 0
			r = rectangle{w - 1, 0, h - 1, y}
		}
		if x < r.xmin {
			r.xmin = x
		}
		if x > r.xmax {
			r.xmax = x
		}
		if y < r.ymin {
			r.ymin = y
		}
		if y > r.ymax {
			r.ymax = y
		}
		n++
	}
	rects = append(rects, r)
	for _, r := range rects {
		tk.UpdateRectangle(r.xmin, r.ymin, r.xmax, r.ymax)
	}
}

func (tk *Driver) UpdateRectangle(xmin, ymin, xmax, ymax int) {
	if xmin > xmax || ymin > ymax {
		return
	}
	pngbuf := &bytes.Buffer{}
	subimg := tk.canvas.SubImage(image.Rect(xmin*tk.tw, ymin*tk.th, (xmax+1)*tk.tw, (ymax+1)*tk.th))
	png.Encode(pngbuf, subimg)
	png := base64.StdEncoding.EncodeToString(pngbuf.Bytes())
	tk.ir.Eval("gamescreen put %{0%s} -format png -to %{1%d} %{2%d} %{3%d} %{4%d}", png,
		xmin*tk.tw, ymin*tk.th, (xmax+1)*tk.tw, (ymax+1)*tk.th)
}

func (tk *Driver) draw(gd *gorltk.Grid, cs gorltk.Cell, x, y int) {
	var img *image.RGBA
	if im, ok := tk.cache[cs]; ok {
		img = im
	} else {
		img = tk.TileManager.GetImage(cs)
		tk.cache[cs] = img // TODO: do something if image is nil?
	}
	draw.Draw(tk.canvas, image.Rect(x*tk.tw, tk.th*y, (x+1)*tk.tw, (y+1)*tk.th), img, image.Point{0, 0}, draw.Over)
}

func (tk *Driver) Interrupt() {
	tk.interrupt <- true
}

func (tk *Driver) Close() {
	// do nothing
}

func (tk *Driver) ClearCache() {
	for c, _ := range tk.cache {
		delete(tk.cache, c)
	}
}
