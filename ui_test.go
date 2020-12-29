package gruid

import (
	"context"
	"testing"
)

type testModel struct {
	gd    Grid
	count int
	quit  bool
}

const niter = 90

func (tm *testModel) Update(msg Msg) Effect {
	if tm.quit {
		return nil
	}
	switch msg := msg.(type) {
	case MsgInit:
	case MsgKeyDown:
		switch msg.Key {
		case KeyEnter:
			tm.count++
		case KeyEscape:
			tm.quit = true
			return End()
		}
	}
	return nil
}

func (tm *testModel) Draw() Grid {
	if tm.quit {
		return tm.gd.Slice(Range{})
	}
	if tm.count%3 == 0 {
		tm.gd.Fill(Cell{Rune: '0'})
	} else {
		tm.gd.Fill(Cell{Rune: '1'})
	}
	return tm.gd
}

type testDriver struct {
	init   bool
	closed bool
	t      *testing.T
	count  int
}

func (td *testDriver) Init() error {
	td.init = true
	return nil
}

func (td *testDriver) PollMsgs(ctx context.Context, msgs chan<- Msg) error {
	count := 0
	for {
		msg := MsgKeyDown{Key: KeyEnter}
		if count == niter {
			msg = MsgKeyDown{Key: KeyEscape}
		}
		select {
		case msgs <- msg:
		case <-ctx.Done():
			return nil
		}
		count++
	}
}

func (td *testDriver) Flush(fr Frame) {
	td.count++
	if len(fr.Cells) != 8*4 {
		td.t.Errorf("bad frame.Cells length: %d (expected %d)", len(fr.Cells), 8*4)
	}
}

func (td *testDriver) Close() {
	td.closed = true
}

func TestApp(t *testing.T) {
	gd := NewGrid(8, 4)
	m := &testModel{gd: gd}
	td := &testDriver{t: t}
	app := NewApp(AppConfig{
		Driver: td,
		Model:  m,
	})
	if err := app.Start(context.Background()); err != nil {
		t.Errorf("Start returns error: %v", err)
	}
	if m.count != niter {
		t.Errorf("bad count: %d", m.count)
	}
	if td.count != 1+2*niter/3 {
		t.Errorf("bad driver count: %d", td.count)
	}
	m.gd.Iter(func(p Point, c Cell) {
		if c.Rune != '0' {
			t.Errorf("bad rune: %c", c.Rune)
		}
	})
	if !td.closed || !td.init {
		t.Errorf("not closed or not init")
	}
}

func TestApp2(t *testing.T) {
	gd := NewGrid(8, 4)
	m := &testModel{gd: gd}
	m.count += 2
	app := NewApp(AppConfig{
		Driver: &testDriver{t: t},
		Model:  m,
	})
	if err := app.Start(context.Background()); err != nil {
		t.Errorf("Start returns error: %v", err)
	}
	if m.count != niter+2 {
		t.Errorf("bad count: %d", m.count)
	}
	m.gd.Iter(func(p Point, c Cell) {
		if c.Rune != '1' {
			t.Errorf("bad rune: %c", c.Rune)
		}
	})
}
