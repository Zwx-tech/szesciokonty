import { GameApp } from "./app/GameApp";

async function main() {
  const host = document.getElementById("pixi-container");
  if (!host) throw new Error("#pixi-container missing");

  const app = new GameApp();
  await app.init(host);
}

main().catch(console.error);
