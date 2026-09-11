import { App } from "./app/App";
import { applyTheme } from "./ui/theme";

function main() {
  applyTheme();
  const host = document.getElementById("view-root");
  if (!host) throw new Error("#view-root missing");

  const app = new App(host);
  app.start();
}

main();
