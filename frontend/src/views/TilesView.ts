import { el, clear } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import type { Army } from "../protocol";

type ManifestEntry = { id: string; kind: string; file: string };
type Manifest = { armies: Record<string, ManifestEntry[]> };

const ARMIES: Army[] = ["red", "blue", "green", "yellow"];

export class TilesView extends View {
  private tabs: HTMLElement;
  private body: HTMLElement;
  private grid: HTMLElement;
  private panel: HTMLElement;
  private panelTitle: HTMLElement;
  private panelJson: HTMLElement;
  private status: HTMLElement;
  private army: Army = "red";
  private selectedId: string | null = null;
  private manifest: Manifest | null = null;
  private fetchSeq = 0;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("tiles-view");

    const header = el("div", { className: "tiles-header" });
    header.append(el("h1", { text: "Tile gallery" }));
    const back = el("button", { text: "Back" });
    back.addEventListener("click", () => this.router.show("hub"));
    header.append(back);
    this.root.append(header);

    this.status = el("div", { className: "muted" });
    this.root.append(this.status);

    this.tabs = el("div", { className: "row tiles-tabs" });
    for (const army of ARMIES) {
      const btn = el("button", {
        className: `army-${army}`,
        text: army,
        attrs: { "data-army": army },
      });
      btn.addEventListener("click", () => {
        this.army = army;
        this.selectedId = null;
        this.render();
        this.clearPanel();
      });
      this.tabs.append(btn);
    }
    this.root.append(this.tabs);

    this.body = el("div", { className: "tiles-body" });
    this.grid = el("div", { className: "tiles-grid" });
    this.panel = el("aside", { className: "tiles-panel hidden" });
    this.panelTitle = el("h2", { text: "Tile" });
    const close = el("button", {
      className: "tiles-panel-close",
      text: "Close",
    });
    close.addEventListener("click", () => {
      this.selectedId = null;
      this.clearPanel();
      this.render();
    });
    const panelHead = el("div", { className: "tiles-panel-head" });
    panelHead.append(this.panelTitle, close);
    this.panelJson = el("pre", { className: "tiles-panel-json" });
    this.panel.append(panelHead, this.panelJson);
    this.body.append(this.grid, this.panel);
    this.root.append(this.body);
  }

  onEnter(): void {
    this.selectedId = null;
    this.clearPanel();
    void this.load();
  }

  private async load(): Promise<void> {
    this.status.textContent = "Loading tiles…";
    try {
      const res = await fetch("/tiles/manifest.json");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      this.manifest = (await res.json()) as Manifest;
      this.status.textContent =
        "Click a tile for live JSON from the backend. SVGs: npm run generate:tiles";
      this.render();
    } catch (err) {
      this.manifest = null;
      this.status.textContent = `Failed to load /tiles/manifest.json — run npm run generate:tiles (${String(err)})`;
      clear(this.grid);
    }
  }

  private render(): void {
    if (!this.manifest) return;
    for (const btn of this.tabs.querySelectorAll("button")) {
      const a = btn.getAttribute("data-army");
      btn.classList.toggle("selected", a === this.army);
    }

    clear(this.grid);
    const list = this.manifest.armies[this.army] ?? [];
    for (const entry of list) {
      const card = el("div", {
        className:
          "tile-card" + (this.selectedId === entry.id ? " selected" : ""),
      });
      card.addEventListener("click", () => void this.selectTile(entry.id));
      const img = el("img", {
        attrs: {
          src: `/tiles/${entry.file}`,
          alt: entry.id,
          width: "104",
          height: "104",
        },
      });
      card.append(img);
      card.append(el("div", { className: "tile-card-id", text: entry.id }));
      card.append(el("div", { className: "muted", text: entry.kind }));
      this.grid.append(card);
    }
  }

  private async selectTile(id: string): Promise<void> {
    this.selectedId = id;
    this.render();
    this.panel.classList.remove("hidden");
    this.panelTitle.textContent = id;
    this.panelJson.textContent = "Loading…";

    const seq = ++this.fetchSeq;
    try {
      const res = await fetch(`/api/tiles/${encodeURIComponent(id)}`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data: unknown = await res.json();
      if (seq !== this.fetchSeq) return;
      this.panelJson.textContent = JSON.stringify(data, null, 2);
    } catch (err) {
      if (seq !== this.fetchSeq) return;
      this.panelJson.textContent = `Failed to load tile from backend:\n${String(err)}\n\nIs the Go server running?`;
    }
  }

  private clearPanel(): void {
    this.panel.classList.add("hidden");
    this.panelTitle.textContent = "Tile";
    this.panelJson.textContent = "";
  }
}
