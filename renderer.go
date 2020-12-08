package gruid

import (
	"context"
	"time"
)

type renderer struct {
	driver     Driver
	rec        bool
	fps        time.Duration
	framerec   []Frame
	frames     chan Frame
	ticker     *time.Ticker
	done       chan bool
	grid       Grid
	frameQueue []Frame
}

// Init initializes the renderer.
func (r *renderer) Init() {
	if r.fps <= 60 {
		// Use always at least 60 FPS.
		r.fps = 60
	}
	if r.fps >= 240 {
		// More than 240 FPS does not make any sense.
		r.fps = 240
	}
	if r.ticker == nil {
		r.ticker = time.NewTicker(time.Second / r.fps)
	}
	r.done = make(chan bool)

	// We buffer a few frames so that very fast input (such as mouse
	// motion) does not produce a drag in display.
	r.frames = make(chan Frame, 4)
}

// Listen waits for drawing ticks or a message for the end of rendering.
func (r *renderer) Listen(ctx context.Context) {
	for {
		select {
		case <-r.ticker.C:
			r.flush()
		case <-ctx.Done():
			r.ticker.Stop()
			r.flush() // draw remaining frame if any
			close(r.done)
			return
		}
	}
}

func (r *renderer) flush() {
	r.frameQueue = r.frameQueue[:0]
	for {
		select {
		case frame := <-r.frames:
			r.frameQueue = append(r.frameQueue, frame)
		default:
			if len(r.frameQueue) == 0 {
				// no new frame for this tick
				return
			}
			if r.rec {
				r.framerec = append(r.framerec, r.frameQueue...)
			}
			if len(r.frameQueue) == 1 {
				r.driver.Flush(r.frameQueue[0])
				return
			}

			// We combine several frames into one before sending to
			// the driver.
			if r.grid.ug == nil {
				r.grid = NewGrid(GridConfig{
					Width:  r.frameQueue[0].Width,
					Height: r.frameQueue[0].Height,
				})
			} else {
				r.grid.ClearCache()
				for i := 0; i < len(r.grid.ug.cellBuffer); i++ {
					r.grid.ug.cellBuffer[i] = Cell{}
				}
			}
			for _, frame := range r.frameQueue {
				r.grid.Resize(frame.Width, frame.Height)
				ug := r.grid.ug
				for _, c := range frame.Cells {
					ug.cellBuffer[c.Pos.X+ug.width*c.Pos.Y] = c.Cell
				}
			}
			frame := r.grid.ComputeFrame()
			r.driver.Flush(frame)
			return
		}
	}
}
