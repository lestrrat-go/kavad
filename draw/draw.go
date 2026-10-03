// Package draw has the drawing helpers scenes share: a Pen that applies
// opacity and an offset to everything it draws, and functions that build
// common outlines as point lists.
package draw

import (
	"math"

	"github.com/lestrrat-go/kavad"
)

// Pen draws on a Canvas with an opacity and an offset applied to everything,
// like an SVG <g opacity transform>. Pens are values: Fade and Shift return a
// new pen and leave the receiver unchanged. Nothing fully transparent and no
// empty shape reaches the canvas.
type Pen struct {
	c      kavad.Canvas
	alpha  float64
	dx, dy float64
}

// New returns an opaque pen with no offset that draws on c.
func New(c kavad.Canvas) Pen { return Pen{c: c, alpha: 1} }

// Fade returns a pen whose opacity is multiplied by a.
func (p Pen) Fade(a float64) Pen { p.alpha *= a; return p }

// Shift returns a pen whose drawing is moved by (dx, dy).
func (p Pen) Shift(dx, dy float64) Pen { p.dx += dx; p.dy += dy; return p }

func (p Pen) col(c kavad.Color) (kavad.Color, bool) {
	c.A *= p.alpha
	return c, c.A > 0.001
}

func (p Pen) pts(pts []kavad.Point) []kavad.Point {
	if p.dx == 0 && p.dy == 0 {
		return pts
	}
	out := make([]kavad.Point, len(pts))
	for i, q := range pts {
		out[i] = kavad.Point{X: q.X + p.dx, Y: q.Y + p.dy}
	}
	return out
}

// Fill fills the closed polygon pts.
func (p Pen) Fill(pts []kavad.Point, c kavad.Color) {
	if c, ok := p.col(c); ok && len(pts) > 2 {
		p.c.FillPolygon(p.pts(pts), c)
	}
}

// Stroke strokes the open polyline pts.
func (p Pen) Stroke(pts []kavad.Point, width float64, c kavad.Color) {
	if c, ok := p.col(c); ok && len(pts) > 1 && width > 0 {
		p.c.StrokePolyline(p.pts(pts), width, false, c)
	}
}

// StrokeClosed strokes the outline of the closed polygon pts.
func (p Pen) StrokeClosed(pts []kavad.Point, width float64, c kavad.Color) {
	if c, ok := p.col(c); ok && len(pts) > 1 && width > 0 {
		p.c.StrokePolyline(p.pts(pts), width, true, c)
	}
}

// Circle fills a circle.
func (p Pen) Circle(x, y, r float64, c kavad.Color) {
	if c, ok := p.col(c); ok && r > 0 {
		p.c.FillCircle(x+p.dx, y+p.dy, r, c)
	}
}

// Ring strokes a circle.
func (p Pen) Ring(x, y, r, width float64, c kavad.Color) {
	if c, ok := p.col(c); ok && r > 0 && width > 0 {
		p.c.StrokeCircle(x+p.dx, y+p.dy, r, width, c)
	}
}

// Text draws t.
func (p Pen) Text(t kavad.Text) {
	c, ok := p.col(t.Color)
	if !ok || t.S == "" || t.Size <= 0 {
		return
	}
	t.Color, t.X, t.Y = c, t.X+p.dx, t.Y+p.dy
	p.c.DrawText(t)
}

// Rect fills an axis-aligned rectangle.
func (p Pen) Rect(x, y, w, h float64, c kavad.Color) {
	p.Fill([]kavad.Point{{X: x, Y: y}, {X: x + w, Y: y}, {X: x + w, Y: y + h}, {X: x, Y: y + h}}, c)
}

// RoundRect returns the outline of a rectangle with corners of radius r.
func RoundRect(x, y, w, h, r float64) []kavad.Point {
	var pts []kavad.Point
	corners := [4][3]float64{
		{x + w - r, y + r, -math.Pi / 2},
		{x + w - r, y + h - r, 0},
		{x + r, y + h - r, math.Pi / 2},
		{x + r, y + r, math.Pi},
	}
	for _, c := range corners {
		for i := range 7 {
			a := c[2] + float64(i)/6*math.Pi/2
			pts = append(pts, kavad.Point{X: c[0] + r*math.Cos(a), Y: c[1] + r*math.Sin(a)})
		}
	}
	return pts
}

// Quad samples the quadratic Bézier curve from a through control point c to
// b as 33 points.
func Quad(a, c, b kavad.Point) []kavad.Point {
	const n = 32
	pts := make([]kavad.Point, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / n
		u := 1 - t
		pts = append(pts, kavad.Point{
			X: u*u*a.X + 2*u*t*c.X + t*t*b.X,
			Y: u*u*a.Y + 2*u*t*c.Y + t*t*b.Y,
		})
	}
	return pts
}

// Trim returns the first fraction f (0..1) of the polyline pts, measured by
// length. Animating f from 0 to 1 makes a line draw itself.
func Trim(pts []kavad.Point, f float64) []kavad.Point {
	if f >= 1 {
		return pts
	}
	if f <= 0 || len(pts) < 2 {
		return nil
	}
	total := 0.0
	for i := 1; i < len(pts); i++ {
		total += math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
	}
	want := total * f
	out := []kavad.Point{pts[0]}
	for i := 1; i < len(pts); i++ {
		seg := math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
		if seg >= want {
			k := want / seg
			return append(out, kavad.Point{
				X: pts[i-1].X + (pts[i].X-pts[i-1].X)*k,
				Y: pts[i-1].Y + (pts[i].Y-pts[i-1].Y)*k,
			})
		}
		want -= seg
		out = append(out, pts[i])
	}
	return out
}

// Rotated returns pts rotated by deg degrees about the origin and moved to
// (x, y).
func Rotated(x, y, deg float64, pts ...kavad.Point) []kavad.Point {
	s, c := math.Sincos(deg * math.Pi / 180)
	out := make([]kavad.Point, len(pts))
	for i, q := range pts {
		out[i] = kavad.Point{X: x + q.X*c - q.Y*s, Y: y + q.X*s + q.Y*c}
	}
	return out
}

// Arrowhead returns an arrowhead at (x, y) pointing along the unit vector
// (dx, dy). Its tip is 8 units ahead of (x, y) and its base 10 units behind.
func Arrowhead(x, y, dx, dy float64) []kavad.Point {
	px, py := -dy, dx
	return []kavad.Point{
		{X: x + dx*8, Y: y + dy*8},
		{X: x - dx*10 + px*9, Y: y - dy*10 + py*9},
		{X: x - dx*10 - px*9, Y: y - dy*10 - py*9},
	}
}
