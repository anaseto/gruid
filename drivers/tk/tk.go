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
	GetImage(gorltk.GridCell) *image.RGBA
}

type Driver struct {
	TileManager TileManager
	ir          *gothic.Interpreter
	cache       map[gorltk.GridCell]*image.RGBA
	Width       int
	Height      int
	width       int
	height      int
	mousepos    gorltk.Position
	canvas      *image.RGBA
}

func (tk *Driver) Init() error {
	tk.canvas = image.NewRGBA(image.Rect(0, 0, tk.Width*16, tk.Height*24))
	tk.ir = gothic.NewInterpreter(fmt.Sprintf(`
wm title . "Harmonist Tk"
wm resizable . 0 0
set width [expr {16 * %d}]
set height [expr {24 * %d}]
wm geometry . =${width}x$height
set can [canvas .c -width $width -height $height -background #002b36]
grid $can -row 0 -column 0
focus $can
image create photo gamescreen -width $width -height $height -palette 256/256/256
image create photo bufscreen -width $width -height $height -palette 256/256/256
$can create image 0 0 -anchor nw -image gamescreen
`, tk.Width, tk.Height))
	tk.initElements()
	tk.ir.RegisterCommand("GetKey", func(c, keysym string) {
		var s string
		if c != "" {
			s = c
		} else {
			s = keysym
		}
		if len(eventCh) < cap(eventCh) {
			eventCh <- gorltk.EventKeyDown{Key: s, Time: time.Now()}
		}
	})
	tk.ir.RegisterCommand("MouseDown", func(x, y, n int) {
		if len(eventCh) < cap(eventCh) {
			eventCh <- gorltk.EventMouseDown{MousePos: gorltk.Position{X: (x - 1) / tk.width, Y: (y - 1) / tk.height}, Button: gorltk.MouseButton(n - 1), Time: time.Now()}
		}
	})
	tk.ir.RegisterCommand("MouseMotion", func(x, y int) {
		nx := (x - 1) / tk.width
		ny := (y - 1) / tk.height
		if nx != tk.mousepos.X || ny != tk.mousepos.Y {
			if len(eventCh) < cap(eventCh) {
				tk.mousepos.X = nx
				tk.mousepos.Y = ny
				eventCh <- gorltk.EventMouseMove{MousePos: gorltk.Position{X: nx, Y: ny}, Time: time.Now()}
			}
		}
	})
	tk.ir.RegisterCommand("OnClosing", func() {
		if len(eventCh) < cap(eventCh) {
			// TODO: make it not depend on a default string and normal mode
			eventCh <- gorltk.EventKeyDown{Key: "S", Time: time.Now()}
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

	//SolarizedPalette()
	//settingsActions = append(settingsActions, toggleTiles)
	//GameConfig.Tiles = true
	return nil
}

func (tk *Driver) initElements() error {
	tk.width = 16
	tk.height = 24
	tk.cache = make(map[gorltk.GridCell]*image.RGBA)
	return nil
}

var eventCh chan gorltk.Event

func init() {
	eventCh = make(chan gorltk.Event, 5)
}

func (tk *Driver) Interrupt() {
	eventCh <- gorltk.EventInterrupt{Time: time.Now()}
}

func (tk *Driver) Close() {
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
	subimg := tk.canvas.SubImage(image.Rect(xmin*16, ymin*24, (xmax+1)*16, (ymax+1)*24))
	png.Encode(pngbuf, subimg)
	png := base64.StdEncoding.EncodeToString(pngbuf.Bytes())
	tk.ir.Eval("gamescreen put %{0%s} -format png -to %{1%d} %{2%d} %{3%d} %{4%d}", png,
		xmin*16, ymin*24, (xmax+1)*16, (ymax+1)*24) // TODO: optimize this more
}

func (tk *Driver) draw(gd *gorltk.Grid, cs gorltk.GridCell, x, y int) {
	var img *image.RGBA
	if im, ok := tk.cache[cs]; ok {
		img = im
	} else {
		img = tk.TileManager.GetImage(cs)
		tk.cache[cs] = img
	}
	draw.Draw(tk.canvas, image.Rect(x*tk.width, tk.height*y, (x+1)*tk.width, (y+1)*tk.height), img, image.Point{0, 0}, draw.Over)
}

func (tk *Driver) ClearCache() {
	for c, _ := range tk.cache {
		//if c.Style == StyleMap {
		delete(tk.cache, c)
		//}
	}
}

func (tk *Driver) PollEvent() (ev gorltk.Event) {
	ev = <-eventCh
	switch ev := ev.(type) {
	case gorltk.EventKeyDown:
		switch ev.Key {
		case "KP_Enter", "Return", "\r", "\n":
			ev.Key = "."
		case "Left", "KP_Left":
			ev.Key = "4"
		case "Right", "KP_Right":
			ev.Key = "6"
		case "Up", "KP_Up", "BackSpace":
			ev.Key = "8"
		case "Down", "KP_Down":
			ev.Key = "2"
		case "KP_Prior", "Prior":
			ev.Key = "u"
		case "KP_Next", "Next":
			ev.Key = "d"
		case "KP_Begin", "KP_Delete":
			ev.Key = "5"
		default:
			if utf8.RuneCountInString(ev.Key) != 1 {
				ev.Key = ""
			}
		}
		return ev
	}
	return ev
}
