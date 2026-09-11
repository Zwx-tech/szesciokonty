import { LitElement, html } from "lit";
import { classMap } from "lit/directives/class-map.js";
import type { Army, Seat } from "../protocol";

const ARMIES: Army[] = ["red", "blue", "green", "yellow"];

export type LobbyFormState = {
  code: string;
  seats: Seat[];
  youArmy?: Army;
  youReady: boolean;
  youHost: boolean;
  canStart: boolean;
  takenArmies: Set<Army | undefined>;
  status: string;
  error: string;
};

/** Light-DOM lobby panel with seats + army pick. */
export class LobbyForm extends LitElement {
  static properties = {
    state: { attribute: false },
  };

  state: LobbyFormState = {
    code: "—",
    seats: [],
    youReady: false,
    youHost: false,
    canStart: false,
    takenArmies: new Set(),
    status: "",
    error: "",
  };

  onSetArmy: (army: Army) => void = () => {};
  onToggleReady: () => void = () => {};
  onStart: () => void = () => {};
  onLeave: () => void = () => {};

  constructor() {
    super();
    this.classList.add("panel", "stack");
  }

  createRenderRoot(): HTMLElement | DocumentFragment {
    return this;
  }

  render(): unknown {
    const s = this.state;
    const seatsText =
      s.seats
        .map((seat) => {
          const bits = [
            seat.name,
            seat.host ? "host" : "guest",
            seat.army ?? "no army",
            seat.ready ? "ready" : "not ready",
            seat.connected ? "" : "offline",
          ].filter(Boolean);
          return bits.join(" · ");
        })
        .join("\n") || "Waiting…";

    return html`
      <h1>Lobby</h1>
      <h2>Code: ${s.code}</h2>
      <div class="seats">${seatsText}</div>
      <div class="muted">Army</div>
      <div class="row">
        ${ARMIES.map(
          (army) =>
            html`<button
              type="button"
              class=${classMap({
                [`army-${army}`]: true,
                selected: s.youArmy === army,
              })}
              ?disabled=${s.takenArmies.has(army)}
              @click=${() => this.onSetArmy(army)}
            >
              ${army}
            </button>`,
        )}
      </div>
      <div class="row">
        <button type="button" @click=${() => this.onToggleReady()}>
          ${s.youReady ? "Unready" : "Ready"}
        </button>
        <button
          type="button"
          ?disabled=${!(s.youHost && s.canStart)}
          @click=${() => this.onStart()}
        >
          Start
        </button>
      </div>
      <button type="button" @click=${() => this.onLeave()}>Leave</button>
      <div class="muted">${s.status}</div>
      <div class="error">${s.error}</div>
    `;
  }
}

if (!customElements.get("sk-lobby-form")) {
  customElements.define("sk-lobby-form", LobbyForm);
}

declare global {
  interface HTMLElementTagNameMap {
    "sk-lobby-form": LobbyForm;
  }
}
