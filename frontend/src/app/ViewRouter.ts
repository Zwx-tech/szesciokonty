import type { Session } from "../session";
import { View, type ViewId } from "./View";
import { NameView } from "../views/NameView";
import { HubView } from "../views/HubView";
import { LobbyView } from "../views/LobbyView";
import { TableView } from "../views/TableView";
import { ResultView } from "../views/ResultView";
import { TilesView } from "../views/TilesView";

export class ViewRouter {
  readonly session: Session;
  private readonly host: HTMLElement;
  private current: View | null = null;
  private currentId: ViewId | null = null;

  constructor(session: Session, host: HTMLElement) {
    this.session = session;
    this.host = host;
    session.subscribe(() => this.onSession());
  }

  show(id: ViewId): void {
    if (this.currentId === id) {
      this.current?.onSessionUpdate();
      return;
    }
    this.current?.onLeave();
    this.current?.root.remove();
    const next = this.create(id);
    this.current = next;
    this.currentId = id;
    this.host.appendChild(next.root);
    next.onEnter();
    next.onResize();
  }

  resize(): void {
    this.current?.onResize();
  }

  private onSession(): void {
    // Stay on the tile gallery until the user leaves it.
    if (this.currentId === "tiles") {
      this.current?.onSessionUpdate();
      return;
    }

    const { room, match, error } = this.session;

    if (room?.phase === "match" && match?.phase === "ended") {
      this.show("result");
      return;
    }
    if (room?.phase === "match") {
      this.show("table");
      return;
    }
    if (room?.phase === "lobby") {
      this.show("lobby");
      return;
    }
    if (
      !room &&
      error &&
      (this.currentId === "lobby" ||
        this.currentId === "name" ||
        this.currentId === "result" ||
        this.currentId === "table")
    ) {
      this.show("hub");
      return;
    }
    this.current?.onSessionUpdate();
  }

  private create(id: ViewId): View {
    switch (id) {
      case "name":
        return new NameView(this);
      case "hub":
        return new HubView(this);
      case "lobby":
        return new LobbyView(this);
      case "table":
        return new TableView(this);
      case "result":
        return new ResultView(this);
      case "tiles":
        return new TilesView(this);
    }
  }
}
