package gruid

import (
	"context"
)

type renderer struct {
	driver     Driver
	frames     chan Frame
	done       chan struct{}
	frameQueue []Frame
	grid       Grid
	enc        *frameEncoder
}

// ListenAndRender waits for frames and sends them to the driver.
func (r *renderer) ListenAndRender(ctx context.Context) {
	r.done = make(chan struct{})
	r.frames = make(chan Frame, 4) // buffered
	for {
		r.frameQueue = r.frameQueue[:0]
		select {
		case frame := <-r.frames:
			r.frameQueue = append(r.frameQueue, frame)
			r.flushFrames()
		case <-ctx.Done():
			r.flushFrames()
			if r.enc != nil {
				r.enc.gzw.Close()
			}
			close(r.done)
			return
		}
	}
}

func (r *renderer) flushFrames() {
	for {
		select {
		case frame := <-r.frames:
			r.frameQueue = append(r.frameQueue, frame)
		default:
			if len(r.frameQueue) == 1 {
				// the most common case
				r.driver.Flush(r.frameQueue[0])
				if r.enc != nil {
					r.enc.encode(r.frameQueue[0]) // XXX: report errors ?
				}
				return
			}
			if len(r.frameQueue) == 0 {
				// no remaining frame (when ctx.Done())
				return
			}

			// We combine several frames into one before sending to
			// the driver. This should not happen at 60 fps in
			// practice, but it might happen at higher fps with
			// slow drivers.
			if r.grid.ug == nil {
				r.grid = NewGrid(GridConfig{
					Width:  r.frameQueue[0].Width,
					Height: r.frameQueue[0].Height,
				})
			}
			r.grid.ClearCache()
			for i := 0; i < len(r.grid.ug.cellBuffer); i++ {
				r.grid.ug.cellBuffer[i] = Cell{}
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
			if r.enc != nil {
				for _, fr := range r.frameQueue {
					r.enc.encode(fr)
				}
			}
			return
		}
	}
}
