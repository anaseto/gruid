package tk

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	//"log"
	"time"
	"unicode/utf8"

	"github.com/anaseto/gruid"
	"github.com/nsf/gothic"
)

type TileManager interface {
	// GetImage returns the image to be used for a given cell style.
	GetImage(gruid.Cell) *image.RGBA

	// TileSize returns the (width, height) in pixels of the tiles.
	TileSize() (int, int)
}

type Driver struct {
	TileManager TileManager // for retrieving tiles
	Width       int         // initial screen width in cells
	Height      int         // initial screen height in cells
	ir          *gothic.Interpreter
	cache       map[gruid.Cell]*image.RGBA
	tw          int
	th          int
	mousepos    gruid.Position
	canvas      *image.RGBA
	msgs        chan gruid.Msg
	mousedrag   int
}

func (tk *Driver) Init() error {
	tk.msgs = make(chan gruid.Msg, 5)
	tk.tw, tk.th = tk.TileManager.TileSize()
	tk.cache = make(map[gruid.Cell]*image.RGBA)
	tk.canvas = image.NewRGBA(image.Rect(0, 0, tk.Width*tk.tw, tk.Height*tk.th))
	tk.ir = gothic.NewInterpreter(fmt.Sprintf(`
wm title . "gruid Tk"
wm resizable . 0 0
set width [expr {%d * %d}]
set height [expr {%d * %d}]
wm geometry . =${width}x$height
set can [canvas .c -width $width -height $height -background #002b36]
grid $can -row 0 -column 0
focus $can
image create photo appscreen -width $width -height $height -palette 256/256/256
image create photo bufscreen -width $width -height $height -palette 256/256/256
$can create image 0 0 -anchor nw -image appscreen
`, tk.tw, tk.Width, tk.th, tk.Height))
	tk.ir.RegisterCommand("GetKey", func(c, keysym, mod string) {
		var s string
		s = keysym
		if c != "" && s == "" {
			s = c
		}
		if s == "ISO_Left_Tab" {
			s = "Tab"
			mod = "Shift"
		}
		if len(tk.msgs) < cap(tk.msgs) {
			if msg, ok := getMsgKeyDown(s, c); ok {
				if mod == "Shift" {
					msg.Mod |= gruid.ModShift
				}
				if mod == "Control" {
					msg.Mod |= gruid.ModCtrl
				}
				if mod == "Alt" {
					msg.Mod |= gruid.ModAlt
				}
				tk.msgs <- msg
			}
		}
	})
	tk.ir.RegisterCommand("MouseDown", func(x, y, n int) {
		if len(tk.msgs) < cap(tk.msgs) {
			if tk.mousedrag > 0 {
				return
			}
			var action gruid.MouseAction
			switch n {
			case 1, 2, 3:
				action = gruid.MouseAction(n - 1)
				tk.mousedrag = n
				tk.msgs <- gruid.MsgMouse{MousePos: gruid.Position{X: (x - 1) / tk.tw, Y: (y - 1) / tk.th},
					Action: action, Time: time.Now()}
			}
		}
	})
	tk.ir.RegisterCommand("MouseRelease", func(x, y, n int) {
		if len(tk.msgs) < cap(tk.msgs) {
			if tk.mousedrag != n {
				return
			}
			tk.mousedrag = 0
			tk.msgs <- gruid.MsgMouse{MousePos: gruid.Position{X: (x - 1) / tk.tw, Y: (y - 1) / tk.th},
				Action: gruid.MouseRelease, Time: time.Now()}
		}
	})
	tk.ir.RegisterCommand("MouseWheel", func(x, y, delta int) {
		if len(tk.msgs) < cap(tk.msgs) {
			var action gruid.MouseAction
			if delta > 0 {
				action = gruid.MouseWheelUp
			} else if delta < 0 {
				action = gruid.MouseWheelDown
			} else {
				return
			}
			tk.msgs <- gruid.MsgMouse{MousePos: gruid.Position{X: (x - 1) / tk.tw, Y: (y - 1) / tk.th},
				Action: action, Time: time.Now()}
		}
	})
	tk.ir.RegisterCommand("MouseMotion", func(x, y int) {
		nx := (x - 1) / tk.tw
		ny := (y - 1) / tk.th
		if nx != tk.mousepos.X || ny != tk.mousepos.Y {
			if len(tk.msgs) < cap(tk.msgs) {
				tk.mousepos.X = nx
				tk.mousepos.Y = ny
				tk.msgs <- gruid.MsgMouse{MousePos: gruid.Position{X: nx, Y: ny},
					Action: gruid.MouseMove, Time: time.Now()}
			}
		}
	})
	tk.ir.RegisterCommand("OnClosing", func() {
		if len(tk.msgs) < cap(tk.msgs) {
			tk.msgs <- gruid.Quit()
		}
	})
	tk.ir.Eval(`
bind .c <Shift-Key> {
	GetKey %A %K Shift
}
bind .c <Control-Key> {
	GetKey %A %K Control
}
bind .c <Alt-Key> {
	GetKey %A %K Alt
}
bind .c <Key> {
	GetKey %A %K {}
}
bind .c <Motion> {
	MouseMotion %x %y
}
bind .c <ButtonPress> {
	MouseDown %x %y %b
}
bind .c <ButtonRelease> {
	MouseRelease %x %y %b
}
bind .c <MouseWheel> {
	MouseWheel %x %y %D
}
wm protocol . WM_DELETE_WINDOW OnClosing
`)

	return nil
}

func getMsgKeyDown(s, c string) (gruid.MsgKeyDown, bool) {
	var key gruid.Key
	//log.Printf("%#v, %#v", s, c)
	switch s {
	case "Down", "KP_Down":
		key = gruid.KeyArrowDown
	case "Left", "KP_Left":
		key = gruid.KeyArrowLeft
	case "Right", "KP_Right":
		key = gruid.KeyArrowRight
	case "Up", "KP_Up":
		key = gruid.KeyArrowUp
	case "BackSpace":
		key = gruid.KeyBackspace
	case "Delete":
		key = gruid.KeyDelete
	case "End", "KP_End":
		key = gruid.KeyEnd
	case "KP_Enter", "Return", "KP_Begin":
		key = gruid.KeyEnter
	case "Escape":
		key = gruid.KeyEscape
	case "Home", "KP_Home":
		key = gruid.KeyHome
	case "Insert":
		key = gruid.KeyInsert
	case "Prior", "KP_Prior":
		key = gruid.KeyPageUp
	case "Next", "KP_Next":
		key = gruid.KeyPageDown
	case "space":
		key = gruid.KeySpace
	case "Tab":
		key = gruid.KeyTab
	default:
		l := utf8.RuneCountInString(s)
		if l > 1 && c != "" {
			s = c
		}
		if utf8.RuneCountInString(s) != 1 {
			return gruid.MsgKeyDown{}, false
		}
		key = gruid.Key(s)
	}
	return gruid.MsgKeyDown{Key: key, Time: time.Now()}, true
}

func (tk *Driver) PollMsg() (gruid.Msg, error) {
	msg, ok := <-tk.msgs
	if ok {
		return msg, nil
	}
	return nil, nil
}

type rectangle struct {
	xmin, xmax, ymin, ymax int
}

func (tk *Driver) Flush(frame gruid.Frame) {
	w, h := frame.Width, frame.Height
	rects := []rectangle{}
	r := rectangle{w - 1, 0, h - 1, 0}
	n := 0
	for _, cdraw := range frame.Cells {
		cs := cdraw.Cell
		x, y := cdraw.Pos.X, cdraw.Pos.Y
		tk.draw(cs, x, y)
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
	tk.ir.Eval("appscreen put %{0%s} -format png -to %{1%d} %{2%d} %{3%d} %{4%d}", png,
		xmin*tk.tw, ymin*tk.th, (xmax+1)*tk.tw, (ymax+1)*tk.th)
}

func (tk *Driver) draw(cs gruid.Cell, x, y int) {
	var img *image.RGBA
	if im, ok := tk.cache[cs]; ok {
		img = im
	} else {
		img = tk.TileManager.GetImage(cs)
		tk.cache[cs] = img // TODO: do something if image is nil?
	}
	draw.Draw(tk.canvas, image.Rect(x*tk.tw, tk.th*y, (x+1)*tk.tw, (y+1)*tk.th), img, image.Point{0, 0}, draw.Over)
}

func (tk *Driver) Close() {
	tk.ir.Eval("exit 0")
	<-tk.ir.Done
	close(tk.msgs)
	tk.ir = nil
	tk.cache = nil
}

func (tk *Driver) ClearCache() {
	for c, _ := range tk.cache {
		delete(tk.cache, c)
	}
}
