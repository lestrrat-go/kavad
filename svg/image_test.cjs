// Run after rendering examples/imageassets with -fps 2 -seconds 1.
// Needs Playwright with Chromium and ffmpeg, like capture/capture.cjs.
const assert = require("node:assert/strict");
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const { pathToFileURL } = require("node:url");
const { chromium } = require("playwright");

const pagePath = path.resolve(process.argv[2] || "out/image-frames", "frames.html");
const width = 640;
const height = 480;

function sum(pixels, x, y) {
  const i = (y * width + x) * 4;
  return pixels[i] + pixels[i + 1] + pixels[i + 2];
}

function leftEdge(pixels, y) {
  for (let x = 0; x < width; x++) if (sum(pixels, x, y) > 50) return x;
  return width;
}

function textPixels(pixels) {
  let count = 0;
  for (let y = 365; y <= 400; y++) for (let x = 80; x < 560; x++) {
    const i = (y * width + x) * 4;
    if (pixels[i] > 230 && pixels[i + 1] > 230 && pixels[i + 2] > 230) count++;
  }
  return count;
}

(async () => {
  const browser = await chromium.launch();
  try {
    const page = await browser.newPage({ viewport: { width, height } });
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    await page.goto(pathToFileURL(pagePath).href + "?capture=1");
    const info = await page.evaluate(() => ({ size: window.frameSize, count: window.frameCount }));
    assert.deepEqual(info, { size: [width, height], count: 2 });
    const frame = async (i) => {
      await page.evaluate((n) => window.showFrame(n), i);
      const screenshot = await page.screenshot();
      return execFileSync("ffmpeg", ["-loglevel", "error", "-f", "image2pipe", "-c:v", "png",
        "-i", "pipe:0", "-frames:v", "1", "-f", "rawvideo", "-pix_fmt", "rgba", "pipe:1"],
      { input: screenshot, maxBuffer: width * height * 4 + 4096 });
    };
    const first = await frame(0);
    const second = await frame(1);
    assert.ok(leftEdge(second, 200) > leftEdge(first, 200) + 10, "the image moved right");
    const before = sum(first, 300, 250);
    const after = sum(second, 320, 250);
    assert.ok(after < before * 0.9 && after > before * 0.5, "the image faded");
    assert.ok(textPixels(first) > 50 && textPixels(second) > 50, "the caption appears above the image");
    assert.deepEqual(errors, []);
  } finally {
    await browser.close();
  }
  console.log("SVG image placement, opacity, and draw order passed");
})().catch((error) => { console.error(error); process.exitCode = 1; });
