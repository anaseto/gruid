// Package ui defines common UI utilities for gruid.
package ui

import (
	"time"

	"github.com/anaseto/gruid"
)

// NewReplay returns a new Replay with a given configuration.
func NewReplay(cfg ReplayConfig) *Replay {
	rep := &Replay{
		gd:      cfg.Grid,
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
		rep.keys.FrameNext = []gruid.Key{gruid.KeyArrowRight, gruid.KeyArrowDown, gruid.KeyEnter, "j", "n", "f"}
	}
	if rep.keys.FramePrev == nil {
		rep.keys.FramePrev = []gruid.Key{gruid.KeyArrowLeft, gruid.KeyArrowUp, gruid.KeyBackspace, "k", "N", "b"}
	}
	return rep
}

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
	Grid         gruid.Grid         // grid to use for drawing
	FrameDecoder gruid.FrameDecoder // frame decoder
	Keys         ReplayKeys         // optional custom key bindings
}

// Replay represents an application's session with the given recorded frames.
// It implements the gruid.Model interface.
type Replay struct {
	decoder gruid.FrameDecoder
	frames  []gruid.Frame
	gd      gruid.Grid
	undo    [][]gruid.FrameCell
	fidx    int // frame index
	auto    bool
	speed   time.Duration
	action  repAction
	init    bool // Update received MsgInit
	keys    ReplayKeys
}

type repAction int

const (
	replayNone repAction = iota
	replayNext
	replayPrevious
	replayTogglePause
	replayQuit
	replaySpeedMore
	replaySpeedLess
)

type msgTick int // frame number

func (rep *Replay) decodeNext() {
	if rep.fidx >= len(rep.frames)-1 {
		frame, err := rep.decoder.Decode()
		if err == nil {
			rep.frames = append(rep.frames, frame)
		}
	}
}

// Update implements Model.Update for Replay.
func (rep *Replay) Update(msg gruid.Msg) gruid.Effect {
	rep.action = replayNone
	switch msg := msg.(type) {
	case gruid.MsgDraw:
		return nil
	case gruid.MsgInit:
		rep.init = true
		return rep.tick()
	case gruid.MsgKeyDown:
		key := msg.Key
		switch {
		case key.In(rep.keys.Quit):
			if rep.init {
				rep.action = replayQuit
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
	case gruid.MsgMouse:
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
	case msgTick:
		if rep.auto && rep.fidx == int(msg) {
			rep.action = replayNext
		}
	}
	switch rep.action {
	case replayNext:
		rep.decodeNext()
		if rep.fidx >= len(rep.frames) {
			rep.action = replayNone
			break
		} else if rep.fidx < 0 {
			rep.fidx = 0
		}
		rep.fidx++
	case replayPrevious:
		if rep.fidx <= 1 {
			rep.action = replayNone
			break
		} else if rep.fidx >= len(rep.frames) {
			rep.fidx = len(rep.frames)
		}
		rep.fidx--
	case replayQuit:
		return gruid.Quit()
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
	rep.draw()
	if !rep.auto || rep.fidx > len(rep.frames)-1 || rep.fidx < 0 || rep.action == replayNone {
		return nil
	}
	return rep.tick()
}

// The grid state is actually the replay state so we draw the grid on Update
// instead of Draw.
func (rep *Replay) draw() {
	switch rep.action {
	case replayNext:
		frame := rep.frames[rep.fidx-1]
		rep.undo = append(rep.undo, []gruid.FrameCell{})
		j := len(rep.undo) - 1
		w, h := rep.gd.Size()
		if frame.Width > w || frame.Height > h {
			rep.gd = rep.gd.Resize(frame.Width, frame.Height)
		}
		for _, fc := range frame.Cells {
			c := rep.gd.GetCell(fc.Pos)
			rep.undo[j] = append(rep.undo[j], gruid.FrameCell{Cell: c, Pos: fc.Pos})
			rep.gd.SetCell(fc.Pos, fc.Cell)
		}
	case replayPrevious:
		fcells := rep.undo[len(rep.undo)-1]
		for _, fc := range fcells {
			rep.gd.SetCell(fc.Pos, fc.Cell)
		}
		rep.undo = rep.undo[:len(rep.undo)-1]
	}
}

// Draw implements Model.Draw for Replay.
func (rep *Replay) Draw() gruid.Grid {
	return rep.gd
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
