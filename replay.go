package gorltk

import "time"

// Replay is an implementation of Model that replays a game from a list of
// frames.
type Replay struct {
	Frames []Frame
	undo   [][]FrameCell
	frame  int
	auto   bool
	speed  time.Duration
	action repAction
}

type repAction int

const (
	ReplayNone repAction = iota
	ReplayNext
	ReplayPrevious
	ReplayTogglePause
	ReplayQuit
	ReplaySpeedMore
	ReplaySpeedLess
)

type msgTick int // frame number

func (rep *Replay) Init() Cmd {
	rep.auto = true
	rep.speed = 1
	rep.undo = [][]FrameCell{}
	return rep.tick()
}

func (rep *Replay) Update(msg Msg) Cmd {
	switch msg := msg.(type) {
	case MsgKeyDown:
		switch msg.Key {
		case "Q", "q", KeyEscape:
			rep.action = ReplayQuit
		case "p", "P", KeySpace:
			rep.action = ReplayTogglePause
		case "+", ">":
			rep.action = ReplaySpeedMore
		case "-", "<":
			rep.action = ReplaySpeedLess
		case KeyArrowRight, KeyArrowDown, KeyEnter, "j", "n", "f":
			rep.action = ReplayNext
			rep.auto = false
		case KeyArrowLeft, KeyArrowUp, KeyBackspace, "k", "N", "b":
			rep.action = ReplayPrevious
			rep.auto = false
		}
	case MsgMouseDown:
		switch msg.Button {
		case ButtonMain:
			rep.action = ReplayTogglePause
		case ButtonAuxiliary:
			rep.action = ReplayNext
			rep.action = ReplayTogglePause
			rep.auto = false
		case ButtonSecondary:
			rep.action = ReplayPrevious
			rep.auto = false
		}
	case msgTick:
		if rep.auto && rep.frame == int(msg) {
			rep.action = ReplayNext
		}
	}
	switch rep.action {
	case ReplayNext:
		if rep.frame >= len(rep.Frames) {
			rep.action = ReplayNone
			break
		} else if rep.frame < 0 {
			rep.frame = 0
		}
		rep.frame++
	case ReplayPrevious:
		if rep.frame <= 1 {
			rep.action = ReplayNone
			break
		} else if rep.frame >= len(rep.Frames) {
			rep.frame = len(rep.Frames)
		}
		rep.frame--
	case ReplayQuit:
		return Quit
	case ReplayTogglePause:
		rep.auto = !rep.auto
	case ReplaySpeedMore:
		rep.speed *= 2
		if rep.speed > 16 {
			rep.speed = 16
		}
	case ReplaySpeedLess:
		rep.speed /= 2
		if rep.speed < 1 {
			rep.speed = 1
		}
	}
	return rep.tick()
}

func (rep *Replay) Draw(gd *Grid) {
	switch rep.action {
	case ReplayNext:
		df := rep.Frames[rep.frame-1]
		rep.undo = append(rep.undo, []FrameCell{})
		j := len(rep.undo) - 1
		for _, dr := range df.Cells {
			i := gd.GetIndex(dr.Pos)
			c := gd.cellBuffer[i]
			rep.undo[j] = append(rep.undo[j], FrameCell{Cell: c, Pos: dr.Pos})
			gd.SetCell(dr.Pos, dr.Cell)
		}
	case ReplayPrevious:
		df := rep.undo[len(rep.undo)-1]
		for _, dr := range df {
			gd.SetCell(dr.Pos, dr.Cell)
		}
		rep.undo = rep.undo[:len(rep.undo)-1]
	}
}

func (rep *Replay) tick() Cmd {
	if !rep.auto || rep.frame > len(rep.Frames)-1 || rep.frame < 0 {
		return nil
	}
	var d time.Duration
	if rep.frame > 0 {
		d = rep.Frames[rep.frame].Time.Sub(rep.Frames[rep.frame-1].Time)
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
	n := rep.frame
	return func() Msg {
		t := time.NewTimer(d)
		<-t.C
		return msgTick(n)
	}
}
