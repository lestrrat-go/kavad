// Package stage is the promo's host: Host is what the promo machine's
// actions can call, and Screen implements it and draws every scene.
package stage

import "github.com/lestrrat-go/kavad"

// Host is everything the promo machine can do to the outside world. The
// machine file (package promofsm) calls it as host.
type Host interface {
	// Scene and Schedule come from kavad; Screen gets them from the
	// embedded *kavad.Stage.
	kavad.Host
	// Caption sets the line of text under the picture ("" for none).
	Caption(text string)
	// Progress lights the first n of the progress dots.
	Progress(n int)
	// Token moves the event token of the "events" scene to a state.
	Token(state string)
}

// The picture is W×H; renderers scale it to their output.
const (
	W = 1280
	H = 720
)

// Screen implements Host and kavad.Painter.
type Screen struct {
	*kavad.Stage
	caption  string
	capAt    int
	progress int
	token    string
	prev     string
	tokenAt  int
}

var (
	_ Host          = (*Screen)(nil)
	_ kavad.Painter = (*Screen)(nil)
)

// New returns a screen at time 0.
func New() *Screen { return &Screen{Stage: kavad.NewStage()} }

// Caption implements Host.
func (s *Screen) Caption(text string) {
	if text != s.caption {
		s.caption, s.capAt = text, s.Now()
	}
}

// Progress implements Host.
func (s *Screen) Progress(n int) { s.progress = n }

// Token implements Host.
func (s *Screen) Token(state string) {
	s.prev, s.token, s.tokenAt = s.token, state, s.Now()
}

// Palette.
var (
	colBG     = kavad.MustHex("#0f1419")
	colPanel  = kavad.MustHex("#171e26")
	colFG     = kavad.MustHex("#f4f1ea")
	colDim    = kavad.MustHex("#3a4250")
	colLabel  = kavad.MustHex("#8a94a6")
	colAccent = kavad.MustHex("#ff6b35")
	colTeal   = kavad.MustHex("#2ec4b6")
	colYellow = kavad.MustHex("#ffd23f")
	colRed    = kavad.MustHex("#ef476f")
	colGreen  = kavad.MustHex("#06d6a0")
)
