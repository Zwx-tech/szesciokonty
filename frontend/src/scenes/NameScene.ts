import { Scene } from "../app/Scene";
import type { SceneManager } from "../app/SceneManager";
import { Button, Label, Panel, TextField } from "../ui/primitives";
import { fonts } from "../ui/theme";

export class NameScene extends Scene {
  private panel: Panel;
  private field: TextField;
  private errorLabel: Label;
  private busy = false;

  constructor(scenes: SceneManager) {
    super(scenes);

    this.panel = new Panel({ width: 360, height: 240 });
    this.addChild(this.panel);

    const title = new Label("Hex", { size: fonts.title });
    title.position.set(24, 24);
    this.panel.content.addChild(title);

    const hint = new Label("Display name", { muted: true, size: 14 });
    hint.position.set(24, 80);
    this.panel.content.addChild(hint);

    this.field = new TextField({
      width: 312,
      placeholder: "name",
      maxLength: 16,
    });
    this.field.position.set(24, 104);
    this.panel.content.addChild(this.field);

    this.errorLabel = new Label("", { muted: true, size: 14, color: 0xcc6666 });
    this.errorLabel.position.set(24, 154);
    this.panel.content.addChild(this.errorLabel);

    const continueBtn = new Button("Continue", () => void this.continue());
    continueBtn.position.set(80, 180);
    this.panel.content.addChild(continueBtn);
  }

  onEnter(): void {
    if (this.session.name) this.field.value = this.session.name;
    this.field.focus();
  }

  onLeave(): void {
    this.field.blur();
  }

  onResize(width: number, height: number): void {
    this.panel.position.set((width - 360) / 2, (height - 240) / 2);
  }

  private async continue(): Promise<void> {
    if (this.busy) return;
    const name = this.field.value.trim();
    if (!name) {
      this.errorLabel.setText("Enter a name");
      return;
    }
    this.busy = true;
    this.errorLabel.setText("");
    this.session.setName(name);
    try {
      await this.session.ensureConnected();
      const reconnecting = await this.session.tryReconnect();
      if (!reconnecting) this.scenes.show("hub");
      // reconnect success navigates via room_state → lobby/table
    } catch {
      this.errorLabel.setText("Could not connect");
    } finally {
      this.busy = false;
    }
  }
}
