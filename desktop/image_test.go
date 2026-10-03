//go:build desktop_pixels

package desktop

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lestrrat-go/kavad"
	"github.com/stretchr/testify/require"
)

func TestCanvasImage(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	pixels.Set(0, 0, color.NRGBA{R: 255, A: 128})
	pixels.Set(1, 0, color.NRGBA{R: 255, A: 128})
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, pixels))
	img, err := kavad.LoadPNG(fstest.MapFS{"a.png": {Data: encoded.Bytes()}}, "a.png")
	require.NoError(t, err)

	c, err := newCanvas()
	require.NoError(t, err)
	c.preload([]*kavad.Image{img, img})
	require.Len(t, c.images, 1)
	g := &imageGame{canvas: c, img: img}
	err = ebiten.RunGame(g)
	require.NoError(t, err)
	require.Zero(t, g.outside.A)
	require.InDelta(t, float64(g.full.A)/2, g.inside.A, 2)
	require.InDelta(t, 255, g.inside.R, 2)
	require.Equal(t, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, g.above)
}

type imageGame struct {
	canvas  *canvas
	img     *kavad.Image
	drawn   bool
	outside color.NRGBA
	full    color.NRGBA
	inside  color.NRGBA
	above   color.NRGBA
}

func (g *imageGame) Update() error {
	if g.drawn {
		return ebiten.Termination
	}
	return nil
}

func (g *imageGame) Draw(*ebiten.Image) {
	g.canvas.dst = ebiten.NewImage(8, 8)
	g.canvas.DrawImage(g.img, 2, 1, 4, 4, 1)
	g.full = color.NRGBAModel.Convert(g.canvas.dst.At(3, 3)).(color.NRGBA)
	g.canvas.dst = ebiten.NewImage(8, 8)
	g.canvas.DrawImage(g.img, 2, 1, 4, 4, 0.5)
	g.outside = color.NRGBAModel.Convert(g.canvas.dst.At(3, 1)).(color.NRGBA)
	g.inside = color.NRGBAModel.Convert(g.canvas.dst.At(3, 3)).(color.NRGBA)
	g.canvas.FillPolygon([]kavad.Point{{X: 2, Y: 2}, {X: 5, Y: 2}, {X: 5, Y: 5}, {X: 2, Y: 5}},
		kavad.Color{R: 255, G: 255, B: 255, A: 1})
	g.above = color.NRGBAModel.Convert(g.canvas.dst.At(3, 3)).(color.NRGBA)
	g.drawn = true
}

func (g *imageGame) Layout(int, int) (int, int) { return 8, 8 }
