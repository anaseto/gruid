package gorltk

import "time"

type Replay struct {
	Grid     *Grid
	Driver   Driver
	Frames   []Frame
	undo     [][]FrameCell
	frame    int
	auto     bool
	speed    time.Duration
	evch     chan repEvent
	color256 bool
}

type repEvent int

const (
	ReplayNext repEvent = iota
	ReplayPrevious
	ReplayTogglePause
	ReplayQuit
	ReplaySpeedMore
	ReplaySpeedLess
)

func (rep *Replay) Run() {
	rep.auto = true
	rep.speed = 1
	rep.evch = make(chan repEvent, 100)
	rep.undo = [][]FrameCell{}
	go func(r *Replay) {
		r.pollKeyboardEvents()
	}(rep)
	for {
		e := rep.pollEvent()
		switch e {
		case ReplayNext:
			if rep.frame >= len(rep.Frames) {
				break
			} else if rep.frame < 0 {
				rep.frame = 0
			}
			rep.drawFrame()
			rep.frame++
		case ReplayPrevious:
			if rep.frame <= 1 {
				break
			} else if rep.frame >= len(rep.Frames) {
				rep.frame = len(rep.Frames)
			}
			rep.frame--
			rep.undoFrame()
		case ReplayQuit:
			return
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
	}
}

func (rep *Replay) drawFrame() {
	df := rep.Frames[rep.frame]
	rep.undo = append(rep.undo, []FrameCell{})
	j := len(rep.undo) - 1
	for _, dr := range df.Cells {
		i := rep.Grid.GetIndex(dr.Pos)
		c := rep.Grid.cellBuffer[i]
		//if rep.color256 {
		//dr.Cell.Fg = rep.ui.Map16ColorTo256(dr.Cell.Fg)
		//dr.Cell.Bg = rep.ui.Map16ColorTo256(dr.Cell.Bg)
		//} else {
		//dr.Cell.Bg = ui.Map256ColorTo16(dr.Cell.Bg)
		//dr.Cell.Fg = ui.Map256ColorTo16(dr.Cell.Fg)
		//}
		rep.undo[j] = append(rep.undo[j], FrameCell{Cell: c, Pos: dr.Pos})
		rep.Grid.SetCell(dr.Pos, dr.Cell)
	}
	rep.Grid.Draw()
	rep.Driver.Flush(rep.Grid)
	//rep.Grid.frame = nil
}

func (rep *Replay) undoFrame() {
	df := rep.undo[len(rep.undo)-1]
	for _, dr := range df {
		rep.Grid.SetCell(dr.Pos, dr.Cell)
	}
	rep.undo = rep.undo[:len(rep.undo)-1]
	rep.Grid.Draw()
	rep.Driver.Flush(rep.Grid)
	//rep.Grid.frame = nil
}

func (rep *Replay) pollEvent() (in repEvent) {
	if rep.auto && rep.frame <= len(rep.Frames)-1 && rep.frame >= 0 {
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
		t := time.NewTimer(d)
		select {
		case in = <-rep.evch:
		case <-t.C:
			in = ReplayNext
		}
		t.Stop()
	} else {
		in = <-rep.evch
	}
	return in
}

func (rep *Replay) pollKeyboardEvents() {
	for {
		ev := rep.Driver.PollMsg()
		switch ev := ev.(type) {
		case MsgInterrupt:
			rep.evch <- ReplayNext
			continue
		case MsgKeyDown:
			switch ev.Key {
			case "Q", "q", KeyEscape:
				rep.evch <- ReplayQuit
				return
			case "p", "P", KeySpace:
				rep.evch <- ReplayTogglePause
			case "+", ">":
				rep.evch <- ReplaySpeedMore
			case "-", "<":
				rep.evch <- ReplaySpeedLess
			case KeyArrowRight, KeyArrowDown, "j", "n", "f":
				rep.evch <- ReplayNext
			case KeyArrowLeft, KeyArrowUp, "k", "N", "b":
				rep.evch <- ReplayPrevious
			}
		case MsgMouseDown:
			switch ev.Button {
			case ButtonMain:
				rep.evch <- ReplayNext
			case ButtonAuxiliary:
				rep.evch <- ReplayTogglePause
			case ButtonSecondary:
				rep.evch <- ReplayPrevious
			}
		}
	}
}
