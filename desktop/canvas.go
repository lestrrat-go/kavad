package desktop

import (
	"bytes"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/lestrrat-go/kavad"
)

// canvas is the kavad.Canvas that draws on an Ebitengine image: polygons and
// polylines become vector paths, text uses the Go fonts.
type canvas struct {
	dst    *ebiten.Image
	fonts  map[kavad.Font]*text.GoTextFaceSource
	faces  map[faceKey]*text.GoTextFace
	images map[string]*ebiten.Image
}

var _ kavad.Canvas = (*canvas)(nil)

type faceKey struct {
	font kavad.Font
	size float64
}

func newCanvas() (*canvas, error) {
	c := &canvas{
		fonts: map[kavad.Font]*text.GoTextFaceSource{}, faces: map[faceKey]*text.GoTextFace{},
		images: map[string]*ebiten.Image{},
	}
	for f, ttf := range map[kavad.Font][]byte{kavad.Sans: goregular.TTF, kavad.SansBold: gobold.TTF, kavad.Mono: gomono.TTF} {
		src, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
		if err != nil {
			return nil, err
		}
		c.fonts[f] = src
	}
	return c, nil
}

func (c *canvas) preload(images []*kavad.Image) {
	for _, img := range images {
		if _, ok := c.images[img.ID()]; ok {
			continue
		}
		c.images[img.ID()] = ebiten.NewImageFromImage(img.Pixels())
	}
}

func nrgba(c kavad.Color) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: uint8(c.A*255 + 0.5)}
}

func drawOpts(c kavad.Color) *vector.DrawPathOptions {
	o := &vector.DrawPathOptions{AntiAlias: true}
	o.ColorScale.ScaleWithColor(nrgba(c))
	return o
}

func path(pts []kavad.Point, closed bool) *vector.Path {
	var p vector.Path
	for i, q := range pts {
		if i == 0 {
			p.MoveTo(float32(q.X), float32(q.Y))
		} else {
			p.LineTo(float32(q.X), float32(q.Y))
		}
	}
	if closed {
		p.Close()
	}
	return &p
}

func (c *canvas) FillPolygon(pts []kavad.Point, col kavad.Color) {
	vector.FillPath(c.dst, path(pts, true), nil, drawOpts(col))
}

func (c *canvas) StrokePolyline(pts []kavad.Point, width float64, closed bool, col kavad.Color) {
	vector.StrokePath(c.dst, path(pts, closed), &vector.StrokeOptions{
		Width:    float32(width),
		LineCap:  vector.LineCapRound,
		LineJoin: vector.LineJoinRound,
	}, drawOpts(col))
}

func (c *canvas) FillCircle(x, y, r float64, col kavad.Color) {
	vector.FillCircle(c.dst, float32(x), float32(y), float32(r), nrgba(col), true)
}

func (c *canvas) StrokeCircle(x, y, r, width float64, col kavad.Color) {
	// A closed arc path: vector.StrokeCircle leaves a seam where the arc
	// starts and ends.
	var p vector.Path
	p.Arc(float32(x), float32(y), float32(r), 0, 2*math.Pi, vector.Clockwise)
	p.Close()
	vector.StrokePath(c.dst, &p, &vector.StrokeOptions{Width: float32(width)}, drawOpts(col))
}

func (c *canvas) face(f kavad.Font, size float64) *text.GoTextFace {
	k := faceKey{f, size}
	if face, ok := c.faces[k]; ok {
		return face
	}
	face := &text.GoTextFace{Source: c.fonts[f], Size: size}
	if len(c.faces) > 512 { // animated sizes: keep the cache bounded
		clear(c.faces)
	}
	c.faces[k] = face
	return face
}

func (c *canvas) DrawText(t kavad.Text) {
	face := c.face(t.Font, t.Size)
	// stage positions text by its baseline; Ebitengine by the line's top.
	y := t.Y - face.Metrics().HAscent
	if t.Spacing == 0 {
		o := &text.DrawOptions{}
		o.GeoM.Translate(t.X, y)
		o.ColorScale.ScaleWithColor(nrgba(t.Color))
		o.PrimaryAlign = map[kavad.Align]text.Align{kavad.AlignStart: text.AlignStart, kavad.AlignCenter: text.AlignCenter, kavad.AlignEnd: text.AlignEnd}[t.Align]
		text.Draw(c.dst, t.S, face, o)
		return
	}
	// Letter-spaced text: place each character, aligning the whole run.
	width := 0.0
	for _, r := range t.S {
		width += text.Advance(string(r), face) + t.Spacing
	}
	x := t.X
	switch t.Align {
	case kavad.AlignCenter:
		x -= width / 2
	case kavad.AlignEnd:
		x -= width
	}
	for _, r := range t.S {
		o := &text.DrawOptions{}
		o.GeoM.Translate(x, y)
		o.ColorScale.ScaleWithColor(nrgba(t.Color))
		text.Draw(c.dst, string(r), face, o)
		x += text.Advance(string(r), face) + t.Spacing
	}
}

func (c *canvas) DrawImage(img *kavad.Image, x, y, w, h, opacity float64) {
	if opacity <= 0 {
		return
	}
	x, y, w, h = img.Fit(x, y, w, h)
	if w == 0 || h == 0 {
		return
	}
	bitmap, ok := c.images[img.ID()]
	if !ok {
		panic("desktop: image " + img.ID() + " was not preloaded")
	}
	iw, ih := img.Size()
	opts := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	opts.GeoM.Scale(w/float64(iw), h/float64(ih))
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleAlpha(float32(min(opacity, 1)))
	c.dst.DrawImage(bitmap, opts)
}
