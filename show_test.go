package kavad_test

import (
	"context"
	"testing"

	"github.com/lestrrat-go/kavad"
	"github.com/stretchr/testify/require"
)

// countdown is a stand-in machine: it shows scene "a", then "b" when it gets
// "next", and is done after the second "next".
type countdown struct {
	st   *kavad.Stage
	sent int
}

func (m *countdown) Send(_ context.Context, event string) error {
	if event != "next" {
		return nil
	}
	m.sent++
	switch m.sent {
	case 1:
		m.st.Scene("b")
		m.st.Schedule("next", 500)
	case 2:
		m.st.Scene("end")
	}
	return nil
}

func (m *countdown) Done() bool { return m.sent >= 2 }

// painter records the moments it is asked to paint.
type painter struct{ at []int }

func (p *painter) Paint(_ kavad.Canvas, now int) { p.at = append(p.at, now) }

type show struct {
	starts  int
	painter *painter
}

func (s *show) Size() (int, int) { return 160, 90 }

func (s *show) Start(context.Context) (*kavad.Run, error) {
	s.starts++
	st := kavad.NewStage()
	st.Scene("a")
	st.Schedule("next", 1000)
	return kavad.NewRun(&countdown{st: st}, st, s.painter), nil
}

func TestRun(t *testing.T) {
	ctx := t.Context()
	s := &show{painter: &painter{}}
	r, err := s.Start(ctx)
	require.NoError(t, err)

	require.NoError(t, r.Advance(ctx, 999))
	require.False(t, r.Done())
	require.NoError(t, r.Advance(ctx, 1500))
	require.True(t, r.Done())
	require.Equal(t, []kavad.Change{{At: 0, Scene: "a"}, {At: 1000, Scene: "b"}, {At: 1500, Scene: "end"}}, r.Stage().History())

	r.Draw(nil, 1500)
	require.Equal(t, []int{1500}, s.painter.at)
}

// The loop holds the last moment for the hold time, then restarts with run
// time back at 0.
func TestLoopRestarts(t *testing.T) {
	ctx := t.Context()
	s := &show{painter: &painter{}}
	l, err := kavad.NewLoop(ctx, s, 300)
	require.NoError(t, err)
	require.Equal(t, 1, s.starts)

	require.NoError(t, l.Update(ctx, 1500)) // finishes at 1500
	require.True(t, l.Run().Done())
	require.NoError(t, l.Update(ctx, 1799))
	require.Equal(t, 1, s.starts)
	require.Equal(t, 1799, l.Now())

	require.NoError(t, l.Update(ctx, 1800)) // hold is over
	require.Equal(t, 2, s.starts)
	require.False(t, l.Run().Done())
	require.Equal(t, 0, l.Now())

	require.NoError(t, l.Update(ctx, 1800+1000))
	l.Draw(nil)
	require.Equal(t, []int{1000}, s.painter.at)
	scene, _ := l.Run().Stage().Current()
	require.Equal(t, "b", scene)
}

func TestLoopNegativeHoldNeverRestarts(t *testing.T) {
	ctx := t.Context()
	s := &show{painter: &painter{}}
	l, err := kavad.NewLoop(ctx, s, -1)
	require.NoError(t, err)
	require.NoError(t, l.Update(ctx, 1500))
	require.NoError(t, l.Update(ctx, 100000))
	require.Equal(t, 1, s.starts)
	require.Equal(t, 100000, l.Now())
}
