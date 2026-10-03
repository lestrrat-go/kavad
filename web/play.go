//go:build js && wasm

package web

import (
	"context"
	"syscall/js"

	"github.com/lestrrat-go/kavad"
)

// maxStep caps how far the clock moves between two animation frames. The
// browser stops sending frames to a hidden tab; with the cap, the show
// resumes where it was instead of jumping ahead.
const maxStep = 100.0 // ms

// hold is how long the last moment of a run stays up before the show
// restarts.
const hold = 1000 // ms

// Main, built for js/wasm, plays show on the canvas kavad.js hands it, and
// never returns.
func Main(show kavad.Show) {
	el := js.Global().Get("kavadCanvas") // set by kavad.js
	if el.IsUndefined() || el.IsNull() {
		fail("no canvas (load kavad.js, not show.wasm)")
		return
	}
	w, h := show.Size()
	c := newCanvas(el, w, h)
	ctx := context.Background()
	loop, err := kavad.NewLoop(ctx, show, hold)
	if err != nil {
		fail(err.Error())
		return
	}
	clock, last := 0.0, -1.0
	var frame js.Func
	frame = js.FuncOf(func(_ js.Value, args []js.Value) any {
		ts := args[0].Float()
		if last >= 0 {
			clock += min(max(ts-last, 0), maxStep)
		}
		last = ts
		if err := loop.Update(ctx, int(clock)); err != nil {
			fail(err.Error())
			return nil
		}
		c.begin()
		loop.Draw(c)
		js.Global().Call("requestAnimationFrame", frame)
		return nil
	})
	js.Global().Call("requestAnimationFrame", frame)
	select {} // requestAnimationFrame drives everything from here
}

func fail(msg string) {
	js.Global().Get("console").Call("error", "kavad: "+msg)
}
