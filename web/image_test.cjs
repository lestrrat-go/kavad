// Run against the imageassets WASM bundle with Playwright installed:
//   node web/image_test.cjs out/web
const assert = require("node:assert/strict");
const fs = require("node:fs");
const http = require("node:http");
const path = require("node:path");
const { chromium } = require("playwright");

const root = path.resolve(process.argv[2] || "out/web");
const overlay = `<!doctype html><html><head><style>
html,body{margin:0;background:rgb(20,120,60)}
#kavad{display:block;width:640px;height:480px}
</style></head><body><canvas id="kavad"></canvas><script src="kavad.js"></script></body></html>`;
const types = { ".js": "text/javascript", ".wasm": "application/wasm" };
const server = http.createServer((req, res) => {
  if (req.url === "/overlay.html") {
    res.setHeader("Content-Type", "text/html");
    res.end(overlay);
    return;
  }
  const name = path.basename(req.url);
  if (!["kavad.js", "wasm_exec.js", "show.wasm"].includes(name)) {
    res.writeHead(404).end();
    return;
  }
  res.setHeader("Content-Type", types[path.extname(name)]);
  fs.createReadStream(path.join(root, name)).pipe(res);
});

function alpha(data, x, y) { return data[(y * 640 + x) * 4 + 3]; }

function leftEdge(data, y) {
  for (let x = 0; x < 640; x++) if (alpha(data, x, y) > 5) return x;
  return 640;
}

(async () => {
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  let browser;
  try {
    browser = await chromium.launch();
    const page = await browser.newPage({ viewport: { width: 640, height: 480 } });
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    page.on("console", (message) => {
      if (message.type() === "error") errors.push(message.text());
    });
    await page.addInitScript(() => {
      const create = window.createImageBitmap;
      window.imageBitmapCalls = 0;
      window.createImageBitmap = (...args) => {
        window.imageBitmapCalls++;
        return create(...args);
      };
    });
    await page.goto(`http://127.0.0.1:${server.address().port}/overlay.html`);
    await page.waitForFunction(() => {
      const c = document.getElementById("kavad");
      return c.width === 640 && c.getContext("2d").getImageData(350, 150, 1, 1).data[3] > 0;
    });
    const pixels = () => page.evaluate(() => Array.from(document.getElementById("kavad")
      .getContext("2d").getImageData(0, 0, 640, 480).data));
    const first = await pixels();
    await page.waitForTimeout(1400);
    const second = await pixels();
    const bitmapCalls = await page.evaluate(() => window.imageBitmapCalls);
    let cleared = 0;
    let whiteText = 0;
    for (let y = 0; y < 480; y++) for (let x = 0; x < 640; x++) {
      const i = (y * 640 + x) * 4;
      if (first[i + 3] > 0 && second[i + 3] === 0) cleared++;
      if (y >= 365 && y <= 400 && second[i] > 230 && second[i + 1] > 230 &&
          second[i + 2] > 230 && second[i + 3] > 200) whiteText++;
    }
    assert.equal(alpha(second, 0, 0), 0, "page background remains visible");
    assert.ok(cleared > 100, "the previous transparent frame was cleared");
    assert.ok(leftEdge(second, 200) > leftEdge(first, 200) + 20, "the image moved right");
    assert.ok(alpha(second, 350, 150) < alpha(first, 350, 150), "image opacity decreased");
    assert.ok(whiteText > 50, "the caption is visible above the image");
    assert.equal(bitmapCalls, 1, "the image was decoded only once");
    assert.deepEqual(errors, []);
  } finally {
    await browser?.close();
    await new Promise((resolve) => server.close(resolve));
  }
  console.log("web image placement, alpha, order, reuse, and transparent clearing passed");
})().catch((error) => { console.error(error); process.exitCode = 1; });
