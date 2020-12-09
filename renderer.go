package gruid

import (
	"context"
)

type renderer struct {
	driver Driver
	frames chan Frame
	done   chan bool
}

// ListenAndRender waits for frames and sends them to the driver.
func (r *renderer) ListenAndRender(ctx context.Context) {
	r.done = make(chan bool)
	r.frames = make(chan Frame)
	for {
		select {
		case frame := <-r.frames:
			r.driver.Flush(frame)
		case <-ctx.Done():
			// send any remaining frames
			for {
				select {
				case frame := <-r.frames:
					r.driver.Flush(frame)
				default:
					return
				}
			}
		}
	}
}
