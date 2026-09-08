import { Application } from "pixi.js";

async function main() {
  const app = new Application();
  await app.init({
    resizeTo: window,
    background: "#1a1a1a",
    antialias: true,
    resolution: window.devicePixelRatio,
    autoDensity: true,
  });

  document.getElementById("pixi-container")!.appendChild(app.canvas);
}

main().catch(console.error);
