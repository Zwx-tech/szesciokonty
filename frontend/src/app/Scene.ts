import { Container } from "pixi.js";
import type { SceneManager } from "./SceneManager";
import type { Session } from "../session";

export abstract class Scene extends Container {
  protected readonly scenes: SceneManager;
  protected readonly session: Session;

  constructor(scenes: SceneManager) {
    super();
    this.scenes = scenes;
    this.session = scenes.session;
  }

  abstract onEnter(): void;

  onLeave(): void {}

  onSessionUpdate(): void {}

  onResize(width: number, height: number): void {
    void width;
    void height;
  }
}

export type SceneId = "name" | "hub" | "lobby" | "table" | "result";
