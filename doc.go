// Package kavad turns an fsm state machine into an animated show: a promo, a
// signage loop, a kiosk attract screen. The machine is the storyboard. Each
// scene is a state, and the machine sets the order and the timing by asking
// its host to schedule events. Go code draws each moment on a small Canvas.
//
// A show is three pieces:
//
//   - An fsm machine file: a Go file whose Definition returns the machine.
//     A scene state's entry action calls host.Scene, then asks the host to
//     schedule the event that moves on to the next scene.
//   - A host: a type that embeds *Stage, which implements the Scene and
//     Schedule calls kavad gives every show, plus whatever other calls the
//     machine's actions make. The host also implements Painter, drawing the
//     current moment.
//   - A Show value whose Start compiles the machine against a fresh host and
//     returns a Run.
//
// Runners own the clock. Each one advances a Run to the current time, which
// delivers the events the machine scheduled up to that moment, and then
// draws. The machine never sees a clock, so every runner plays the same show
// the same way:
//
//   - package desktop plays it in a native window (Ebitengine).
//   - package web compiles it to WebAssembly and draws on a <canvas>; a page
//     embeds it with one script tag.
//   - package svg renders it frame by frame on a virtual clock, for tests,
//     previews and video capture (capture/capture.cjs).
//
// examples/promo is a complete show: the 30-second promo for fsm.
package kavad
