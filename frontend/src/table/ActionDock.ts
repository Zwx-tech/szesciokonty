import { LitElement, html } from "lit";
import { classMap } from "lit/directives/class-map.js";
import { facingLabel } from "../tiles/format";

export type DockDisabledStates = {
  discard: boolean;
  play: boolean;
  end: boolean;
  unlucky: boolean;
  rotLeft: boolean;
  rotRight: boolean;
  skip?: boolean;
  recon?: boolean;
  quartermaster?: boolean;
  skipPassenger?: boolean;
};

export class ActionDock extends LitElement {
  static properties = {
    facingText: { type: String },
    unluckyTitle: { type: String },
    skipVisible: { type: Boolean },
    skipPassengerVisible: { type: Boolean },
    discardDisabled: { type: Boolean },
    playDisabled: { type: Boolean },
    endDisabled: { type: Boolean },
    unluckyDisabled: { type: Boolean },
    rotLeftDisabled: { type: Boolean },
    rotRightDisabled: { type: Boolean },
    skipDisabled: { type: Boolean },
    reconDisabled: { type: Boolean },
    quartermasterDisabled: { type: Boolean },
    skipPassengerDisabled: { type: Boolean },
  };

  facingText = facingLabel(0);
  unluckyTitle = "Redraw this turn’s draw once";
  skipVisible = false;
  skipPassengerVisible = false;
  discardDisabled = true;
  playDisabled = true;
  endDisabled = true;
  unluckyDisabled = true;
  rotLeftDisabled = true;
  rotRightDisabled = true;
  skipDisabled = true;
  reconDisabled = true;
  quartermasterDisabled = true;
  skipPassengerDisabled = true;

  onRotate: (delta: number) => void = () => {};
  onDiscard: () => void = () => {};
  onPlay: () => void = () => {};
  onUnlucky: () => void = () => {};
  onEndTurn: () => void = () => {};
  onSkip: () => void = () => {};
  onRecon: () => void = () => {};
  onQuartermaster: () => void = () => {};
  onSkipPassenger: () => void = () => {};

  constructor() {
    super();
    this.classList.add("action-dock");
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

  setFacingLabel(text: string): void {
    this.facingText = text;
  }

  setUnluckyTitle(title: string): void {
    this.unluckyTitle = title;
  }

  setDisabledStates(states: DockDisabledStates): void {
    this.discardDisabled = states.discard;
    this.playDisabled = states.play;
    this.endDisabled = states.end;
    this.unluckyDisabled = states.unlucky;
    this.rotLeftDisabled = states.rotLeft;
    this.rotRightDisabled = states.rotRight;
    if (states.skip != null) this.skipDisabled = states.skip;
    this.reconDisabled = states.recon ?? true;
    this.quartermasterDisabled = states.quartermaster ?? true;
    this.skipPassengerDisabled = states.skipPassenger ?? true;
  }

  setSkipVisible(visible: boolean): void {
    this.skipVisible = visible;
  }

  setSkipPassengerVisible(visible: boolean): void {
    this.skipPassengerVisible = visible;
  }

  setLocked(locked: boolean): void {
    this.discardDisabled = locked;
    this.playDisabled = locked;
    this.endDisabled = locked;
    this.unluckyDisabled = locked;
    this.rotLeftDisabled = locked;
    this.rotRightDisabled = locked;
    this.skipDisabled = !locked;
    this.reconDisabled = locked;
    this.quartermasterDisabled = locked;
    this.skipPassengerDisabled = locked;
  }

  render(): unknown {
    return html`
      <div class="dock-group facing-group">
        <span class="dock-label">Facing</span>
        <button
          type="button"
          ?disabled=${this.rotLeftDisabled}
          @click=${() => this.onRotate(-1)}
        >
          ◀
        </button>
        <span class="facing">${this.facingText}</span>
        <button
          type="button"
          ?disabled=${this.rotRightDisabled}
          @click=${() => this.onRotate(1)}
        >
          ▶
        </button>
      </div>
      <div class="dock-group turn-group">
        <span class="dock-label">Turn</span>
        <button
          type="button"
          ?disabled=${this.discardDisabled}
          @click=${() => this.onDiscard()}
        >
          Discard
        </button>
        <button
          type="button"
          ?disabled=${this.playDisabled}
          @click=${() => this.onPlay()}
        >
          Play
        </button>
        <button
          type="button"
          title=${this.unluckyTitle}
          ?disabled=${this.unluckyDisabled}
          @click=${() => this.onUnlucky()}
        >
          Unlucky
        </button>
        <button
          type="button"
          ?disabled=${this.endDisabled}
          @click=${() => this.onEndTurn()}
        >
          End turn
        </button>
      </div>
      <div class="dock-group ability-group">
        <span class="dock-label">Abilities</span>
        <button
          type="button"
          title="Peek 3 tiles from the opponent deck"
          ?disabled=${this.reconDisabled}
          @click=${() => this.onRecon()}
        >
          Recon
        </button>
        <button
          type="button"
          title="Recycle one discard tile to the bottom of your deck"
          ?disabled=${this.quartermasterDisabled}
          @click=${() => this.onQuartermaster()}
        >
          Recycle
        </button>
        <button
          type="button"
          class=${classMap({
            hidden: !this.skipPassengerVisible,
          })}
          ?disabled=${this.skipPassengerDisabled}
          @click=${() => this.onSkipPassenger()}
        >
          Skip passenger
        </button>
      </div>
      <button
        type="button"
        class=${classMap({ "skip-battle": true, hidden: !this.skipVisible })}
        ?disabled=${this.skipDisabled}
        @click=${() => this.onSkip()}
      >
        Skip battle
      </button>
    `;
  }
}

if (!customElements.get("sk-action-dock")) {
  customElements.define("sk-action-dock", ActionDock);
}

declare global {
  interface HTMLElementTagNameMap {
    "sk-action-dock": ActionDock;
  }
}
