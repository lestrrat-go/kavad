// Package imageassets is a small transparent show using the lestrrat-3d
// site's tools/hero PNG. The same embedded bytes reach every runner.
package imageassets

import (
	"context"
	"embed"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/draw"
	"github.com/lestrrat-go/kavad/ease"
)

//go:embed hero.png
var heroFS embed.FS

// Show places the site's CAD render beneath a moving caption.
type Show struct{}

func (Show) Size() (int, int) { return 640, 480 }

func (Show) Start(context.Context) (*kavad.Run, error) {
	hero, err := kavad.LoadPNG(heroFS, "hero.png")
	if err != nil {
		return nil, err
	}
	return kavad.NewRun(still{}, kavad.NewStage(), picture{hero: hero}, hero), nil
}

type still struct{}

func (still) Send(context.Context, string) error { return nil }
func (still) Done() bool                         { return false }

type picture struct{ hero *kavad.Image }

func (p picture) Paint(c kavad.Canvas, now int) {
	progress := ease.InOut(ease.Progress(now, 0, 1000))
	pen := draw.New(c)
	pen.Image(p.hero, 60+40*progress, 35, 520, 380, 1-0.5*progress)
	pen.Fade(0.7).Rect(80, 350, 480, 70, kavad.MustHex("#17232c"))
	pen.Text(kavad.Text{
		X: 110 + 130*progress, Y: 393, S: "CAD imagery in motion",
		Font: kavad.SansBold, Size: 24, Color: kavad.MustHex("#ffffff"),
	})
}
