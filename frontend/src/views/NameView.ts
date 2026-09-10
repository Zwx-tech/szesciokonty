import { el } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";

export class NameView extends View {
  private field: HTMLInputElement;
  private errorLabel: HTMLElement;
  private busy = false;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");

    const panel = el("div", { className: "panel stack" });
    panel.append(el("h1", { text: "Hex" }));

    const label = el("label", { className: "field" });
    label.append(el("span", { className: "muted", text: "Display name" }));
    this.field = el("input", {
      attrs: { type: "text", maxlength: "16", placeholder: "name" },
    });
    label.append(this.field);
    panel.append(label);

    this.errorLabel = el("div", { className: "error" });
    panel.append(this.errorLabel);

    const btn = el("button", { text: "Continue" });
    btn.addEventListener("click", () => void this.continue());
    this.field.addEventListener("keydown", (e) => {
      if (e.key === "Enter") void this.continue();
    });
    panel.append(btn);

    this.root.append(panel);
  }

  onEnter(): void {
    if (this.session.name) this.field.value = this.session.name;
    this.field.focus();
  }

  private async continue(): Promise<void> {
    if (this.busy) return;
    const name = this.field.value.trim();
    if (!name) {
      this.errorLabel.textContent = "Enter a name";
      return;
    }
    this.busy = true;
    this.errorLabel.textContent = "";
    this.session.setName(name);
    try {
      await this.session.ensureConnected();
      const reconnecting = await this.session.tryReconnect();
      if (!reconnecting) this.router.show("hub");
    } catch {
      this.errorLabel.textContent = "Could not connect";
    } finally {
      this.busy = false;
    }
  }
}
