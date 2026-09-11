import { LitElement, html } from "lit";
import { repeat } from "lit/directives/repeat.js";
import { classMap } from "lit/directives/class-map.js";
import type { HandTile } from "../protocol";
import { tileSvgUrl } from "../hex/tileAssets";
import { shortDefName } from "../tiles/format";
import { kindChip } from "../rules/matchUi";

/**
 * Light-DOM hand strip so global style.css applies.
 * Cards are keyed by tile.id via lit `repeat`.
 */
export class HandPanel extends LitElement {
  static properties = {
    hand: { attribute: false },
    interactive: { type: Boolean },
    selectedId: { type: String },
  };

  hand: HandTile[] = [];
  interactive = false;
  selectedId: string | null = null;
  onSelect: (tileId: string | null, tile?: HandTile) => void = () => {};

  constructor() {
    super();
    this.classList.add("hand");
  }

  createRenderRoot(): HTMLElement | DocumentFragment {
    return this;
  }

  get element(): HTMLElement {
    return this;
  }

  get root(): HTMLElement {
    return this;
  }

  /** Sync hand contents from the table shell. */
  setHand(
    hand: HandTile[],
    opts: { interactive: boolean; selectedId: string | null },
  ): void {
    this.hand = hand;
    this.interactive = opts.interactive;
    this.selectedId = opts.selectedId;
  }

  render(): unknown {
    if (this.hand.length === 0) {
      return html`<span class="hand-empty">Empty hand</span>`;
    }
    return html`${repeat(
      this.hand,
      (t) => t.id,
      (tile) => {
        const label = shortDefName(tile.defId);
        const selected = this.interactive && this.selectedId === tile.id;
        return html`<button
          type="button"
          class=${classMap({
            tile: true,
            [`kind-${tile.kind}`]: true,
            selected,
          })}
          ?disabled=${!this.interactive}
          @click=${() => this.onCardClick(tile)}
        >
          <span class="kind-chip">${kindChip(tile.kind)}</span>
          <img src=${tileSvgUrl(tile.defId)} alt=${label} draggable="false" />
          <span class="tile-label">${label}</span>
        </button>`;
      },
    )}`;
  }

  private onCardClick(tile: HandTile): void {
    if (!this.interactive) return;
    if (this.selectedId === tile.id) this.onSelect(null);
    else this.onSelect(tile.id, tile);
  }
}

if (!customElements.get("sk-hand-panel")) {
  customElements.define("sk-hand-panel", HandPanel);
}

declare global {
  interface HTMLElementTagNameMap {
    "sk-hand-panel": HandPanel;
  }
}
