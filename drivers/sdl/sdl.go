package sdl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"log"
	"time"
	"unicode/utf8"

	"golang.org/x/image/bmp"

	"github.com/anaseto/gruid"
	"github.com/veandco/go-sdl2/sdl"
)

// TileManager manages tiles fetching.
type TileManager interface {
	// GetImage returns the image to be used for a given cell style.
	GetImage(gruid.Cell) *image.RGBA

	// TileSize returns the (width, height) in pixels of the tiles. Both
	// should be positive and non-zero.
	TileSize() (int, int)
}

// Driver implements gruid.Driver using the go-sdl2 bindings for the SDL
// library. When using an gruid.App, Start has to be used on the main routine,
// as the video functions of SDL are not thread safe.
type Driver struct {
	tm         TileManager
	width      int32
	height     int32
	fullscreen bool
	tw         int32
	th         int32

	window    *sdl.Window
	renderer  *sdl.Renderer
	textures  map[gruid.Cell]*sdl.Texture
	surfaces  map[gruid.Cell]*sdl.Surface
	mousepos  gruid.Position
	mousedrag gruid.MouseAction
	done      chan struct{}
	init      bool
	reqredraw chan bool // request redraw
	noQuit    bool      // do not quit on close
	actions   chan func()
}

// Config contains configurations options for the driver.
type Config struct {
	TileManager TileManager // for retrieving tiles
	Width       int32       // initial screen width in cells (default: 80)
	Height      int32       // initial screen height in cells (default: 24)
	Fullscreen  bool        // use “real” fullscreen with a videomode change
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
	dr.fullscreen = cfg.Fullscreen
	dr.SetTileManager(cfg.TileManager)
	return dr
}

// SetTileManager allows to change the used tile manager. If the driver is
// already running, change will take effect with next Flush so that the
// function is thread safe.
func (dr *Driver) SetTileManager(tm TileManager) {
	fn := func() {
		dr.tm = tm
		w, h := tm.TileSize()
		dr.tw, dr.th = int32(w), int32(h)
		if dr.tw <= 0 {
			dr.tw = 1
		}
		if dr.th <= 0 {
			dr.th = 1
		}
		if dr.init {
			dr.ClearCache()
			dr.window.SetSize(dr.width*dr.tw, dr.height*dr.th)
			select {
			case dr.reqredraw <- true:
			default:
			}
		}
	}
	if dr.init {
		select {
		case dr.actions <- fn:
		default:
		}
	} else {
		fn()
	}
}

// PreventQuit will make next call to Close keep sdl and the main window
// running. It can be used to chain two applications with the same sdl session
// and window. It is then your reponsibility to either run another application
// or call Close manually to properly quit.
func (dr *Driver) PreventQuit() {
	dr.noQuit = true
}

// Init implements gruid.Driver.Init. It initializes structures and calls
// sdl.Init().
func (dr *Driver) Init() error {
	dr.reqredraw = make(chan bool, 1)
	dr.actions = make(chan func(), 1)
	if dr.tm == nil {
		return errors.New("no tile manager provided")
	}
	var err error
	if dr.init {
		dr.window.SetSize(dr.width*dr.tw, dr.height*dr.th)
	} else {
		if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
			return err
		}
		dr.window, err = sdl.CreateWindow("gruid go-sdl2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			dr.width*dr.tw, dr.height*dr.th, sdl.WINDOW_SHOWN)
		if err != nil {
			return fmt.Errorf("failed to create sdl window: %v", err)
		}
		dr.renderer, err = sdl.CreateRenderer(dr.window, -1, sdl.RENDERER_ACCELERATED)
		if err != nil {
			return fmt.Errorf("failed to create sdl renderer: %v", err)
		}
		dr.window.SetResizable(false)
		if dr.fullscreen {
			dr.window.SetFullscreen(sdl.WINDOW_FULLSCREEN)
		}
		dr.renderer.Clear()
		sdl.StartTextInput()
		rect := sdl.Rect{X: 0, Y: 0, W: 100, H: 100}
		sdl.SetTextInputRect(&rect)
	}
	dr.textures = make(map[gruid.Cell]*sdl.Texture)
	dr.surfaces = make(map[gruid.Cell]*sdl.Surface)
	dr.mousedrag = -1
	dr.init = true
	return nil
}

// PollMsgs implements gruid.Driver.PollMsgs.
func (dr *Driver) PollMsgs(ctx context.Context, msgs chan<- gruid.Msg) error {
	send := func(msg gruid.Msg) {
		select {
		case msgs <- msg:
		case <-ctx.Done():
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		select {
		case <-dr.reqredraw:
			w, h := dr.window.GetSize()
			send(gruid.MsgScreen{Width: int(w / dr.tw), Height: int(h / dr.th), Time: time.Now()})
		default:
		}
		event := sdl.PollEvent()
		if event == nil {
			t := time.NewTimer(5 * time.Millisecond)
			select {
			case <-ctx.Done():
				return nil
			case <-t.C:
				continue
			}
		}
		switch ev := event.(type) {
		case *sdl.QuitEvent:
			send(gruid.MsgQuit(time.Now()))
		case *sdl.TextInputEvent:
			s := ev.GetText()
			if utf8.RuneCountInString(s) != 1 {
				// TODO: handle the case where an input
				// event would produce several
				// characters? We would have to keep
				// track of those characters, and send
				// several messages in a row.
				continue
			}
			msg := gruid.MsgKeyDown{}
			msg.Key = gruid.Key(s)
			msg.Time = time.Now()
			send(msg)
		//case *sdl.TextEditingEvent:
		// TODO: Handling this would allow to use an input
		// method for making compositions and chosing text.
		// I'm not sure what the API for this should be in
		// gruid or the driver.
		case *sdl.KeyboardEvent:
			c := ev.Keysym.Sym
			if ev.Type == sdl.KEYUP {
				continue
			}
			msg := gruid.MsgKeyDown{}
			if sdl.KMOD_LALT&ev.Keysym.Mod != 0 {
				msg.Mod |= gruid.ModAlt
			}
			if sdl.KMOD_LSHIFT&ev.Keysym.Mod != 0 || sdl.KMOD_RSHIFT&ev.Keysym.Mod != 0 {
				msg.Mod |= gruid.ModShift
			}
			if sdl.KMOD_LCTRL&ev.Keysym.Mod != 0 || sdl.KMOD_RCTRL&ev.Keysym.Mod != 0 {
				msg.Mod |= gruid.ModCtrl
			}
			if sdl.KMOD_RGUI&ev.Keysym.Mod != 0 {
				msg.Mod |= gruid.ModMeta
			}
			switch c {
			case sdl.K_DOWN:
				msg.Key = gruid.KeyArrowDown
			case sdl.K_LEFT:
				msg.Key = gruid.KeyArrowLeft
			case sdl.K_RIGHT:
				msg.Key = gruid.KeyArrowRight
			case sdl.K_UP:
				msg.Key = gruid.KeyArrowUp
			case sdl.K_BACKSPACE:
				msg.Key = gruid.KeyBackspace
			case sdl.K_DELETE:
				msg.Key = gruid.KeyDelete
			case sdl.K_END:
				msg.Key = gruid.KeyEnd
			case sdl.K_ESCAPE:
				msg.Key = gruid.KeyEscape
			case sdl.K_RETURN:
				msg.Key = gruid.KeyEnter
			case sdl.K_HOME:
				msg.Key = gruid.KeyHome
			case sdl.K_INSERT:
				msg.Key = gruid.KeyInsert
			case sdl.K_PAGEUP:
				msg.Key = gruid.KeyPageUp
			case sdl.K_PAGEDOWN:
				msg.Key = gruid.KeyPageDown
			case sdl.K_TAB:
				msg.Key = gruid.KeyTab
			}
			if ev.Keysym.Mod&sdl.KMOD_NUM == 0 {
				switch c {
				case sdl.K_KP_2:
					msg.Key = gruid.KeyArrowDown
				case sdl.K_KP_4:
					msg.Key = gruid.KeyArrowLeft
				case sdl.K_KP_6:
					msg.Key = gruid.KeyArrowRight
				case sdl.K_KP_8:
					msg.Key = gruid.KeyArrowUp
				case sdl.K_KP_BACKSPACE:
					msg.Key = gruid.KeyBackspace
				case sdl.K_KP_PERIOD:
					msg.Key = gruid.KeyDelete
				case sdl.K_KP_1:
					msg.Key = gruid.KeyEnd
				case sdl.K_KP_5, sdl.K_KP_ENTER:
					msg.Key = gruid.KeyEnter
				case sdl.K_KP_7:
					msg.Key = gruid.KeyHome
				case sdl.K_KP_0:
					msg.Key = gruid.KeyInsert
				case sdl.K_KP_9:
					msg.Key = gruid.KeyPageUp
				case sdl.K_KP_3:
					msg.Key = gruid.KeyPageDown
				}
			}
			if msg.Key == "" {
				continue
			}
			msg.Time = time.Now()
			send(msg)
		case *sdl.MouseButtonEvent:
			var action gruid.MouseAction
			switch ev.Button {
			case sdl.BUTTON_LEFT:
				action = gruid.MouseMain
			case sdl.BUTTON_MIDDLE:
				action = gruid.MouseAuxiliary
			case sdl.BUTTON_RIGHT:
				action = gruid.MouseSecondary
			default:
				continue
			}
			msg := gruid.MsgMouse{}
			msg.Pos = gruid.Position{X: int((ev.X - 1) / dr.tw), Y: int((ev.Y - 1) / dr.th)}
			switch ev.Type {
			case sdl.MOUSEBUTTONDOWN:
				if dr.mousedrag != -1 {
					continue
				}
				if msg.Pos.X < 0 || msg.Pos.X >= int(dr.width) ||
					msg.Pos.Y < 0 || msg.Pos.Y >= int(dr.height) {
					continue
				}
				msg.Time = time.Now()
				msg.Action = action
				dr.mousedrag = action
			case sdl.MOUSEBUTTONUP:
				if dr.mousedrag != action {
					continue
				}
				if msg.Pos.X < 0 || msg.Pos.X >= int(dr.width) ||
					msg.Pos.Y < 0 || msg.Pos.Y >= int(dr.height) {
					msg.Pos = gruid.Position{}
				}
				msg.Time = time.Now()
				msg.Action = gruid.MouseRelease
				dr.mousedrag = -1
			}
			mod := sdl.GetModState()
			if sdl.KMOD_LALT&mod != 0 {
				msg.Mod |= gruid.ModAlt
			}
			if sdl.KMOD_LSHIFT&mod != 0 || sdl.KMOD_RSHIFT&mod != 0 {
				msg.Mod |= gruid.ModShift
			}
			if sdl.KMOD_LCTRL&mod != 0 || sdl.KMOD_RCTRL&mod != 0 {
				msg.Mod |= gruid.ModCtrl
			}
			if sdl.KMOD_RGUI&mod != 0 {
				msg.Mod |= gruid.ModMeta
			}
			dr.mousepos = msg.Pos
			send(msg)
		case *sdl.MouseMotionEvent:
			msg := gruid.MsgMouse{}
			msg.Pos = gruid.Position{X: int((ev.X - 1) / dr.tw), Y: int((ev.Y - 1) / dr.th)}
			if msg.Pos == dr.mousepos {
				continue
			}
			if msg.Pos.X < 0 || msg.Pos.X >= int(dr.width) ||
				msg.Pos.Y < 0 || msg.Pos.Y >= int(dr.height) {
				continue
			}
			msg.Time = time.Now()
			msg.Action = gruid.MouseMove
			dr.mousepos = msg.Pos
			mod := sdl.GetModState()
			if sdl.KMOD_LALT&mod != 0 {
				msg.Mod |= gruid.ModAlt
			}
			if sdl.KMOD_LSHIFT&mod != 0 || sdl.KMOD_RSHIFT&mod != 0 {
				msg.Mod |= gruid.ModShift
			}
			if sdl.KMOD_LCTRL&mod != 0 || sdl.KMOD_RCTRL&mod != 0 {
				msg.Mod |= gruid.ModCtrl
			}
			if sdl.KMOD_RGUI&mod != 0 {
				msg.Mod |= gruid.ModMeta
			}
			send(msg)
		case *sdl.MouseWheelEvent:
			msg := gruid.MsgMouse{}
			if ev.Y > 0 {
				msg.Action = gruid.MouseWheelUp
			} else if ev.Y < 0 {
				msg.Action = gruid.MouseWheelDown
			} else {
				continue
			}
			msg.Pos = dr.mousepos
			msg.Time = time.Now()
			send(msg)
		case *sdl.WindowEvent:
			switch ev.Event {
			case sdl.WINDOWEVENT_EXPOSED:
				w, h := dr.window.GetSize()
				send(gruid.MsgScreen{Width: int(w / dr.tw), Height: int(h / dr.th), Time: time.Now()})
				//case sdl.WINDOWEVENT_SHOWN:
				//log.Print("shown")
				//case sdl.WINDOWEVENT_HIDDEN:
				//log.Print("hidden")
				//case sdl.WINDOWEVENT_MOVED:
				//log.Print("moved")
				//case sdl.WINDOWEVENT_RESIZED:
				//log.Print("resized")
				//case sdl.WINDOWEVENT_SIZE_CHANGED:
				//log.Print("size changed")
				//case sdl.WINDOWEVENT_MINIMIZED:
				//log.Print("minimized")
				//case sdl.WINDOWEVENT_MAXIMIZED:
				//log.Print("maximized")
				//case sdl.WINDOWEVENT_RESTORED:
				//log.Print("restored")
				//case sdl.WINDOWEVENT_ENTER:
				//log.Print("enter")
				//case sdl.WINDOWEVENT_LEAVE:
				//log.Print("leave")
				//case sdl.WINDOWEVENT_FOCUS_GAINED:
				//log.Print("focus gained")
				//case sdl.WINDOWEVENT_FOCUS_LOST:
				//log.Print("focus lost")
				//case sdl.WINDOWEVENT_CLOSE:
				//log.Print("close")
				//case sdl.WINDOWEVENT_TAKE_FOCUS:
				//log.Print("take focus")
				//case sdl.WINDOWEVENT_HIT_TEST:
				//log.Print("hit test")
				//case sdl.WINDOWEVENT_NONE:
				//log.Print("none")
			}
		}
	}
}

// Flush implements gruid.Driver.Flush.
func (dr *Driver) Flush(frame gruid.Frame) {
	select {
	case fn := <-dr.actions:
		fn()
	default:
	}
	if frame.Width != int(dr.width) || frame.Height != int(dr.height) {
		dr.width = int32(frame.Width)
		dr.height = int32(frame.Height)
		dr.window.SetSize(dr.width*dr.tw, dr.height*dr.th)
	}
	for _, cdraw := range frame.Cells {
		cs := cdraw.Cell
		x, y := cdraw.Pos.X, cdraw.Pos.Y
		dr.draw(cs, x, y)
	}
	dr.renderer.Present()
}

func (dr *Driver) draw(cs gruid.Cell, x, y int) {
	var tx *sdl.Texture
	if t, ok := dr.textures[cs]; ok {
		tx = t
	} else {
		img := dr.tm.GetImage(cs)
		if img == nil {
			log.Printf("no tile for %+v", cs)
			return
		}
		buf := bytes.Buffer{}
		err := bmp.Encode(&buf, img)
		if err != nil {
			log.Println(err)
			return
		}
		src, err := sdl.RWFromMem(buf.Bytes())
		if err != nil {
			log.Println(err)
			return
		}
		sf, err := sdl.LoadBMPRW(src, true)
		if err != nil {
			log.Println(err)
			return
		}
		dr.surfaces[cs] = sf
		tx, err = dr.renderer.CreateTextureFromSurface(sf)
		if err != nil {
			log.Println(err)
			return
		}
		dr.textures[cs] = tx
	}
	rect := sdl.Rect{X: int32(x) * dr.tw, Y: int32(y) * dr.th, W: dr.tw, H: dr.th}
	dr.renderer.Copy(tx, nil, &rect)
}

// Close implements gruid.Driver.Close. It releases some resources and calls sdl.Quit.
func (dr *Driver) Close() {
	if !dr.init {
		return
	}
	dr.ClearCache()
	dr.surfaces = nil
	dr.textures = nil
	if !dr.noQuit {
		sdl.StopTextInput()
		dr.renderer.Destroy()
		dr.window.Destroy()
		sdl.Quit()
		dr.init = false
		dr.noQuit = false
	}
}

// ClearCache clears the tile textures internal cache.
func (dr *Driver) ClearCache() {
	for i, s := range dr.surfaces {
		s.Free()
		delete(dr.surfaces, i)
	}
	for i, s := range dr.textures {
		s.Destroy()
		delete(dr.textures, i)
	}
}
