import { Application } from "pixi.js";
import { SceneManager } from "./SceneManager";
import { Session } from "../session";
import { colors } from "../ui/theme";

export class GameApp {
  readonly pixi = new Application();
  readonly session = new Session();
  readonly scenes = new SceneManager(this.session);

  async init(host: HTMLElement): Promise<void> {
    await this.pixi.init({
      resizeTo: window,
      background: colors.bg,
      antialias: true,
      resolution: window.devicePixelRatio,
      autoDensity: true,
    });
    host.appendChild(this.pixi.canvas);
    this.pixi.stage.addChild(this.scenes.root);

    const resize = () => {
      this.scenes.resize(this.pixi.screen.width, this.pixi.screen.height);
    };
    window.addEventListener("resize", resize);
    resize();

    this.scenes.show("name");
  }
}
