// Package desktop plays a show in a native window with Ebitengine.
// Ebitengine calls Update 60 times a second; that tick is the show's clock.
//
// A show's play command is one line:
//
//	func main() { desktop.Main(promo.Show{}) }
//
// Flags: -fullscreen starts fullscreen (F toggles it, Esc quits);
// -shots 2000,11000 saves PNGs of those moments (ms) to -out and exits.
// The show loops with a one-second hold at the end.
package desktop

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/lestrrat-go/kavad"
)

// hold is how long the last moment of a run stays up before the show
// restarts.
const hold = 1000 // ms

// Main plays show in a window until the user quits, then exits the program.
func Main(show kavad.Show) {
	fullscreen := flag.Bool("fullscreen", false, "start fullscreen")
	shots := flag.String("shots", "", "comma-separated times (ms) to save as PNG, then exit")
	out := flag.String("out", ".", "directory for -shots")
	title := flag.String("title", "kavad", "window title")
	flag.Parse()
	if err := play(show, *title, *fullscreen, *shots, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func play(show kavad.Show, title string, fullscreen bool, shots, out string) error {
	c, err := newCanvas()
	if err != nil {
		return err
	}
	loop, err := kavad.NewLoop(context.Background(), show, hold)
	if err != nil {
		return err
	}
	w, h := show.Size()
	g := &game{canvas: c, loop: loop, w: w, h: h, out: out}
	for f := range strings.SplitSeq(shots, ",") {
		if f = strings.TrimSpace(f); f == "" {
			continue
		}
		ms, err := strconv.Atoi(f)
		if err != nil {
			return fmt.Errorf("-shots: %w", err)
		}
		g.shots = append(g.shots, ms)
	}
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(fullscreen)
	if err := ebiten.RunGame(g); err != nil && !errors.Is(err, ebiten.Termination) {
		return err
	}
	return nil
}

type game struct {
	canvas *canvas
	loop   *kavad.Loop
	w, h   int
	tick   int // Update calls so far
	shots  []int
	out    string
	err    error
}

func (g *game) Update() error {
	if g.err != nil {
		return g.err
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape) && len(g.shots) == 0:
		return ebiten.Termination
	case inpututil.IsKeyJustPressed(ebiten.KeyF):
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	g.tick++
	return g.loop.Update(context.Background(), g.tick*1000/ebiten.TPS())
}

func (g *game) Draw(screen *ebiten.Image) {
	g.canvas.dst = screen
	g.loop.Draw(g.canvas)
	if len(g.shots) > 0 && g.loop.Now() >= g.shots[0] {
		g.err = g.save(screen, g.shots[0])
		g.shots = g.shots[1:]
		if len(g.shots) == 0 && g.err == nil {
			g.err = ebiten.Termination
		}
	}
}

func (g *game) save(screen *ebiten.Image, ms int) error {
	f, err := os.Create(filepath.Join(g.out, fmt.Sprintf("shot_%05d.png", ms)))
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, screen)
}

func (g *game) Layout(int, int) (int, int) { return g.w, g.h }
