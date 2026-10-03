// kavad.js plays a kavad show (show.wasm) on a <canvas>. Embed it with
//
//   <script src="path/to/kavad.js"></script>
//
// It loads wasm_exec.js and show.wasm from its own directory, and draws into
// <canvas id="kavad">, adding one after the script tag when the page has
// none. The canvas fills its container's width at the show's aspect ratio.
(() => {
  const script = document.currentScript;
  const base = new URL(".", script.src);

  let canvas = document.getElementById("kavad");
  if (!canvas) {
    canvas = document.createElement("canvas");
    canvas.id = "kavad";
    script.after(canvas);
  }
  // :where() has no specificity, so any rule in the page overrides these.
  const style = document.createElement("style");
  style.textContent = ":where(#kavad){display:block;width:100%;height:auto}";
  document.head.prepend(style);

  const loadSupport = () => {
    if (typeof Go !== "undefined") return Promise.resolve();
    return new Promise((resolve, reject) => {
      const s = document.createElement("script");
      s.src = new URL("wasm_exec.js", base).href;
      s.onload = resolve;
      s.onerror = () => reject(new Error("cannot load " + s.src));
      document.head.append(s);
    });
  };

  loadSupport()
    .then(() => fetch(new URL("show.wasm", base)))
    .then((res) => {
      if (!res.ok) throw new Error(`${res.url}: ${res.status}`);
      return res.arrayBuffer();
    })
    .then((bytes) => {
      const go = new Go();
      return WebAssembly.instantiate(bytes, go.importObject).then(({ instance }) => {
        globalThis.kavadCanvas = canvas; // web.Main reads this
        go.run(instance);
      });
    })
    .catch((err) => console.error("kavad:", err));
})();
