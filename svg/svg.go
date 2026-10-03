// Package svg renders a show as SVG, one document per frame, on a virtual
// clock. Nothing depends on wall-clock time, so two renders of a show are
// identical. Tests assert on the frames; Main writes them into a
// self-playing HTML page that capture/capture.cjs turns into video.
package svg

import (
	"context"
	"fmt"
	"strings"

	"github.com/lestrrat-go/kavad"
)

// Frames plays a fresh run of show and returns one SVG document per frame,
// fps frames a second for the given number of seconds.
func Frames(ctx context.Context, show kavad.Show, fps, seconds int) ([]string, error) {
	r, err := show.Start(ctx)
	if err != nil {
		return nil, err
	}
	w, h := show.Size()
	frames := make([]string, 0, fps*seconds)
	for f := range fps * seconds {
		now := f * 1000 / fps
		if err := r.Advance(ctx, now); err != nil {
			return nil, err
		}
		frames = append(frames, Frame(r, w, h, now))
	}
	return frames, nil
}

// Frame renders r at time now (ms) as a w×h SVG document.
func Frame(r *kavad.Run, w, h, now int) string {
	c := &Canvas{}
	c.f(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h, w, h)
	r.Draw(c, now)
	c.f(`</svg>`)
	return c.b.String()
}

// Canvas is the kavad.Canvas that writes SVG elements.
type Canvas struct{ b strings.Builder }

var _ kavad.Canvas = (*Canvas)(nil)

// String returns the elements written so far.
func (c *Canvas) String() string { return c.b.String() }

func (c *Canvas) f(format string, args ...any) { fmt.Fprintf(&c.b, format, args...) }

func paint(attr string, col kavad.Color) string {
	if col.A >= 1 {
		return fmt.Sprintf(`%s="%s"`, attr, col.Hex())
	}
	return fmt.Sprintf(`%s="%s" %s-opacity="%.3f"`, attr, col.Hex(), attr, col.A)
}

func points(pts []kavad.Point) string {
	var b strings.Builder
	for i, p := range pts {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%.1f,%.1f", p.X, p.Y)
	}
	return b.String()
}

func (c *Canvas) FillPolygon(pts []kavad.Point, col kavad.Color) {
	c.f(`<polygon points="%s" %s/>`, points(pts), paint("fill", col))
}

func (c *Canvas) StrokePolyline(pts []kavad.Point, width float64, closed bool, col kavad.Color) {
	el := "polyline"
	if closed {
		el = "polygon"
	}
	c.f(`<%s points="%s" fill="none" %s stroke-width="%.1f" stroke-linecap="round" stroke-linejoin="round"/>`, el, points(pts), paint("stroke", col), width)
}

func (c *Canvas) FillCircle(x, y, r float64, col kavad.Color) {
	c.f(`<circle cx="%.1f" cy="%.1f" r="%.1f" %s/>`, x, y, r, paint("fill", col))
}

func (c *Canvas) StrokeCircle(x, y, r, width float64, col kavad.Color) {
	c.f(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" %s stroke-width="%.1f"/>`, x, y, r, paint("stroke", col), width)
}

// FontFamily is the CSS font stack for each kavad.Font. The web renderer
// uses the same stacks.
var FontFamily = map[kavad.Font]string{
	kavad.Sans:     `"DejaVu Sans", "Liberation Sans", sans-serif`,
	kavad.SansBold: `"DejaVu Sans", "Liberation Sans", sans-serif`,
	kavad.Mono:     `"DejaVu Sans Mono", "Liberation Mono", monospace`,
}

var (
	svgAnchor = map[kavad.Align]string{kavad.AlignStart: "start", kavad.AlignCenter: "middle", kavad.AlignEnd: "end"}
	svgEscape = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
)

func (c *Canvas) DrawText(t kavad.Text) {
	weight := ""
	if t.Font == kavad.SansBold {
		weight = ` font-weight="bold"`
	}
	spacing := ""
	if t.Spacing != 0 {
		spacing = fmt.Sprintf(` letter-spacing="%.1f"`, t.Spacing)
	}
	c.f(`<text x="%.1f" y="%.1f" font-family="%s"%s font-size="%.1f" %s text-anchor="%s"%s xml:space="preserve">%s</text>`,
		t.X, t.Y, svgEscape.Replace(FontFamily[t.Font]), weight, t.Size, paint("fill", t.Color), svgAnchor[t.Align], spacing, svgEscape.Replace(t.S))
}
