package svg_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/svg"
	"github.com/stretchr/testify/require"
)

var orange = kavad.Color{R: 0xff, G: 0x6b, B: 0x35, A: 1}

func TestCanvas(t *testing.T) {
	c := &svg.Canvas{}
	c.FillPolygon([]kavad.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}}, orange)
	c.StrokePolyline([]kavad.Point{{X: 0, Y: 0}, {X: 2, Y: 2}}, 3, false, kavad.Color{R: 0xff, G: 0x6b, B: 0x35, A: 0.25})
	c.StrokeCircle(5, 6, 7, 2, orange)
	c.DrawText(kavad.Text{X: 1, Y: 2, S: `a<b & "c"`, Font: kavad.SansBold, Size: 20, Color: orange, Align: kavad.AlignCenter, Spacing: 3})
	require.Equal(t, `<polygon points="0.0,0.0 1.0,0.0 1.0,1.0" fill="#ff6b35"/>`+
		`<polyline points="0.0,0.0 2.0,2.0" fill="none" stroke="#ff6b35" stroke-opacity="0.250" stroke-width="3.0" stroke-linecap="round" stroke-linejoin="round"/>`+
		`<circle cx="5.0" cy="6.0" r="7.0" fill="none" stroke="#ff6b35" stroke-width="2.0"/>`+
		`<text x="1.0" y="2.0" font-family="&quot;DejaVu Sans&quot;, &quot;Liberation Sans&quot;, sans-serif" font-weight="bold" font-size="20.0" fill="#ff6b35" text-anchor="middle" letter-spacing="3.0" xml:space="preserve">a&lt;b &amp; &quot;c&quot;</text>`,
		c.String())
}

// blink is a show without a machine: it paints a circle whose radius is the
// time in seconds, and finishes at once.
type blink struct{}

type done struct{}

func (done) Send(context.Context, string) error { return nil }
func (done) Done() bool                         { return true }

type circle struct{}

func (circle) Paint(c kavad.Canvas, now int) { c.FillCircle(0, 0, float64(now)/1000, orange) }

func (blink) Size() (int, int) { return 64, 36 }
func (blink) Start(context.Context) (*kavad.Run, error) {
	return kavad.NewRun(done{}, kavad.NewStage(), circle{}), nil
}

func TestFramesAndPage(t *testing.T) {
	frames, err := svg.Frames(t.Context(), blink{}, 2, 2)
	require.NoError(t, err)
	require.Equal(t, []string{
		`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="36" viewBox="0 0 64 36"><circle cx="0.0" cy="0.0" r="0.0" fill="#ff6b35"/></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="36" viewBox="0 0 64 36"><circle cx="0.0" cy="0.0" r="0.5" fill="#ff6b35"/></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="36" viewBox="0 0 64 36"><circle cx="0.0" cy="0.0" r="1.0" fill="#ff6b35"/></svg>`,
		`<svg xmlns="http://www.w3.org/2000/svg" width="64" height="36" viewBox="0 0 64 36"><circle cx="0.0" cy="0.0" r="1.5" fill="#ff6b35"/></svg>`,
	}, frames)

	path := filepath.Join(t.TempDir(), "frames.html")
	require.NoError(t, svg.WritePage(path, frames, 2, 64, 36))
	html, err := os.ReadFile(path)
	require.NoError(t, err)
	page := string(html)
	require.Contains(t, page, "const fps = 2;")
	require.Contains(t, page, "window.frameSize = [64, 36];")
	require.Contains(t, page, "aspect-ratio:64/36")
	require.True(t, strings.Contains(page, `r=\"1.5\"`), "frames are embedded as JSON")
}
