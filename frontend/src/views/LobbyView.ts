import { el, setDisabled } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import type { Army } from "../protocol";

const ARMIES: Army[] = ["red", "blue", "green", "yellow"];

export class LobbyView extends View {
  private codeLabel: HTMLElement;
  private seatsLabel: HTMLElement;
  private statusLabel: HTMLElement;
  private errorLabel: HTMLElement;
  private readyBtn: HTMLButtonElement;
  private startBtn: HTMLButtonElement;
  private armyButtons = new Map<Army, HTMLButtonElement>();

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");

    const panel = el("div", { className: "panel stack" });
    panel.append(el("h1", { text: "Lobby" }));

    this.codeLabel = el("h2", { text: "Code: —" });
    panel.append(this.codeLabel);

    this.seatsLabel = el("div", { className: "seats" });
    panel.append(this.seatsLabel);

    panel.append(el("div", { className: "muted", text: "Army" }));
    const armyRow = el("div", { className: "row" });
    for (const army of ARMIES) {
      const btn = el("button", {
        className: `army-${army}`,
        text: army,
      });
      btn.addEventListener("click", () => this.session.setArmy(army));
      armyRow.append(btn);
      this.armyButtons.set(army, btn);
    }
    panel.append(armyRow);

    const actions = el("div", { className: "row" });
    this.readyBtn = el("button", { text: "Ready" });
    this.readyBtn.addEventListener("click", () => {
      if (!this.session.room?.you.army) {
        this.session.setError("Select an army first");
        return;
      }
      this.session.setReady(!this.session.room.you.ready);
    });
    this.startBtn = el("button", { text: "Start" });
    this.startBtn.disabled = true;
    this.startBtn.addEventListener("click", () => this.session.start());
    actions.append(this.readyBtn, this.startBtn);
    panel.append(actions);

    const leaveBtn = el("button", { text: "Leave" });
    leaveBtn.addEventListener("click", () => {
      this.session.leaveRoom();
      this.router.show("hub");
    });
    panel.append(leaveBtn);

    this.statusLabel = el("div", { className: "muted" });
    panel.append(this.statusLabel);

    this.errorLabel = el("div", { className: "error" });
    panel.append(this.errorLabel);

    this.root.append(panel);
  }

  onEnter(): void {
    this.sync();
  }

  onSessionUpdate(): void {
    this.sync();
  }

  private sync(): void {
    const room = this.session.room;
    if (!room) {
      this.codeLabel.textContent = "Code: —";
      this.seatsLabel.textContent = "No room";
      return;
    }

    this.codeLabel.textContent = `Code: ${room.code}`;
    this.seatsLabel.textContent =
      room.seats
        .map((s) => {
          const bits = [
            s.name,
            s.host ? "host" : "guest",
            s.army ?? "no army",
            s.ready ? "ready" : "not ready",
            s.connected ? "" : "offline",
          ].filter(Boolean);
          return bits.join(" · ");
        })
        .join("\n") || "Waiting…";

    const taken = new Set(
      room.seats
        .filter((s) => s.id !== room.you.id && s.army)
        .map((s) => s.army),
    );
    for (const [army, btn] of this.armyButtons) {
      setDisabled(btn, taken.has(army));
      btn.classList.toggle("selected", room.you.army === army);
    }

    this.readyBtn.textContent = room.you.ready ? "Unready" : "Ready";
    setDisabled(this.startBtn, !(room.you.host && room.canStart));

    this.statusLabel.textContent = !room.you.army
      ? "Select an army"
      : room.seats.length < 2
        ? "Waiting for opponent…"
        : room.canStart
          ? "Ready to start"
          : room.you.ready
            ? "Waiting for opponent ready…"
            : "Ready up when set";
    this.errorLabel.textContent = this.session.error ?? "";
  }
}
