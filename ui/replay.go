// Package ui defines common UI utilities for gruid.
package ui

import (
	"time"

	"github.com/anaseto/gruid"
)

// NewReplay returns a new Replay with a given configuration.
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

// Replay represents an application's session with the given recorded frames.
// It implements the gruid.Model interface.
type Replay struct {
	frames []gruid.Frame
	gd     gruid.Grid
	undo   [][]gruid.FrameCell
	fidx   int // frame index
	auto   bool
	speed  time.Duration
	action repAction
	app    bool // whether running as main gruid.App model
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

// Update implements Model.Update for Replay.
func (rep *Replay) Update(msg gruid.Msg) gruid.Effect {
	rep.action = replayNone
	switch msg := msg.(type) {
	case gruid.MsgInit:
		rep.app = true
		return rep.tick()
	case gruid.MsgKeyDown:
		switch msg.Key {
		case "Q", "q", gruid.KeyEscape:
			if rep.app {
				rep.action = replayQuit
			}
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
	case gruid.MsgMouse:
		switch msg.Action {
		case gruid.MouseMain:
			rep.action = replayTogglePause
		case gruid.MouseAuxiliary:
			rep.action = replayNext
			rep.action = replayTogglePause
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
