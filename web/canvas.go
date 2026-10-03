//go:build js && wasm

package web

import (
	"math"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/svg"
)

// canvas is the kavad.Canvas that draws on an HTML <canvas> through its 2D
// context. Each polygon or polyline becomes one Path2D built from an SVG
// path string, so a shape costs a few JavaScript calls, not one per point.
type canvas struct {
	el     js.Value // the <canvas> element
	ctx    js.Value // its CanvasRenderingContext2D
	path2D js.Value // the Path2D constructor
	w, h   float64  // the show's size
}

var _ kavad.Canvas = (*canvas)(nil)

// newCanvas sets el's drawing buffer to w×h, which gives it the show's
// aspect ratio, and the CSS variable --kavad-aspect to w/h, for page styles
// that size the canvas by height.
func newCanvas(el js.Value, w, h int) *canvas {
	el.Set("width", w)
	el.Set("height", h)
	el.Get("style").Call("setProperty", "--kavad-aspect", strconv.FormatFloat(float64(w)/float64(h), 'f', 4, 64))
	return &canvas{
		el:     el,
		ctx:    el.Call("getContext", "2d"),
		path2D: js.Global().Get("Path2D"),
		w:      float64(w),
		h:      float64(h),
	}
}

// begin sizes the drawing buffer to the element's on-screen width times the
// device pixel ratio, at the show's aspect ratio, and maps the show's space
// onto it, so the picture stays sharp at any embed size.
func (c *canvas) begin() {
	dpr := js.Global().Get("devicePixelRatio").Float()
	if dpr <= 0 {
		dpr = 1
	}
	w := c.el.Get("clientWidth").Float()
	if w <= 0 {
		w = c.w
	}
	pw, ph := int(math.Round(w*dpr)), int(math.Round(w*dpr*c.h/c.w))
	if c.el.Get("width").Int() != pw || c.el.Get("height").Int() != ph {
		c.el.Set("width", pw)
		c.el.Set("height", ph)
	}
	s := float64(pw) / c.w
	c.ctx.Call("setTransform", s, 0, 0, s, 0, 0)
}

func rgba(col kavad.Color) string {
	return "rgba(" + strconv.Itoa(int(col.R)) + "," + strconv.Itoa(int(col.G)) + "," +
		strconv.Itoa(int(col.B)) + "," + strconv.FormatFloat(col.A, 'f', 3, 64) + ")"
}

func (c *canvas) path(pts []kavad.Point, closed bool) js.Value {
	var b strings.Builder
	for i, q := range pts {
		if i == 0 {
			b.WriteString("M")
		} else {
			b.WriteString(" L")
		}
		b.WriteString(strconv.FormatFloat(q.X, 'f', 1, 64))
		b.WriteByte(' ')
		b.WriteString(strconv.FormatFloat(q.Y, 'f', 1, 64))
	}
	if closed {
		b.WriteString(" Z")
	}
	return c.path2D.New(b.String())
}

func (c *canvas) stroke(p js.Value, width float64, col kavad.Color) {
	c.ctx.Set("strokeStyle", rgba(col))
	c.ctx.Set("lineWidth", width)
	c.ctx.Set("lineCap", "round")
	c.ctx.Set("lineJoin", "round")
	c.ctx.Call("stroke", p)
}

func (c *canvas) FillPolygon(pts []kavad.Point, col kavad.Color) {
	c.ctx.Set("fillStyle", rgba(col))
	c.ctx.Call("fill", c.path(pts, true))
}

func (c *canvas) StrokePolyline(pts []kavad.Point, width float64, closed bool, col kavad.Color) {
	c.stroke(c.path(pts, closed), width, col)
}

func (c *canvas) circle(x, y, r float64) js.Value {
	p := c.path2D.New()
	p.Call("arc", x, y, r, 0, 2*math.Pi)
	return p
}

func (c *canvas) FillCircle(x, y, r float64, col kavad.Color) {
	c.ctx.Set("fillStyle", rgba(col))
	c.ctx.Call("fill", c.circle(x, y, r))
}

func (c *canvas) StrokeCircle(x, y, r, width float64, col kavad.Color) {
	c.stroke(c.circle(x, y, r), width, col)
}

var cssAlign = map[kavad.Align]string{kavad.AlignStart: "left", kavad.AlignCenter: "center", kavad.AlignEnd: "right"}

func (c *canvas) DrawText(t kavad.Text) {
	font := strconv.FormatFloat(t.Size, 'f', 1, 64) + "px " + svg.FontFamily[t.Font]
	if t.Font == kavad.SansBold {
		font = "bold " + font
	}
	c.ctx.Set("font", font)
	c.ctx.Set("fillStyle", rgba(t.Color))
	c.ctx.Set("textBaseline", "alphabetic")
	if t.Spacing == 0 {
		c.ctx.Set("textAlign", cssAlign[t.Align])
		c.ctx.Call("fillText", t.S, t.X, t.Y)
		return
	}
	// Letter-spaced text: place each character, aligning the whole run.
	// (ctx.letterSpacing would do this, but not every browser has it.)
	c.ctx.Set("textAlign", "left")
	advance := func(r rune) float64 {
		return c.ctx.Call("measureText", string(r)).Get("width").Float() + t.Spacing
	}
	width := 0.0
	for _, r := range t.S {
		width += advance(r)
	}
	x := t.X
	switch t.Align {
	case kavad.AlignCenter:
		x -= width / 2
	case kavad.AlignEnd:
		x -= width
	}
	for _, r := range t.S {
		c.ctx.Call("fillText", string(r), x, t.Y)
		x += advance(r)
	}
}
