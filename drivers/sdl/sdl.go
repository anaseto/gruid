package sdl

import (
	"bytes"
	"fmt"
	"image"
	//"log"
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
	TileManager TileManager // for retrieving tiles
	Width       int32       // initial screen width in cells
	Height      int32       // initial screen height in cells

	window    *sdl.Window
	renderer  *sdl.Renderer
	textures  map[gruid.Cell]*sdl.Texture
	surfaces  map[gruid.Cell]*sdl.Surface
	tw        int32
	th        int32
	mousepos  gruid.Position
	mousedrag gruid.MouseAction
	done      chan struct{}
}

// Init implements gruid.Driver.Init. It initializes structures and calls
// sdl.Init().
func (dr *Driver) Init() error {
	dr.done = make(chan struct{})
	w, h := dr.TileManager.TileSize()
	dr.tw, dr.th = int32(w), int32(h)
	dr.textures = make(map[gruid.Cell]*sdl.Texture)
	dr.surfaces = make(map[gruid.Cell]*sdl.Surface)
	dr.mousedrag = -1
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
		return err
	}
	dr.window, err = sdl.CreateWindow("gruid go-sdl2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		dr.Width*dr.tw, dr.Height*dr.th, sdl.WINDOW_SHOWN)
	if err != nil {
		return fmt.Errorf("failed to create sdl window: %v", err)
	}
	dr.renderer, err = sdl.CreateRenderer(dr.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return fmt.Errorf("failed to create sdl renderer: %v", err)
	}
	dr.renderer.Clear()
	sdl.StartTextInput()
	return nil
}

// PollMsg implements gruid.Driver.PollMsg.
func (dr *Driver) PollMsg() (gruid.Msg, error) {
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch ev := event.(type) {
			//case *sdl.QuitEvent:
			//return nil, errors.New("Quit") // TODO handle quit properly (send special message?)
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
				return msg, nil
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
				return msg, nil
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
				msg.MousePos = gruid.Position{X: int((ev.X - 1) / dr.tw), Y: int((ev.Y - 1) / dr.th)}
				switch ev.Type {
				case sdl.MOUSEBUTTONDOWN:
					if dr.mousedrag != -1 {
						continue
					}
					if msg.MousePos.X < 0 || msg.MousePos.X >= int(dr.Width) ||
						msg.MousePos.Y < 0 || msg.MousePos.Y >= int(dr.Height) {
						continue
					}
					msg.Time = time.Now()
					msg.Action = action
					dr.mousedrag = action
				case sdl.MOUSEBUTTONUP:
					if dr.mousedrag != action {
						continue
					}
					if msg.MousePos.X < 0 || msg.MousePos.X >= int(dr.Width) ||
						msg.MousePos.Y < 0 || msg.MousePos.Y >= int(dr.Height) {
						msg.MousePos = gruid.Position{}
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
				dr.mousepos = msg.MousePos
				return msg, nil
			case *sdl.MouseMotionEvent:
				msg := gruid.MsgMouse{}
				msg.MousePos = gruid.Position{X: int((ev.X - 1) / dr.tw), Y: int((ev.Y - 1) / dr.th)}
				if msg.MousePos.X < 0 || msg.MousePos.X >= int(dr.Width) ||
					msg.MousePos.Y < 0 || msg.MousePos.Y >= int(dr.Height) {
					continue
				}
				msg.Time = time.Now()
				msg.Action = gruid.MouseMove
				dr.mousepos = msg.MousePos
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
				return msg, nil
			case *sdl.MouseWheelEvent:
				msg := gruid.MsgMouse{}
				if ev.Y > 0 {
					msg.Action = gruid.MouseWheelUp
				} else if ev.Y < 0 {
					msg.Action = gruid.MouseWheelDown
				} else {
					continue
				}
				msg.MousePos = dr.mousepos
				msg.Time = time.Now()
				return msg, nil
			case *sdl.WindowEvent:
				switch ev.Type {
				case sdl.WINDOWEVENT_SIZE_CHANGED:
					return gruid.MsgScreenSize{Width: int(ev.Data1 / dr.tw), Height: int(ev.Data2 / dr.th), Time: time.Now()}, nil
				}
			}
		}
		t := time.NewTimer(5 * time.Millisecond)
		select {
		case <-dr.done:
			return nil, nil
		case <-t.C:
		}
	}
}

// Flush implements gruid.Driver.Flush.
func (dr *Driver) Flush(frame gruid.Frame) {
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
		img := dr.TileManager.GetImage(cs)
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
	rect := sdl.Rect{int32(x) * dr.tw, int32(y) * dr.th, dr.tw, dr.th}
	dr.renderer.Copy(tx, nil, &rect)
}

// Close implements gruid.Driver.Close. It releases some resources and calls sdl.Quit.
func (dr *Driver) Close() {
	dr.ClearCache()
	dr.surfaces = nil
	dr.textures = nil
	sdl.StopTextInput()
	dr.renderer.Destroy()
	dr.window.Destroy()
	close(dr.done)
	// wait a little for any last PollMsg to complete: it's not ideal, but
	// good enough.
	time.Sleep(2 * time.Millisecond)
	sdl.Quit()
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
