import type { Session } from "../session";
import type { ViewRouter } from "./ViewRouter";

export type ViewId = "name" | "hub" | "lobby" | "table" | "result" | "tiles";

export abstract class View {
  protected readonly router: ViewRouter;
  protected readonly session: Session;
  readonly root: HTMLElement;

  constructor(router: ViewRouter) {
    this.router = router;
    this.session = router.session;
    this.root = document.createElement("div");
    this.root.className = "view";
  }

  abstract onEnter(): void;

  onLeave(): void {}

  onSessionUpdate(): void {}

  onResize(): void {}
}
