// The package models defines some basic UI elements as gruid models. They can
// be used either as the main model for an application, or used inside Update
// and Draw of the main model.
package models

import (
	"time"

	"github.com/anaseto/gruid"
)

// NewReplay returns a Model that runs a replay of an application's session with
// the given recorded frames.
func NewReplay(cfg ReplayConfig) *Replay {
	return &Replay{
		gd:     cfg.Grid,
		frames: cfg.Frames,
		auto:   true,
		speed:  1,
		undo:   [][]gruid.FrameCell{},
	}
}

// ReplayConfig contains replay configuration.
type ReplayConfig struct {
	Grid   gruid.Grid    // grid to use for drawing
	Frames []gruid.Frame // recorded frames to replay
}

type Replay struct {
	frames []gruid.Frame
	gd     gruid.Grid
	undo   [][]gruid.FrameCell
	fidx   int // frame index
	auto   bool
	speed  time.Duration
	action repAction
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

// Init implements Model.Init for Replay. It returns a timer command for
// starting automatic replay.
func (rep *Replay) Init() gruid.Cmd {
	return rep.tick()
}

// Update implements Model.Update for Replay.
func (rep *Replay) Update(msg gruid.Msg) gruid.Cmd {
	rep.action = replayNone
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		switch msg.Key {
		case "Q", "q", gruid.KeyEscape:
			rep.action = replayQuit
		case "p", "P", gruid.KeySpace:
			rep.action = replayTogglePause
		case "+", ">":
			rep.action = replaySpeedMore
		case "-", "<":
			rep.action = replaySpeedLess
		case gruid.KeyArrowRight, gruid.KeyArrowDown, gruid.KeyEnter, "j", "n", "f":
			rep.action = replayNext
			rep.auto = false
		case gruid.KeyArrowLeft, gruid.KeyArrowUp, gruid.KeyBackspace, "k", "N", "b":
			rep.action = replayPrevious
			rep.auto = false
		}
	case gruid.MsgMouseDown:
		switch msg.Button {
		case gruid.ButtonMain:
			rep.action = replayTogglePause
		case gruid.ButtonAuxiliary:
			rep.action = replayNext
			rep.action = replayTogglePause
			rep.auto = false
		case gruid.ButtonSecondary:
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
		return gruid.Quit
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
	if !rep.auto || rep.fidx > len(rep.frames)-1 || rep.fidx < 0 || rep.action == replayNone {
		return nil
	}
	return rep.tick()
}

// Draw implements Model.Draw for Replay.
func (rep *Replay) Draw() gruid.Grid {
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
	if d <= 5*time.Millisecond {
		d = 5 * time.Millisecond
	}
	n := rep.fidx
	return func() gruid.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgTick(n)
	}
}
