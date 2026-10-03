// Package promo is a 30-second marketing animation for fsm, and the
// reference kavad show. The storyboard is a state machine (promofsm/promo.go):
// each scene is a state, and the machine asks the host to schedule the events
// that cut between them. The host (stage/) draws letters and geometric
// shapes.
//
// Commands (from the repository root):
//
//	go run github.com/lestrrat-go/fsm/cmd/fsm check examples/promo/promofsm/promo.go
//	go run ./examples/promo/cmd/play                  # native window
//	go run ./examples/promo/cmd/web -out out/web      # embeddable web page
//	go run ./examples/promo/cmd/render -out out       # out/frames.html
//	node capture/capture.cjs out                      # out/video.mp4, out/video.gif
package promo

import (
	"context"

	"github.com/lestrrat-go/kavad"
	"github.com/lestrrat-go/kavad/examples/promo/promofsm"
	"github.com/lestrrat-go/kavad/examples/promo/stage"
)

// Show is the promo as a kavad.Show.
type Show struct{}

var _ kavad.Show = Show{}

// Size implements kavad.Show.
func (Show) Size() (int, int) { return stage.W, stage.H }

// Start implements kavad.Show: a fresh screen and a fresh machine instance,
// which enters "intro" at t=0.
func (Show) Start(ctx context.Context) (*kavad.Run, error) {
	return start(ctx, stage.New())
}

// StartOn is Start with a caller-supplied screen, for tests that inspect
// it.
func StartOn(ctx context.Context, s *stage.Screen) (*kavad.Run, error) {
	return start(ctx, s)
}

func start(ctx context.Context, s *stage.Screen) (*kavad.Run, error) {
	prog, err := promofsm.Program(s)
	if err != nil {
		return nil, err
	}
	inst, err := prog.Start(ctx)
	if err != nil {
		return nil, err
	}
	return kavad.NewRun(inst, s.Stage, s), nil
}
