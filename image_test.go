package kavad_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/lestrrat-go/kavad"
	"github.com/stretchr/testify/require"
)

func TestLoadPNG(t *testing.T) {
	pixels := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	pixels.Set(0, 0, color.NRGBA{R: 255, A: 255})
	pixels.Set(1, 0, color.NRGBA{B: 255, A: 128})
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, pixels))
	fsys := fstest.MapFS{"image.png": {Data: encoded.Bytes()}, "bad.png": {Data: []byte("bad")}}

	img, err := kavad.LoadPNG(fsys, "image.png")
	require.NoError(t, err)
	w, h := img.Size()
	require.Equal(t, 2, w)
	require.Equal(t, 1, h)
	x, y, fw, fh := img.Fit(10, 20, 100, 100)
	require.Equal(t, []float64{10, 45, 100, 50}, []float64{x, y, fw, fh})
	require.Equal(t, pixels.At(1, 0), img.Pixels().At(1, 0))
	copyOfPNG := img.PNG()
	copyOfPNG[0] = 0
	require.NotEqual(t, copyOfPNG[0], img.PNG()[0])

	_, err = kavad.LoadPNG(fsys, "missing.png")
	require.ErrorContains(t, err, "reading image")
	_, err = kavad.LoadPNG(fsys, "bad.png")
	require.ErrorContains(t, err, "decoding image")
}
