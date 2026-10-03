package draw_test

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/draw"
	"github.com/lestrrat-go/kavad/svg"
	"github.com/stretchr/testify/require"
)

var white = kavad.Color{R: 255, G: 255, B: 255, A: 1}

func TestPen(t *testing.T) {
	c := &svg.Canvas{}
	p := draw.New(c).Fade(0.5).Shift(10, 20)
	p.Circle(1, 2, 3, white)
	p.Fade(0).Circle(1, 2, 3, white)      // invisible: dropped
	p.Rect(0, 0, 4, 5, white)             // shifted and faded
	p.Stroke([]kavad.Point{{}}, 2, white) // one point: dropped
	require.Equal(t,
		`<circle cx="11.0" cy="22.0" r="3.0" fill="#ffffff" fill-opacity="0.500"/>`+
			`<polygon points="10.0,20.0 14.0,20.0 14.0,25.0 10.0,25.0" fill="#ffffff" fill-opacity="0.500"/>`,
		c.String())
}

func TestPenImage(t *testing.T) {
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 2, 1))))
	img, err := kavad.LoadPNG(fstest.MapFS{"a.png": {Data: encoded.Bytes()}}, "a.png")
	require.NoError(t, err)
	c := &svg.Canvas{}
	draw.New(c).Fade(0.5).Shift(10, 20).Image(img, 0, 0, 4, 4, 0.6)
	require.Contains(t, c.String(), `x="10.00" y="21.00" width="4.00" height="2.00" opacity="0.300"`)
}

func TestTrim(t *testing.T) {
	line := []kavad.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}}
	require.Nil(t, draw.Trim(line, 0))
	require.Equal(t, line, draw.Trim(line, 1))
	require.Equal(t, []kavad.Point{{X: 0, Y: 0}, {X: 5, Y: 0}}, draw.Trim(line, 0.25))
	require.Equal(t, []kavad.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 5}}, draw.Trim(line, 0.75))
}

func TestQuadEnds(t *testing.T) {
	a, c, b := kavad.Point{X: 0, Y: 0}, kavad.Point{X: 50, Y: 100}, kavad.Point{X: 100, Y: 0}
	pts := draw.Quad(a, c, b)
	require.Len(t, pts, 33)
	require.Equal(t, a, pts[0])
	require.Equal(t, b, pts[32])
	require.Equal(t, kavad.Point{X: 50, Y: 50}, pts[16])
}
