// Package web plays a show in a web page: the show's machine and painter,
// compiled to WebAssembly, drawing on a <canvas>. The browser's
// requestAnimationFrame is the clock.
//
// A show's web command is one line:
//
//	func main() { web.Main(promo.Show{}) }
//
// Main does different things depending on the build:
//
//   - Built for js/wasm, it plays the show on the page's canvas, looping it
//     with a one-second hold at the end.
//   - Built natively (go run), it compiles the calling main package for
//     js/wasm and writes the files a page needs to -out (default "out/web").
//
// The output directory holds show.wasm, wasm_exec.js (from the Go
// installation that built show.wasm; the two must match), kavad.js and
// index.html. Serve the directory over HTTP, then embed the show with one
// script tag:
//
//	<script src="web/kavad.js"></script>
//
// kavad.js loads wasm_exec.js and show.wasm from its own directory and draws
// into <canvas id="kavad">, adding that canvas after the script tag when the
// page has none. The canvas fills the width of its container at the show's
// aspect ratio. index.html is a page holding only the show, for use in an
// <iframe>. A page can hold one show.
package web
