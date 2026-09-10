import { App } from "./app/App";

function main() {
  const host = document.getElementById("view-root");
  if (!host) throw new Error("#view-root missing");

  const app = new App(host);
  app.start();
}

main();
