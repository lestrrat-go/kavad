# kavad

kavad turns an [fsm](https://github.com/lestrrat-go/fsm) state machine into an
animated show: a promo video, a signage loop, a kiosk attract screen. The name
comes from the kaavad of Rajasthan, a wooden box whose painted panels a
storyteller opens one after another.

The machine is the storyboard. Each scene is a state, and the machine sets the
order and the timing by asking its host to schedule events
(`host.Schedule("next", 5000)`). Go code draws each moment on a small canvas
(polygons, polylines, circles and text). The machine never sees a clock: the
program that plays the show advances time and delivers the scheduled events,
so the same show plays the same way in every runner.

## Runners

| Package | Plays the show | Clock |
|---|---|---|
| `desktop` | in a native window (Ebitengine) | Ebitengine's 60 Hz update tick |
| `web` | in a web page, as WebAssembly on a `<canvas>` | the browser's `requestAnimationFrame` |
| `svg` | as one SVG document per frame | a virtual clock, for tests and video capture |

`desktop` and `web` loop the show, holding the last moment for one second
before restarting.

## Writing a show

A show has four parts. [`examples/promo`](examples/promo) has all of them.

1. **The storyboard machine.** Write it as an fsm machine file: a Go file
   whose `Definition` returns the machine, with typed Go functions as entry
   actions. A scene state's entry action calls `host.Scene(id)`, does the
   scene's own work, and calls `host.Schedule("next", ms)`; the state moves to
   the next scene on `"next"`. `fsm import -host` writes a first version of
   the file from a JSON document. Run `fsm check` on the file after each edit.
2. **The host interface.** Embed `kavad.Host` (which has `Scene` and
   `Schedule`) and add the calls your actions make.
3. **The host.** Embed `*kavad.Stage` to get `Scene` and `Schedule`, implement
   the rest, and implement `kavad.Painter`. `Paint(c, now)` draws the moment
   `now` (ms). The `draw` package has a pen that fades and shifts what it
   draws, plus outline helpers. The `ease` package has easing curves.
4. **A `kavad.Show` value.** `Size` returns the canvas size. `Start` compiles
   the machine against a fresh host and returns
   `kavad.NewRun(instance, host.Stage, host)`.

Each runner then needs a one-line command:

```go
func main() { desktop.Main(promo.Show{}) } // cmd/play
func main() { web.Main(promo.Show{}) }     // cmd/web
func main() { svg.Main(promo.Show{}) }     // cmd/render
```

## Running the promo

From the repository root:

```sh
go run github.com/lestrrat-go/fsm/cmd/fsm check examples/promo/promofsm/promo.go
go run ./examples/promo/cmd/play                # native window; F toggles fullscreen, Esc quits
go run ./examples/promo/cmd/web -out out/web    # files for a web page
go run ./examples/promo/cmd/render -out out     # out/frames.html
node capture/capture.cjs out                    # out/video.mp4 and out/video.gif
```

`capture.cjs` needs the `playwright` package (with Chromium) and `ffmpeg`.
The GIF is 800 px wide at 20 fps, for places such as a GitHub README that show
images but not video or scripts.

## Embedding in a web page

`go run ./examples/promo/cmd/web -out out/web` writes `show.wasm` (about 4 MB,
1.1 MB gzipped), `wasm_exec.js`, `kavad.js` and `index.html`. Serve that
directory, then add one tag where the show should appear:

```html
<script src="web/kavad.js"></script>
```

`kavad.js` loads the other files from its own directory and draws into
`<canvas id="kavad">`, adding the canvas after the script tag when the page
has none. The canvas fills the width of its container at the show's aspect
ratio. To use an `<iframe>` instead, point it at `index.html`. A page can hold
one show.

## Status

fsm has no tagged release yet. `go.mod` requires a pseudo-version of a commit
on fsm's `main` branch.
