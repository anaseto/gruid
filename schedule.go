package gruid

import "time"

// Schedule allows to plan consecutive actions, and execute due actions on
// demand, or cancel them before their due date.
//
// It does not belong to core gruid functionality, but may come handy for
// example to plan cancellable animations that animate intermediate model state
// changes during Update. In Update, you would make several calls to After,
// adding functions drawing animation frames. Also, you would call Cancel as
// needed (for example on user keyboard input). Then, in Draw you would call
// Execute and check Done to know whether you should already draw the final
// model state or not.
type Schedule struct {
	sfs []schedfn
}

type schedfn struct {
	t  time.Time
	fn func()
}

// Cancel removes any remaining scheduled actions so that Done returns true.
func (s Schedule) Cancel() {
	s.sfs = nil
}

// Finish completes the scheduled actions early. It returns the number of actions that were performed.
func (s Schedule) Finish() int {
	count := len(s.sfs)
	for _, sf := range s.sfs {
		sf.fn()
	}
	s.sfs = nil
	return count
}

// After adds a planned action given by a function. Its due date is computed as
// an additional delay after the last previously scheduled function due date,
// if any, or after time.Now() otherwise.
func (s Schedule) After(d time.Duration, fn func()) {
	var ot time.Time
	if len(s.sfs) == 0 {
		ot = time.Now()
	} else {
		ot = s.sfs[len(s.sfs)-1].t
	}
	s.sfs = append(s.sfs, schedfn{t: ot.Add(d), fn: fn})
}

// Execute calls in order scheduled functions whose time is already due. It
// returns the number of functions that were called.
func (s Schedule) Execute() int {
	t := time.Now()
	count := 0
	for i, sf := range s.sfs {
		if sf.t.Before(t) {
			sf.fn()
			s.sfs = s.sfs[i+1:]
			count++
		}
	}
	if len(s.sfs) == 0 {
		s.sfs = nil
	}
	return count
}

// Done returns true if there are no more scheduled actions.
func (s Schedule) Done() bool {
	return len(s.sfs) == 0
}
