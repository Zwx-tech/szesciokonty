import type { Session } from "../session";
import { View, type ViewId } from "./View";
import { NameView } from "../views/NameView";
import { HubView } from "../views/HubView";
import { LobbyView } from "../views/LobbyView";
import { TableView } from "../table/TableView";
import { ResultView } from "../views/ResultView";
import { TilesView } from "../views/TilesView";

type ViewFactory = (router: ViewRouter) => View;

export class ViewRouter {
  readonly session: Session;
  private readonly host: HTMLElement;
  private current: View | null = null;
  private currentId: ViewId | null = null;
  private readonly registry: Record<ViewId, ViewFactory> = {
    name: (r) => new NameView(r),
    hub: (r) => new HubView(r),
    lobby: (r) => new LobbyView(r),
    table: (r) => new TableView(r),
    result: (r) => new ResultView(r),
    tiles: (r) => new TilesView(r),
  };

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
    const factory = this.registry[id];
    return factory(this);
  }
}
