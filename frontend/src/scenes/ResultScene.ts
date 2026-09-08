import { Scene } from "../app/Scene";
import type { SceneManager } from "../app/SceneManager";
import { Button, Label, Panel } from "../ui/primitives";
import { fonts } from "../ui/theme";

export class ResultScene extends Scene {
  private panel: Panel;

  constructor(scenes: SceneManager) {
    super(scenes);

    this.panel = new Panel({ width: 360, height: 220 });
    this.addChild(this.panel);

    const title = new Label("Result", { size: fonts.title });
    title.position.set(24, 24);
    this.panel.content.addChild(title);

    const stub = new Label("Outcome wired in t12", { muted: true });
    stub.position.set(24, 90);
    this.panel.content.addChild(stub);

    const rematch = new Button("Rematch", () => this.scenes.show("lobby"));
    rematch.position.set(24, 140);
    this.panel.content.addChild(rematch);

    const leave = new Button("Leave", () => this.scenes.show("hub"), {
      width: 120,
    });
    leave.position.set(240, 140);
    this.panel.content.addChild(leave);
  }

  onEnter(): void {}

  onResize(width: number, height: number): void {
    this.panel.position.set((width - 360) / 2, (height - 220) / 2);
  }
}
