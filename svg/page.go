package svg

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lestrrat-go/kavad"
)

// WritePage writes frames (w×h SVG documents) into a self-playing HTML page
// at path. It writes each declared image once to an assets directory beside
// the page. Keep that directory with the page when moving it. Opened normally,
// the page loops the frames at fps.
// Opened with ?capture=1, it shows frame 0 and waits: window.showFrame(i)
// shows frame i, and window.fps, window.frameCount and window.frameSize
// describe the frames. capture/capture.cjs uses that to record video.
func WritePage(path string, frames []string, fps, w, h int, images ...*kavad.Image) error {
	data, err := json.Marshal(frames)
	if err != nil {
		return err
	}
	sources := make([]string, 0, len(images))
	if len(images) > 0 {
		assetsDir := filepath.Join(filepath.Dir(path), "assets")
		if err := os.MkdirAll(assetsDir, 0o755); err != nil {
			return err
		}
		seen := make(map[string]struct{}, len(images))
		for _, img := range images {
			if img == nil {
				return fmt.Errorf("svg: nil image")
			}
			if _, ok := seen[img.ID()]; ok {
				continue
			}
			seen[img.ID()] = struct{}{}
			name := img.ID() + ".png"
			if err := os.WriteFile(filepath.Join(assetsDir, name), img.PNG(), 0o644); err != nil {
				return err
			}
			sources = append(sources, "assets/"+name)
		}
	}
	assetData, err := json.Marshal(sources)
	if err != nil {
		return err
	}
	return os.WriteFile(path, fmt.Appendf(nil, page, w, h, w, h, fps, w, h, data, assetData), 0o644)
}

// Main is the body of a show's render command:
//
//	func main() { svg.Main(promo.Show{}) }
//
// It renders the show with Frames and writes <out>/frames.html with
// WritePage. Flags: -out (default "out"), -fps (default 30), -seconds
// (default 30).
func Main(show kavad.Show) {
	out := flag.String("out", "out", "output directory")
	fps := flag.Int("fps", 30, "frames per second")
	seconds := flag.Int("seconds", 30, "length to render")
	flag.Parse()
	if err := render(show, *out, *fps, *seconds); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func render(show kavad.Show, out string, fps, seconds int) error {
	frames, images, err := FramesWithImages(context.Background(), show, fps, seconds)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	path := filepath.Join(out, "frames.html")
	w, h := show.Size()
	if err := WritePage(path, frames, fps, w, h, images...); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "wrote %s (%d frames at %d fps)\n", path, len(frames), fps)
	return nil
}

const page = `<!doctype html>
<html><head><meta charset="utf-8"><title>kavad frames</title>
<style>html,body{margin:0;background:#000;height:100%%;display:flex;align-items:center;justify-content:center}
#stage{width:min(100vw,calc(100vh*%d/%d));aspect-ratio:%d/%d}#stage svg{width:100%%;height:100%%;display:block}</style>
</head><body><div id="stage"></div>
<script>
const fps = %d;
window.frameSize = [%d, %d];
const frames = %s;
const imageSources = %s;
const stage = document.getElementById("stage");
window.fps = fps;
window.frameCount = frames.length;
window.assetsReady = Promise.all(imageSources.map((src) => {
  const image = new Image();
  image.src = src;
  return image.decode();
}));
window.assetsReady.catch((err) => {
  stage.textContent = "kavad: " + err;
  stage.style.color = "#f88";
});
window.showFrame = async (i) => {
  await window.assetsReady;
  stage.innerHTML = frames[i];
  await Promise.all(Array.from(stage.querySelectorAll("image"), (image) => new Promise((resolve, reject) => {
    image.addEventListener("load", resolve, {once: true});
    image.addEventListener("error", () => reject(new Error("cannot load SVG image")), {once: true});
  })));
};
if (!new URLSearchParams(location.search).has("capture")) {
  const t0 = performance.now();
  const tick = (now) => {
    const i = Math.floor((now - t0) / 1000 * fps) %% frames.length;
    if (stage.dataset.i != i) { stage.dataset.i = i; window.showFrame(i); }
    requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
} else {
  window.showFrame(0);
}
</script></body></html>
`
