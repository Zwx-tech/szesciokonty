import { LitElement, html } from "lit";

/** Light-DOM name entry panel. */
export class NameForm extends LitElement {
  static properties = {
    nameValue: { type: String },
    error: { type: String },
    busy: { type: Boolean },
  };

  nameValue = "";
  error = "";
  busy = false;
  onContinue: (name: string) => void = () => {};

  constructor() {
    super();
    this.classList.add("panel", "stack");
  }

  createRenderRoot(): HTMLElement | DocumentFragment {
    return this;
  }

  focusField(): void {
    this.renderRoot.querySelector("input")?.focus();
  }

  render(): unknown {
    return html`
      <h1>Hex</h1>
      <label class="field">
        <span class="muted">Display name</span>
        <input
          type="text"
          maxlength="16"
          placeholder="name"
          .value=${this.nameValue}
          @input=${(e: Event) => {
            this.nameValue = (e.target as HTMLInputElement).value;
          }}
          @keydown=${(e: KeyboardEvent) => {
            if (e.key === "Enter") this.submit();
          }}
        />
      </label>
      <div class="error">${this.error}</div>
      <button
        type="button"
        ?disabled=${this.busy}
        @click=${() => this.submit()}
      >
        Continue
      </button>
    `;
  }

  private submit(): void {
    if (this.busy) return;
    this.onContinue(this.nameValue.trim());
  }
}

if (!customElements.get("sk-name-form")) {
  customElements.define("sk-name-form", NameForm);
}

declare global {
  interface HTMLElementTagNameMap {
    "sk-name-form": NameForm;
  }
}
