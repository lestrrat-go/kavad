package kavad

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"math"
)

// Image is a decoded PNG and its original bytes. It is safe to share between
// runs. The source filesystem and path are chosen by the show, not a runner.
type Image struct {
	id     string
	data   []byte
	pixels image.Image
	w, h   int
}

// LoadPNG reads and decodes name from fsys before playback. fsys must not be
// nil. An embed.FS can put the same asset in native and WASM builds. Invalid
// or missing PNGs return an error to Show.Start.
func LoadPNG(fsys fs.FS, name string) (*Image, error) {
	if fsys == nil {
		return nil, fmt.Errorf("kavad: nil image filesystem")
	}
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("kavad: reading image %q: %w", name, err)
	}
	pixels, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("kavad: decoding image %q: %w", name, err)
	}
	sum := sha256.Sum256(data)
	bounds := pixels.Bounds()
	return &Image{
		id: hex.EncodeToString(sum[:]), data: data, pixels: pixels,
		w: bounds.Dx(), h: bounds.Dy(),
	}, nil
}

// ID identifies equal PNG bytes across runs and renderers.
func (i *Image) ID() string { return i.id }

// Size returns the decoded image dimensions in pixels.
func (i *Image) Size() (int, int) { return i.w, i.h }

// PNG returns a copy of the encoded PNG for bundle output.
func (i *Image) PNG() []byte { return bytes.Clone(i.data) }

// Pixels returns the decoded image. Callers must not mutate it.
func (i *Image) Pixels() image.Image { return i.pixels }

// Fit returns the largest rectangle within x, y, w, h with the image's aspect
// ratio, centered on the unused axis. Nonpositive sizes return an empty rect.
func (i *Image) Fit(x, y, w, h float64) (float64, float64, float64, float64) {
	if w <= 0 || h <= 0 || i.w <= 0 || i.h <= 0 {
		return x, y, 0, 0
	}
	s := math.Min(w/float64(i.w), h/float64(i.h))
	fw, fh := float64(i.w)*s, float64(i.h)*s
	return x + (w-fw)/2, y + (h-fh)/2, fw, fh
}
