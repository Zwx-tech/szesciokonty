import { Scene } from "../app/Scene";
import type { SceneManager } from "../app/SceneManager";
import { Button, Label } from "../ui/primitives";
import { fonts } from "../ui/theme";

export class TableScene extends Scene {
  private title = new Label("Table", { size: fonts.title });
  private stub = new Label("Board arrives in t6 / t9", { muted: true });
  private info = new Label("", { muted: true, size: 14 });
  private endBtn: Button;

  constructor(scenes: SceneManager) {
    super(scenes);
    this.endBtn = new Button("End (stub)", () => this.scenes.show("result"));
    this.addChild(this.title, this.stub, this.info, this.endBtn);
  }

  onEnter(): void {
    this.sync();
  }

  onSessionUpdate(): void {
    this.sync();
  }

  onResize(): void {
    this.title.position.set(24, 24);
    this.stub.position.set(24, 72);
    this.info.position.set(24, 100);
    this.endBtn.position.set(24, 140);
  }

  private sync(): void {
    const room = this.session.room;
    this.info.setText(room ? `Room ${room.code} · match started` : "");
  }
}
