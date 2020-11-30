package models

import (
	"time"

	"github.com/anaseto/gorltk"
)

// NewReplay returns a Model that runs a replay of an application's session with
// the given recorded frames.
func NewReplay(frames []gorltk.Frame) gorltk.Model {
	return &replay{Frames: frames}
}

type replay struct {
	Frames []gorltk.Frame
	undo   [][]gorltk.FrameCell
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

func (rep *replay) Init() gorltk.Cmd {
	rep.auto = true
	rep.speed = 1 // default to real time speed
	rep.undo = [][]gorltk.FrameCell{}
	return rep.tick()
}

func (rep *replay) Update(msg gorltk.Msg) gorltk.Cmd {
	rep.action = replayNone
	switch msg := msg.(type) {
	case gorltk.MsgKeyDown:
		switch msg.Key {
		case "Q", "q", gorltk.KeyEscape:
			rep.action = replayQuit
		case "p", "P", gorltk.KeySpace:
			rep.action = replayTogglePause
		case "+", ">":
			rep.action = replaySpeedMore
		case "-", "<":
			rep.action = replaySpeedLess
		case gorltk.KeyArrowRight, gorltk.KeyArrowDown, gorltk.KeyEnter, "j", "n", "f":
			rep.action = replayNext
			rep.auto = false
		case gorltk.KeyArrowLeft, gorltk.KeyArrowUp, gorltk.KeyBackspace, "k", "N", "b":
			rep.action = replayPrevious
			rep.auto = false
		}
	case gorltk.MsgMouseDown:
		switch msg.Button {
		case gorltk.ButtonMain:
			rep.action = replayTogglePause
		case gorltk.ButtonAuxiliary:
			rep.action = replayNext
			rep.action = replayTogglePause
			rep.auto = false
		case gorltk.ButtonSecondary:
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
		if rep.fidx >= len(rep.Frames) {
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
		} else if rep.fidx >= len(rep.Frames) {
			rep.fidx = len(rep.Frames)
		}
		rep.fidx--
	case replayQuit:
		return gorltk.Quit
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
	if !rep.auto || rep.fidx > len(rep.Frames)-1 || rep.fidx < 0 || rep.action == replayNone {
		return nil
	}
	return rep.tick()
}

func (rep *replay) Draw(gd *gorltk.Grid) {
	switch rep.action {
	case replayNext:
		frame := rep.Frames[rep.fidx-1]
		rep.undo = append(rep.undo, []gorltk.FrameCell{})
		j := len(rep.undo) - 1
		w, h := gd.Size()
		if frame.Width > w || frame.Height > h {
			gd.Resize(frame.Width, frame.Height)
		}
		for _, fc := range frame.Cells {
			c := gd.GetCell(fc.Pos)
			rep.undo[j] = append(rep.undo[j], gorltk.FrameCell{Cell: c, Pos: fc.Pos})
			gd.SetCell(fc.Pos, fc.Cell)
		}
	case replayPrevious:
		fcells := rep.undo[len(rep.undo)-1]
		for _, fc := range fcells {
			gd.SetCell(fc.Pos, fc.Cell)
		}
		rep.undo = rep.undo[:len(rep.undo)-1]
	}
}

func (rep *replay) tick() gorltk.Cmd {
	var d time.Duration
	if rep.fidx > 0 {
		d = rep.Frames[rep.fidx].Time.Sub(rep.Frames[rep.fidx-1].Time)
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
	return func() gorltk.Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgTick(n)
	}
}
