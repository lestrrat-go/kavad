// capture.cjs turns <out-dir>/frames.html (written by a show's svg.Main
// render command) into <out-dir>/video.mp4 and <out-dir>/video.gif. It opens
// the page in headless Chromium (Playwright), screenshots every frame at the
// show's size, and encodes them with ffmpeg at the page's frame rate.
//
// Usage: node capture.cjs <out-dir>
// Needs the playwright package (with Chromium) and ffmpeg on PATH.
const { chromium } = require("playwright");
const { execFileSync } = require("child_process");
const fs = require("fs");
const path = require("path");

(async () => {
  const out = path.resolve(process.argv[2] || "out");
  const framesDir = path.join(out, "frames");
  fs.rmSync(framesDir, { recursive: true, force: true });
  fs.mkdirSync(framesDir, { recursive: true });

  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto("file://" + path.join(out, "frames.html") + "?capture=1");
  await page.evaluate(() => window.assetsReady);
  const { fps, n, size } = await page.evaluate(() => ({
    fps: window.fps, n: window.frameCount, size: window.frameSize,
  }));
  await page.setViewportSize({ width: size[0], height: size[1] });
  for (let i = 0; i < n; i++) {
    await page.evaluate((i) => window.showFrame(i), i);
    await page.screenshot({ path: path.join(framesDir, String(i).padStart(5, "0") + ".png") });
  }
  await browser.close();

  const mp4 = path.join(out, "video.mp4");
  execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-framerate", String(fps),
    "-i", path.join(framesDir, "%05d.png"),
    "-c:v", "libx264", "-pix_fmt", "yuv420p", "-crf", "18", mp4]);
  fs.rmSync(framesDir, { recursive: true, force: true });
  console.log(`wrote ${mp4} (${n} frames at ${fps} fps)`);

  // video.gif is for places that show images but not video or scripts, such
  // as a GitHub README: 800px wide, 20 fps, one shared 64-colour palette.
  const gif = path.join(out, "video.gif");
  execFileSync("ffmpeg", ["-y", "-loglevel", "error", "-i", mp4,
    "-vf", "fps=20,scale=800:-1:flags=lanczos,split[a][b];[a]palettegen=max_colors=64[p];[b][p]paletteuse=dither=none",
    "-loop", "0", gif]);
  console.log(`wrote ${gif}`);
})();
