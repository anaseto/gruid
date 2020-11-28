package gorltk

import "time"

// NewReplay returns a Model that runs a replay of a game with the given
// recorded frames.
func NewReplay(frames []Frame) Model {
	return &replay{Frames: frames}
}

type replay struct {
	Frames []Frame
	undo   [][]FrameCell
	frame  int
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

func (rep *replay) Init() Cmd {
	rep.auto = true
	rep.speed = 1
	rep.undo = [][]FrameCell{}
	return rep.tick()
}

func (rep *replay) Update(msg Msg) Cmd {
	switch msg := msg.(type) {
	case MsgKeyDown:
		switch msg.Key {
		case "Q", "q", KeyEscape:
			rep.action = replayQuit
		case "p", "P", KeySpace:
			rep.action = replayTogglePause
		case "+", ">":
			rep.action = replaySpeedMore
		case "-", "<":
			rep.action = replaySpeedLess
		case KeyArrowRight, KeyArrowDown, KeyEnter, "j", "n", "f":
			rep.action = replayNext
			rep.auto = false
		case KeyArrowLeft, KeyArrowUp, KeyBackspace, "k", "N", "b":
			rep.action = replayPrevious
			rep.auto = false
		}
	case MsgMouseDown:
		switch msg.Button {
		case ButtonMain:
			rep.action = replayTogglePause
		case ButtonAuxiliary:
			rep.action = replayNext
			rep.action = replayTogglePause
			rep.auto = false
		case ButtonSecondary:
			rep.action = replayPrevious
			rep.auto = false
		}
	case msgTick:
		if rep.auto && rep.frame == int(msg) {
			rep.action = replayNext
		}
	}
	switch rep.action {
	case replayNext:
		if rep.frame >= len(rep.Frames) {
			rep.action = replayNone
			break
		} else if rep.frame < 0 {
			rep.frame = 0
		}
		rep.frame++
	case replayPrevious:
		if rep.frame <= 1 {
			rep.action = replayNone
			break
		} else if rep.frame >= len(rep.Frames) {
			rep.frame = len(rep.Frames)
		}
		rep.frame--
	case replayQuit:
		return Quit
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
	return rep.tick()
}

func (rep *replay) Draw(gd *Grid) {
	switch rep.action {
	case replayNext:
		df := rep.Frames[rep.frame-1]
		rep.undo = append(rep.undo, []FrameCell{})
		j := len(rep.undo) - 1
		for _, dr := range df.Cells {
			i := gd.getIdx(dr.Pos)
			c := gd.cellBuffer[i]
			rep.undo[j] = append(rep.undo[j], FrameCell{Cell: c, Pos: dr.Pos})
			gd.SetCell(dr.Pos, dr.Cell)
		}
	case replayPrevious:
		df := rep.undo[len(rep.undo)-1]
		for _, dr := range df {
			gd.SetCell(dr.Pos, dr.Cell)
		}
		rep.undo = rep.undo[:len(rep.undo)-1]
	}
}

func (rep *replay) tick() Cmd {
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
