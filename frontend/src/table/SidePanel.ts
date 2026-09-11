import { LitElement, html } from "lit";
import type { HandTile, MatchState } from "../protocol";
import { tileSvgUrl } from "../hex/tileAssets";
import { fetchTileDef } from "../tiles/catalog";
import { describeDef, shortDefName } from "../tiles/format";
import { boardRuntimeLines } from "../rules/matchUi";

export type InspectTarget =
  | null
  | { source: "board"; tileId: string }
  | { source: "hand"; defId: string };

export class SidePanel extends LitElement {
  static properties = {
    collapsed: { type: Boolean, reflect: true },
    inspectTitle: { type: String },
    inspectLines: { attribute: false },
    inspectArt: { type: String },
    inspectPlaceholder: { type: String },
    logLines: { attribute: false },
    reconPeek: { attribute: false },
    discardTiles: { attribute: false },
    discardPickMode: { type: Boolean },
  };

  collapsed = false;
  inspectTitle = "Inspect";
  inspectLines: string[] = [];
  inspectArt = "";
  inspectPlaceholder = "Hover or select a tile";
  logLines: string[] = [];
  reconPeek: string[] = [];
  discardTiles: HandTile[] = [];
  discardPickMode = false;

  private inspect: InspectTarget = null;
  onCollapseChange: (collapsed: boolean) => void = () => {};
  onDiscardPick: (tileId: string) => void = () => {};

  constructor() {
    super();
    this.classList.add("table-side");
  }

  createRenderRoot(): HTMLElement | DocumentFragment {
    return this;
  }

  get element(): HTMLElement {
    return this;
  }

  setCollapsed(collapsed: boolean): void {
    if (this.collapsed === collapsed) return;
    this.collapsed = collapsed;
    this.onCollapseChange(collapsed);
  }

  toggle(): void {
    this.setCollapsed(!this.collapsed);
  }

  setInspect(target: InspectTarget): void {
    this.inspect = target;
  }

  getInspect(): InspectTarget {
    return this.inspect;
  }

  clearInspect(): void {
    this.inspect = null;
    this.inspectTitle = "Inspect";
    this.inspectArt = "";
    this.inspectLines = [];
    this.inspectPlaceholder = "Hover or select a tile";
  }

  renderEventLog(lines: string[]): void {
    this.logLines = lines;
  }

  setReconPeek(defIds: string[]): void {
    this.reconPeek = defIds;
  }

  setDiscard(tiles: HandTile[], pickMode: boolean): void {
    this.discardTiles = tiles;
    this.discardPickMode = pickMode;
  }

  async refreshInspect(match: MatchState | null | undefined): Promise<void> {
    const target = this.inspect;
    if (!target) {
      this.inspectTitle = "Inspect";
      this.inspectArt = "";
      this.inspectLines = [];
      this.inspectPlaceholder = "Hover or select a tile";
      return;
    }

    let defId = "";
    let runtimeLines: string[] = [];
    if (target.source === "hand") {
      defId = target.defId;
      this.inspectTitle = shortDefName(defId);
    } else if (match) {
      const tile = match.board.find((t) => t.id === target.tileId);
      if (!tile) {
        this.inspectArt = "";
        this.inspectLines = [];
        this.inspectPlaceholder = "Tile gone";
        return;
      }
      defId = tile.defId;
      this.inspectTitle = shortDefName(defId);
      runtimeLines = boardRuntimeLines(tile, match);
    }

    const def = await fetchTileDef(defId);
    if (this.inspect !== target) return;
    this.inspectPlaceholder = "";
    this.inspectArt = tileSvgUrl(defId);
    const lines = def ? describeDef(def) : [`id: ${defId}`];
    this.inspectLines = [...runtimeLines, ...lines];
  }

  render(): unknown {
    const toggleLabel = this.collapsed ? "⟨" : "⟩";
    const toggleTitle = this.collapsed ? "Expand panel" : "Collapse panel";
    const log =
      this.logLines.length === 0
        ? html`<div class="muted">No events yet</div>`
        : this.logLines
            .slice()
            .reverse()
            .map((line) => html`<div class="log-line">${line}</div>`);

    return html`
      <div class="side-head">
        <h2>${this.inspectTitle}</h2>
        <button
          type="button"
          class="side-toggle"
          title=${toggleTitle}
          @click=${() => this.toggle()}
        >
          ${toggleLabel}
        </button>
      </div>
      <div class="side-body">
        <div class="inspect-body muted">
          ${
            this.inspectArt
              ? html`<img
                  class="inspect-art"
                  src=${this.inspectArt}
                  alt=${this.inspectTitle}
                />`
              : this.inspectPlaceholder || null
          }
          ${this.inspectLines.map(
            (line) => html`<div class="inspect-line">${line}</div>`,
          )}
        </div>
        ${
          this.reconPeek.length > 0
            ? html`
                <h3 class="log-head">Recon peek</h3>
                <div class="peek-row">
                  ${this.reconPeek.map(
                    (id) => html`
                      <button
                        type="button"
                        class="peek-tile"
                        title=${shortDefName(id)}
                        @click=${() => {
                          this.setInspect({ source: "hand", defId: id });
                          void this.refreshInspect(null);
                        }}
                      >
                        <img src=${tileSvgUrl(id)} alt=${shortDefName(id)} />
                        <span>${shortDefName(id)}</span>
                      </button>
                    `,
                  )}
                </div>
              `
            : null
        }
        ${
          this.discardTiles.length > 0
            ? html`
                <h3 class="log-head">
                  ${this.discardPickMode
                    ? "Pick discard to recycle"
                    : "Your discard"}
                </h3>
                <div class="peek-row">
                  ${this.discardTiles.map(
                    (t) => html`
                      <button
                        type="button"
                        class="peek-tile"
                        ?disabled=${!this.discardPickMode}
                        title=${shortDefName(t.defId)}
                        @click=${() => {
                          if (this.discardPickMode) this.onDiscardPick(t.id);
                          else {
                            this.setInspect({
                              source: "hand",
                              defId: t.defId,
                            });
                            void this.refreshInspect(null);
                          }
                        }}
                      >
                        <img
                          src=${tileSvgUrl(t.defId)}
                          alt=${shortDefName(t.defId)}
                        />
                        <span>${shortDefName(t.defId)}</span>
                      </button>
                    `,
                  )}
                </div>
              `
            : null
        }
        <h3 class="log-head">Events</h3>
        <div class="event-log">${log}</div>
      </div>
    `;
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has("collapsed")) {
      this.classList.toggle("collapsed", this.collapsed);
    }
  }
}

if (!customElements.get("sk-side-panel")) {
  customElements.define("sk-side-panel", SidePanel);
}

declare global {
  interface HTMLElementTagNameMap {
    "sk-side-panel": SidePanel;
  }
}
