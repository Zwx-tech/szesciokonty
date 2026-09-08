import { Container } from "pixi.js";
import { Scene, type SceneId } from "./Scene";
import type { Session } from "../session";
import { NameScene } from "../scenes/NameScene";
import { HubScene } from "../scenes/HubScene";
import { LobbyScene } from "../scenes/LobbyScene";
import { TableScene } from "../scenes/TableScene";
import { ResultScene } from "../scenes/ResultScene";

export class SceneManager {
  readonly root = new Container();
  readonly session: Session;
  private current: Scene | null = null;
  private currentId: SceneId | null = null;
  private width = 0;
  private height = 0;

  constructor(session: Session) {
    this.session = session;
    session.subscribe(() => this.onSession());
  }

  show(id: SceneId): void {
    if (this.currentId === id) {
      this.current?.onSessionUpdate();
      return;
    }
    const next = this.create(id);
    this.current?.onLeave();
    if (this.current) {
      this.root.removeChild(this.current);
      this.current.destroy({ children: true });
    }
    this.current = next;
    this.currentId = id;
    this.root.addChild(next);
    next.onEnter();
    next.onResize(this.width, this.height);
  }

  resize(width: number, height: number): void {
    this.width = width;
    this.height = height;
    this.current?.onResize(width, height);
  }

  private onSession(): void {
    const { room, error } = this.session;

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
      (this.currentId === "lobby" || this.currentId === "name")
    ) {
      this.show("hub");
      return;
    }
    this.current?.onSessionUpdate();
  }

  private create(id: SceneId): Scene {
    switch (id) {
      case "name":
        return new NameScene(this);
      case "hub":
        return new HubScene(this);
      case "lobby":
        return new LobbyScene(this);
      case "table":
        return new TableScene(this);
      case "result":
        return new ResultScene(this);
    }
  }
}
