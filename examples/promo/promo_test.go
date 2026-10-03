package promo_test

import (
	"math"
	"strings"
	"testing"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/examples/promo"
	"github.com/lestrrat-go/kavad/examples/promo/stage"
	"github.com/lestrrat-go/kavad/svg"
	"github.com/stretchr/testify/require"
)

// The storyboard's timing comes entirely from the events the machine
// schedules, so the scene cuts land exactly where the document says.
func TestTimeline(t *testing.T) {
	ctx := t.Context()
	screen := stage.New()
	r, err := promo.StartOn(ctx, screen)
	require.NoError(t, err)
	require.NoError(t, r.Advance(ctx, 31000))
	require.True(t, r.Done())
	require.Equal(t, []kavad.Change{
		{At: 0, Scene: "intro"},
		{At: 4000, Scene: "states"},
		{At: 9000, Scene: "events"},
		{At: 15000, Scene: "snippets"},
		{At: 20000, Scene: "verify"},
		{At: 25000, Scene: "outro"},
		{At: 30000, Scene: "end"},
	}, screen.History())
}

// Rendering is deterministic: two runs produce identical frames, and the
// frames show what the storyboard says.
func TestFrames(t *testing.T) {
	a, err := svg.Frames(t.Context(), promo.Show{}, 10, 30)
	require.NoError(t, err)
	b, err := svg.Frames(t.Context(), promo.Show{}, 10, 30)
	require.NoError(t, err)
	require.Equal(t, a, b)

	require.Len(t, a, 300)
	for _, f := range []string{a[0], a[150], a[299]} {
		require.True(t, strings.HasPrefix(f, `<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720"`), f[:min(len(f), 80)])
		require.True(t, strings.HasSuffix(f, "</svg>"))
	}
	require.Contains(t, a[170], "Logic is plain Go.")
}

// recorder is a kavad.Canvas that checks every call a renderer would get.
type recorder struct {
	t     *testing.T
	calls int
	texts map[string]bool
}

func (r *recorder) point(x, y float64) {
	require.False(r.t, math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0), "non-finite point")
}

func (r *recorder) color(c kavad.Color) {
	r.calls++
	require.True(r.t, c.A > 0 && c.A <= 1, "alpha %v reached the canvas", c.A)
}

func (r *recorder) FillPolygon(pts []kavad.Point, c kavad.Color) {
	r.color(c)
	require.GreaterOrEqual(r.t, len(pts), 3)
	for _, p := range pts {
		r.point(p.X, p.Y)
	}
}

func (r *recorder) StrokePolyline(pts []kavad.Point, width float64, _ bool, c kavad.Color) {
	r.color(c)
	require.GreaterOrEqual(r.t, len(pts), 2)
	require.Positive(r.t, width)
	for _, p := range pts {
		r.point(p.X, p.Y)
	}
}

func (r *recorder) FillCircle(x, y, rad float64, c kavad.Color) {
	r.color(c)
	r.point(x, y)
	require.Positive(r.t, rad)
}

func (r *recorder) StrokeCircle(x, y, rad, width float64, c kavad.Color) {
	r.color(c)
	r.point(x, y)
	require.Positive(r.t, rad)
	require.Positive(r.t, width)
}

func (r *recorder) DrawText(t kavad.Text) {
	r.color(t.Color)
	r.point(t.X, t.Y)
	require.NotEmpty(r.t, t.S)
	r.texts[t.S] = true
}

func (r *recorder) DrawImage(img *kavad.Image, x, y, w, h, opacity float64) {
	r.calls++
	require.NotNil(r.t, img)
	r.point(x, y)
	r.point(w, h)
	require.Positive(r.t, w)
	require.Positive(r.t, h)
	require.Greater(r.t, opacity, 0.0)
	require.LessOrEqual(r.t, opacity, 1.0)
}

// Every moment of the promo, played through a kavad.Loop the way the desktop
// and web runners play it, hands the canvas only finite, visible primitives.
func TestDrawOnCanvas(t *testing.T) {
	ctx := t.Context()
	loop, err := kavad.NewLoop(ctx, promo.Show{}, -1)
	require.NoError(t, err)

	rec := &recorder{t: t, texts: map[string]bool{}}
	for now := 0; now <= 30000; now += 50 {
		require.NoError(t, loop.Update(ctx, now))
		loop.Draw(rec)
	}
	require.True(t, loop.Run().Done())
	require.Positive(t, rec.calls)
	for _, s := range []string{"STATE MACHINES FOR GO", "event: buy", "quest.fsm.json: ok", "github.com/lestrrat-go/fsm"} {
		require.True(t, rec.texts[s], "never drew %q", s)
	}
}
