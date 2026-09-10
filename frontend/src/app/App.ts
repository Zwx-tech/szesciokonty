import { Session } from "../session";
import { ViewRouter } from "./ViewRouter";

export class App {
  readonly session = new Session();
  readonly router: ViewRouter;

  constructor(host: HTMLElement) {
    this.router = new ViewRouter(this.session, host);
  }

  start(): void {
    window.addEventListener("resize", () => this.router.resize());
    this.router.show("name");
  }
}
