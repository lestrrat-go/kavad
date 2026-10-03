package kavad

import (
	"fmt"
	"strconv"
)

// Canvas is what a show draws on. It is deliberately small (filled and
// stroked polygons, circles and text) so each renderer is a page of code.
// Coordinates are in the show's own space (Show.Size); renderers scale it to
// their output. Text is positioned by its baseline.
type Canvas interface {
	// FillPolygon fills the closed polygon pts.
	FillPolygon(pts []Point, c Color)
	// StrokePolyline strokes pts with round caps and joins, closing the
	// shape if closed is true.
	StrokePolyline(pts []Point, width float64, closed bool, c Color)
	FillCircle(x, y, r float64, c Color)
	StrokeCircle(x, y, r, width float64, c Color)
	DrawText(t Text)
}

// Point is a position on the canvas.
type Point struct{ X, Y float64 }

// Color is an sRGB colour with straight (not premultiplied) alpha in 0..1.
type Color struct {
	R, G, B uint8
	A       float64
}

// Hex returns the colour as #rrggbb, without alpha.
func (c Color) Hex() string {
	const digits = "0123456789abcdef"
	b := []byte{'#', 0, 0, 0, 0, 0, 0}
	for i, v := range []uint8{c.R, c.G, c.B} {
		b[1+2*i], b[2+2*i] = digits[v>>4], digits[v&15]
	}
	return string(b)
}

// ParseHex parses an opaque colour written as #rrggbb.
func ParseHex(s string) (Color, error) {
	if len(s) != 7 || s[0] != '#' {
		return Color{}, fmt.Errorf("kavad: colour %q is not #rrggbb", s)
	}
	v, err := strconv.ParseUint(s[1:], 16, 32)
	if err != nil {
		return Color{}, fmt.Errorf("kavad: colour %q is not #rrggbb", s)
	}
	return Color{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 1}, nil
}

// MustHex is ParseHex for palette declarations; it panics on a malformed
// colour.
func MustHex(s string) Color {
	c, err := ParseHex(s)
	if err != nil {
		panic(err)
	}
	return c
}

// Font selects a typeface. Renderers pick the concrete fonts.
type Font int

const (
	Sans Font = iota
	SansBold
	Mono
)

// Align is the horizontal anchor of a Text.
type Align int

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
)

// Text is one run of text with its baseline at Y.
type Text struct {
	X, Y  float64
	S     string
	Font  Font
	Size  float64
	Color Color
	Align Align
	// Spacing is extra space after each character, in canvas units.
	Spacing float64
}
