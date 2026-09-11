import { el } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";

export class ResultView extends View {
  private title: HTMLElement;
  private detail: HTMLElement;
  private errorLabel: HTMLElement;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");

    const panel = el("div", { className: "panel stack" });
    this.title = el("h1", { text: "Result" });
    panel.append(this.title);

    this.detail = el("div", { className: "muted" });
    panel.append(this.detail);

    this.errorLabel = el("div", { className: "error" });
    panel.append(this.errorLabel);

    const actions = el("div", { className: "row" });
    const rematch = el("button", { text: "Rematch" });
    rematch.addEventListener("click", () => this.session.rematch());
    const leave = el("button", { text: "Leave" });
    leave.addEventListener("click", () => {
      this.session.leaveRoom();
      this.router.show("hub");
    });
    actions.append(rematch, leave);
    panel.append(actions);

    this.root.append(panel);
  }

  onEnter(): void {
    this.sync();
  }

  onSessionUpdate(): void {
    this.sync();
  }

  private sync(): void {
    const match = this.session.match;
    const you = this.session.youId;
    const result = match?.result;

    if (!result) {
      this.title.textContent = "Result";
      this.detail.textContent = "Waiting…";
    } else if (result.draw) {
      this.title.textContent = "Draw";
      this.detail.textContent = this.hpLine(match);
    } else if (result.winnerId === you) {
      this.title.textContent = "Victory";
      this.detail.textContent = this.hpLine(match);
    } else {
      this.title.textContent = "Defeat";
      this.detail.textContent = this.hpLine(match);
    }
    this.errorLabel.textContent = this.session.error ?? "";
  }

  private hpLine(match: typeof this.session.match): string {
    if (!match) return "";
    const room = this.session.room;
    return match.players
      .map((p) => {
        const name = room?.seats.find((s) => s.id === p.id)?.name ?? p.army;
        return `${name} (${p.army}) HQ ${p.hqHp}`;
      })
      .join(" · ");
  }
}
