import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import type { Army } from "../protocol";
import { LobbyForm } from "./LobbyForm";

export class LobbyView extends View {
  private form: LobbyForm;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("view-center");
    this.form = new LobbyForm();
    this.form.onSetArmy = (army) => this.session.setArmy(army);
    this.form.onToggleReady = () => {
      if (!this.session.room?.you.army) {
        this.session.setError("Select an army first");
        return;
      }
      this.session.setReady(!this.session.room.you.ready);
    };
    this.form.onStart = () => this.session.start();
    this.form.onLeave = () => {
      this.session.leaveRoom();
      this.router.show("hub");
    };
    this.root.append(this.form);
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
      this.form.state = {
        ...this.form.state,
        code: "—",
        seats: [],
        status: "No room",
        error: this.session.error ?? "",
      };
      return;
    }

    const taken = new Set(
      room.seats
        .filter((s) => s.id !== room.you.id && s.army)
        .map((s) => s.army as Army),
    );

    const status = !room.you.army
      ? "Select an army"
      : room.seats.length < 2
        ? "Waiting for opponent…"
        : room.canStart
          ? "Ready to start"
          : room.you.ready
            ? "Waiting for opponent ready…"
            : "Ready up when set";

    this.form.state = {
      code: room.code,
      seats: room.seats,
      youArmy: room.you.army,
      youReady: room.you.ready,
      youHost: room.you.host,
      canStart: room.canStart,
      takenArmies: taken,
      status,
      error: this.session.error ?? "",
    };
  }
}
