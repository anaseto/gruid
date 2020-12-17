package gruid

import "time"

// Animation allows to request simple animation frames, and draw them on
// schedule or cancel them before.
type Animation struct {
	afs []animFrame
}

type animFrame struct {
	t    time.Time
	grid Grid
	fn   func(Grid)
}

// Cancel removes any scheduled animation frames. It should be generally used
// on keyboard or mouse button input.
func (a Animation) Cancel() {
	a.afs = nil
}

// RequestFrame schedules the animation drawing specified by a function onto a
// given grid, and with a given additional delay starting after previously
// scheduled animations terminate.
func (a Animation) RequestFrame(d time.Duration, gd Grid, fn func(Grid)) {
	var ot time.Time
	if len(a.afs) == 0 {
		ot = time.Now()
	} else {
		ot = a.afs[len(a.afs)-1].t
	}
	a.afs = append(a.afs, animFrame{t: ot.Add(d), grid: gd, fn: fn})
}

// Draw sends animation frames up to the animation's time. It returns true if
// the animation has drawn animation frames, or if it still has scheduled
// animation frames.
func (a Animation) Draw() bool {
	t := time.Now()
	draw := false
	if len(a.afs) > 0 {
		af := a.afs[0]
		if af.t.Before(t) {
			af.fn(af.grid)
			a.afs = a.afs[1:]
			draw = true
		}
	}
	if len(a.afs) == 0 {
		a.afs = nil
	}
	return draw || len(a.afs) > 0
}
