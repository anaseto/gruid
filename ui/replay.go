// Package ui defines common UI utilities for gruid: menu/table widget,
// pager, text input, label, text drawing facilities and replay functionality.
package ui

import (
	"strings"
	"time"

	"github.com/anaseto/gruid"
)

// ReplayKeys contains key bindings configuration for the replay.
type ReplayKeys struct {
	Quit      []gruid.Key // quit replay (default: q, Q, esc)
	Pause     []gruid.Key // pause replay (default: p, P, space)
	SpeedMore []gruid.Key // increase replay speed (default: +, })
	SpeedLess []gruid.Key // decrease replay speed (default: -, {)
	FrameNext []gruid.Key // go to next frame (default: arrow right, l)
	FramePrev []gruid.Key // go to previous frame (default: arrow left, h)
	Forward   []gruid.Key // go 1 minute forward (default: arrow down, j)
	Backward  []gruid.Key // go 1 minute backward (default: arrow up, k)
	Help      []gruid.Key // key bindings help (default: ?)
}

// ReplayConfig contains replay configuration.
type ReplayConfig struct {
	Grid         gruid.Grid          // grid to use for drawing
	FrameDecoder *gruid.FrameDecoder // frame decoder
	Keys         ReplayKeys          // optional custom key bindings
}

// Replay represents an application's session with the given recorded frames.
//
// Replay implements gruid.Model and can be used as main model of an
// application.
type Replay struct {
	decoder *gruid.FrameDecoder
	frames  []gruid.Frame
	grid    gruid.Grid
	undo    [][]gruid.FrameCell
	fidx    int // frame index
	auto    bool
	speed   time.Duration
	action  repAction
	init    bool // Update received MsgInit
	keys    ReplayKeys
	dirty   bool
	help    bool
	pager   *Pager
}

// NewReplay returns a new Replay with a given configuration.
func NewReplay(cfg ReplayConfig) *Replay {
	rep := &Replay{
		grid:    cfg.Grid,
		decoder: cfg.FrameDecoder,
		auto:    true,
		speed:   1,
		undo:    [][]gruid.FrameCell{},
		keys:    cfg.Keys,
	}
	if rep.keys.Quit == nil {
		rep.keys.Quit = []gruid.Key{gruid.KeyEscape, "Q", "q"}
	}
	if rep.keys.Pause == nil {
		rep.keys.Pause = []gruid.Key{gruid.KeySpace, "P", "p"}
	}
	if rep.keys.SpeedMore == nil {
		rep.keys.SpeedMore = []gruid.Key{"+", "}"}
	}
	if rep.keys.SpeedLess == nil {
		rep.keys.SpeedLess = []gruid.Key{"-", "{"}
	}
	if rep.keys.FrameNext == nil {
		rep.keys.FrameNext = []gruid.Key{gruid.KeyArrowRight, "l"}
	}
	if rep.keys.FramePrev == nil {
		rep.keys.FramePrev = []gruid.Key{gruid.KeyArrowLeft, "h"}
	}
	if rep.keys.Forward == nil {
		rep.keys.Forward = []gruid.Key{gruid.KeyArrowUp, "k"}
	}
	if rep.keys.Backward == nil {
		rep.keys.Backward = []gruid.Key{gruid.KeyArrowDown, "j"}
	}
	if rep.keys.Help == nil {
		rep.keys.Help = []gruid.Key{"?"}
	}
	rep.dirty = true
	max := cfg.Grid.Size()
	rep.pager = NewPager(PagerConfig{
		Grid: gruid.NewGrid(max.X, max.Y),
		Box:  &Box{Title: Text("Help")},
		Keys: PagerKeys{Quit: []gruid.Key{gruid.KeyEscape, "q", "Q", "x", "X", "?"}},
	})
	rep.setPagerLines()
	return rep
}

func (rep *Replay) setPagerLines() {
	lines := []StyledText{}
	fmtLine := func(title string, keys []gruid.Key) {
		b := strings.Builder{}
		for i, k := range keys {
			b.WriteString(string(k))
			if i < len(keys)-1 {
				b.WriteString(" or ")
			}
		}
		lines = append(lines, Textf("%-30s %s", title, b.String()))
	}
	fmtLine("Quit", rep.keys.Quit)
	fmtLine("Pause", rep.keys.Pause)
	fmtLine("Increase speed", rep.keys.SpeedMore)
	fmtLine("Decrease speed", rep.keys.SpeedLess)
	fmtLine("Go to next frame", rep.keys.FrameNext)
	fmtLine("Go to previous frame", rep.keys.FramePrev)
	fmtLine("Go 1 minute forward", rep.keys.Forward)
	fmtLine("Go 1 minute backward", rep.keys.Backward)
	rep.pager.SetLines(lines)
}

type repAction int

const (
	replayNone repAction = iota
	replayNext
	replayPrevious
	replayTogglePause
	replaySpeedMore
	replaySpeedLess
	replayForward
	replayBackward
)

type msgTick int // frame number

func (rep *Replay) decodeNext() {
	if rep.fidx >= len(rep.frames)-1 {
		frame := gruid.Frame{}
		err := rep.decoder.Decode(&frame)
		if err == nil {
			rep.frames = append(rep.frames, frame)
		}
	}
}

// Update implements gruid.Model.Update for Replay. It considers mouse message
// coordinates to be absolute in its grid. If a gruid.MsgInit is passed to
// Update, the replay will behave as if it is the main model of an application,
// and send a gruid.Quit() command on a quit request.
func (rep *Replay) Update(msg gruid.Msg) gruid.Effect {
	rep.action = replayNone
	if rep.help {
		return rep.updateHelp(msg)
	}
	switch msg := msg.(type) {
	case gruid.MsgInit:
		rep.init = true
		rep.decodeNext()
		return rep.tick()
	case gruid.MsgKeyDown:
		eff := rep.updateMsgKeyDown(msg)
		if eff != nil {
			return eff
		}
	case gruid.MsgMouse:
		rep.updateMsgMouse(msg)
	case msgTick:
		if rep.auto && rep.fidx == int(msg) {
			rep.action = replayNext
		}
	}
	rep.handleAction()
	rep.draw()
	if !rep.auto || rep.fidx > len(rep.frames)-1 || rep.action == replayNone {
		return nil
	}
	return rep.tick()
}

func (rep *Replay) updateHelp(msg gruid.Msg) gruid.Effect {
	rep.pager.Update(msg)
	switch rep.pager.Action() {
	case PagerQuit:
		rep.help = false
	}
	return nil
}

func (rep *Replay) updateMsgKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	key := msg.Key
	switch {
	case key.In(rep.keys.Quit):
		if rep.init {
			return gruid.End()
		}
	case key.In(rep.keys.Pause):
		rep.action = replayTogglePause
	case key.In(rep.keys.SpeedMore):
		rep.action = replaySpeedMore
	case key.In(rep.keys.SpeedLess):
		rep.action = replaySpeedLess
	case key.In(rep.keys.FrameNext):
		rep.action = replayNext
		rep.auto = false
	case key.In(rep.keys.FramePrev):
		rep.action = replayPrevious
		rep.auto = false
	case key.In(rep.keys.Forward):
		rep.action = replayForward
	case key.In(rep.keys.Backward):
		rep.action = replayBackward
	case key.In(rep.keys.Help):
		rep.dirty = true
		rep.help = true
		rep.auto = false
	}
	return nil
}

func (rep *Replay) updateMsgMouse(msg gruid.MsgMouse) {
	if !msg.P.In(rep.grid.Bounds()) {
		return
	}
	switch msg.Action {
	case gruid.MouseMain:
		rep.action = replayTogglePause
	case gruid.MouseAuxiliary:
		rep.action = replayNext
		rep.auto = false
	case gruid.MouseSecondary:
		rep.action = replayPrevious
		rep.auto = false
	}
}

// SetFrame sets the current frame number to be displayed.
func (rep *Replay) SetFrame(n int) {
	for rep.fidx < n {
		rep.decodeNext()
		if rep.fidx >= len(rep.frames) {
			break
		}
		rep.fidx++
		rep.next()
	}
	for rep.fidx > n {
		if rep.fidx <= 0 {
			break
		}
		rep.fidx--
		rep.previous()
	}
	rep.dirty = true
}

// Seek moves replay forward/backward by the given duration.
func (rep *Replay) Seek(d time.Duration) {
	rep.decodeNext()
	if len(rep.frames) == 0 {
		return
	}
	n := rep.fidx - 1
	if n < 0 || n >= len(rep.frames) {
		return
	}
	var t time.Time
	t = rep.frames[n].Time.Add(d)
	if d > 0 {
		for t.After(rep.frames[rep.fidx-1].Time) {
			rep.decodeNext()
			if rep.fidx >= len(rep.frames) {
				break
			}
			rep.fidx++
			rep.next()
		}
	} else {
		for t.Before(rep.frames[rep.fidx-1].Time) {
			if rep.fidx <= 1 {
				break
			}
			rep.fidx--
			rep.previous()
		}
	}
	rep.dirty = true
}

func (rep *Replay) handleAction() {
	switch rep.action {
	case replayNext:
		rep.decodeNext()
		if rep.fidx >= len(rep.frames) {
			rep.action = replayNone
			break
		}
		rep.fidx++
	case replayPrevious:
		if rep.fidx <= 0 {
			rep.action = replayNone
			break
		}
		rep.fidx--
	case replayTogglePause:
		rep.auto = !rep.auto
	case replaySpeedMore:
		rep.speed *= 2
		if rep.speed > 64 {
			rep.speed = 64
		}
	case replaySpeedLess:
		rep.speed /= 2
		if rep.speed < 1 {
			rep.speed = 1
		}
	}
}

func (rep *Replay) next() {
	frame := rep.frames[rep.fidx-1]
	rep.undo = append(rep.undo, []gruid.FrameCell{})
	j := len(rep.undo) - 1
	max := rep.grid.Size()
	if frame.Width > max.X || frame.Height > max.Y {
		rep.grid = rep.grid.Resize(frame.Width, frame.Height)
	}
	for _, fc := range frame.Cells {
		c := rep.grid.At(fc.P)
		rep.undo[j] = append(rep.undo[j], gruid.FrameCell{Cell: c, P: fc.P})
		rep.grid.Set(fc.P, fc.Cell)
	}
}

func (rep *Replay) previous() {
	fcells := rep.undo[len(rep.undo)-1]
	for _, fc := range fcells {
		rep.grid.Set(fc.P, fc.Cell)
	}
	rep.undo = rep.undo[:len(rep.undo)-1]
}

// The grid state is actually the replay state so we draw the grid on Update
// instead of Draw.
func (rep *Replay) draw() {
	switch rep.action {
	case replayNext:
		rep.next()
	case replayPrevious:
		rep.previous()
	case replayBackward:
		rep.Seek(-time.Minute)
	case replayForward:
		rep.Seek(time.Minute)
	}
	if rep.action != replayNone {
		rep.dirty = true
	}
}

// Draw implements gruid.Model.Draw for Replay.
func (rep *Replay) Draw() gruid.Grid {
	if rep.help {
		return rep.pager.Draw()
	}
	if rep.init && !rep.dirty {
		return rep.grid.Slice(gruid.Range{})
	}
	return rep.grid
}

func (rep *Replay) tick() gruid.Cmd {
	var d time.Duration
	if rep.fidx > 0 {
		d = rep.frames[rep.fidx].Time.Sub(rep.frames[rep.fidx-1].Time)
	} else {
		d = 0
	}
	if d >= 2*time.Second {
		d = 2 * time.Second
	}
	d = d / rep.speed
	mininterval := time.Second / 240
	if d <= mininterval {
		d = mininterval
	}
	n := rep.fidx
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgTick(n)
	}
}
