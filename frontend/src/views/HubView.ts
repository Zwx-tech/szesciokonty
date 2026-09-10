import { el } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";

export class HubView extends View {
  private codeField: HTMLInputElement;
  private errorLabel: HTMLElement;
  private busy = false;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");

    const panel = el("div", { className: "panel stack" });
    panel.append(el("h1", { text: "Room" }));

    const createBtn = el("button", { text: "Create room" });
    createBtn.addEventListener("click", () => void this.create());
    panel.append(createBtn);

    panel.append(el("div", { className: "muted", text: "Or join with code" }));

    const label = el("label", { className: "field" });
    this.codeField = el("input", {
      attrs: { type: "text", maxlength: "8", placeholder: "code" },
    });
    label.append(this.codeField);
    panel.append(label);

    const joinBtn = el("button", { text: "Join" });
    joinBtn.addEventListener("click", () => void this.join());
    this.codeField.addEventListener("keydown", (e) => {
      if (e.key === "Enter") void this.join();
    });
    panel.append(joinBtn);

    this.errorLabel = el("div", { className: "error" });
    panel.append(this.errorLabel);

    const galleryBtn = el("button", { text: "Tile gallery" });
    galleryBtn.addEventListener("click", () => this.router.show("tiles"));
    panel.append(galleryBtn);

    this.root.append(panel);
  }

  onEnter(): void {
    this.syncError();
  }

  onSessionUpdate(): void {
    this.syncError();
  }

  private syncError(): void {
    this.errorLabel.textContent = this.session.error ?? "";
  }

  private async create(): Promise<void> {
    if (this.busy) return;
    this.busy = true;
    try {
      await this.session.ensureConnected();
      this.session.createRoom();
    } catch {
      this.session.setError("Could not connect");
    } finally {
      this.busy = false;
    }
  }

  private async join(): Promise<void> {
    if (this.busy) return;
    const code = this.codeField.value.trim();
    if (!code) {
      this.errorLabel.textContent = "Enter a room code";
      return;
    }
    this.busy = true;
    try {
      await this.session.ensureConnected();
      this.session.joinRoom(code);
    } catch {
      this.session.setError("Could not connect");
    } finally {
      this.busy = false;
    }
  }
}
