import { Scene } from "../app/Scene";
import type { SceneManager } from "../app/SceneManager";
import { Button, Label, Panel, TextField } from "../ui/primitives";
import { fonts } from "../ui/theme";

export class HubScene extends Scene {
  private panel: Panel;
  private codeField: TextField;
  private errorLabel: Label;
  private busy = false;

  constructor(scenes: SceneManager) {
    super(scenes);

    this.panel = new Panel({ width: 360, height: 300 });
    this.addChild(this.panel);

    const title = new Label("Room", { size: fonts.title });
    title.position.set(24, 24);
    this.panel.content.addChild(title);

    const createBtn = new Button("Create room", () => void this.create());
    createBtn.position.set(80, 90);
    this.panel.content.addChild(createBtn);

    const joinHint = new Label("Or join with code", { muted: true, size: 14 });
    joinHint.position.set(24, 150);
    this.panel.content.addChild(joinHint);

    this.codeField = new TextField({
      width: 312,
      placeholder: "code",
      maxLength: 8,
    });
    this.codeField.position.set(24, 174);
    this.panel.content.addChild(this.codeField);

    const joinBtn = new Button("Join", () => void this.join());
    joinBtn.position.set(80, 226);
    this.panel.content.addChild(joinBtn);

    this.errorLabel = new Label("", { size: 14, color: 0xcc6666 });
    this.errorLabel.position.set(24, 276);
    this.panel.content.addChild(this.errorLabel);
  }

  onEnter(): void {
    this.syncError();
  }

  onSessionUpdate(): void {
    this.syncError();
  }

  onLeave(): void {
    this.codeField.blur();
  }

  onResize(width: number, height: number): void {
    this.panel.position.set((width - 360) / 2, (height - 300) / 2);
  }

  private syncError(): void {
    this.errorLabel.setText(this.session.error ?? "");
  }

  private async create(): Promise<void> {
    if (this.busy) return;
    this.busy = true;
    try {
      await this.session.ensureConnected();
      this.session.createRoom();
    } catch {
      this.session.setError("Could not connect");
    } finally {
      this.busy = false;
    }
  }

  private async join(): Promise<void> {
    if (this.busy) return;
    const code = this.codeField.value.trim();
    if (!code) {
      this.errorLabel.setText("Enter a room code");
      return;
    }
    this.busy = true;
    try {
      await this.session.ensureConnected();
      this.session.joinRoom(code);
    } catch {
      this.session.setError("Could not connect");
    } finally {
      this.busy = false;
    }
  }
}
