import { LitElement, html } from "lit";

/** Light-DOM hub room create/join panel. */
export class HubForm extends LitElement {
  static properties = {
    code: { type: String },
    error: { type: String },
    busy: { type: Boolean },
  };

  code = "";
  error = "";
  busy = false;
  onCreate: () => void = () => {};
  onJoin: (code: string) => void = () => {};
  onGallery: () => void = () => {};

  constructor() {
    super();
    this.classList.add("panel", "stack");
  }

  createRenderRoot(): HTMLElement | DocumentFragment {
    return this;
  }

  render(): unknown {
    return html`
      <h1>Room</h1>
      <button
        type="button"
        ?disabled=${this.busy}
        @click=${() => this.onCreate()}
      >
        Create room
      </button>
      <div class="muted">Or join with code</div>
      <label class="field">
        <input
          type="text"
          maxlength="8"
          placeholder="code"
          .value=${this.code}
          @input=${(e: Event) => {
            this.code = (e.target as HTMLInputElement).value;
          }}
          @keydown=${(e: KeyboardEvent) => {
            if (e.key === "Enter") this.join();
          }}
        />
      </label>
      <button type="button" ?disabled=${this.busy} @click=${() => this.join()}>
        Join
      </button>
      <div class="error">${this.error}</div>
      <button type="button" @click=${() => this.onGallery()}>
        Tile gallery
      </button>
    `;
  }

  private join(): void {
    if (this.busy) return;
    this.onJoin(this.code.trim());
  }
}

if (!customElements.get("sk-hub-form")) {
  customElements.define("sk-hub-form", HubForm);
}

declare global {
  interface HTMLElementTagNameMap {
    "sk-hub-form": HubForm;
  }
}
