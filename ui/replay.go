// Package ui defines common UI utilities for gruid: menu/table widget,
// pager, text input, label, text drawing facilities and replay functionality.
package ui

import (
	"time"

	"github.com/anaseto/gruid"
)

// ReplayKeys contains key bindings configuration for the replay.
type ReplayKeys struct {
	Quit      []gruid.Key // quit replay
	Pause     []gruid.Key // pause replay
	SpeedMore []gruid.Key // increase replay speed
	SpeedLess []gruid.Key // decrease replay speed
	FrameNext []gruid.Key // manually go to next frame
	FramePrev []gruid.Key // manually go to previous frame
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
		rep.keys.SpeedMore = []gruid.Key{"+", ">"}
	}
	if rep.keys.SpeedLess == nil {
		rep.keys.SpeedLess = []gruid.Key{"-", "<"}
	}
	if rep.keys.FrameNext == nil {
		rep.keys.FrameNext = []gruid.Key{gruid.KeyArrowRight, gruid.KeyArrowDown, "j", "n", "f"}
	}
	if rep.keys.FramePrev == nil {
		rep.keys.FramePrev = []gruid.Key{gruid.KeyArrowLeft, gruid.KeyArrowUp, "k", "N", "b"}
	}
	return rep
}

type repAction int

const (
	replayNone repAction = iota
	replayNext
	replayPrevious
	replayTogglePause
	replaySpeedMore
	replaySpeedLess
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
		if rep.speed > 16 {
			rep.speed = 16
		}
	case replaySpeedLess:
		rep.speed /= 2
		if rep.speed < 1 {
			rep.speed = 1
		}
	}
}

// The grid state is actually the replay state so we draw the grid on Update
// instead of Draw.
func (rep *Replay) draw() {
	switch rep.action {
	case replayNext:
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
	case replayPrevious:
		fcells := rep.undo[len(rep.undo)-1]
		for _, fc := range fcells {
			rep.grid.Set(fc.P, fc.Cell)
		}
		rep.undo = rep.undo[:len(rep.undo)-1]
	}
}

// Draw implements gruid.Model.Draw for Replay.
func (rep *Replay) Draw() gruid.Grid {
	if rep.init && rep.action == replayNone {
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
