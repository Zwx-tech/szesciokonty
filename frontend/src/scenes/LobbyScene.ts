import { Container } from "pixi.js";
import { Scene } from "../app/Scene";
import type { SceneManager } from "../app/SceneManager";
import type { Army } from "../protocol";
import { Button, Label, Panel } from "../ui/primitives";
import { colors, fonts } from "../ui/theme";

const ARMIES: Army[] = ["red", "blue", "green", "yellow"];

export class LobbyScene extends Scene {
  private panel: Panel;
  private codeLabel: Label;
  private seatsLabel: Label;
  private statusLabel: Label;
  private errorLabel: Label;
  private readyBtn: Button;
  private startBtn: Button;
  private armyButtons = new Map<Army, Button>();
  private armyRow = new Container();

  constructor(scenes: SceneManager) {
    super(scenes);

    this.panel = new Panel({ width: 440, height: 380 });
    this.addChild(this.panel);

    const title = new Label("Lobby", { size: fonts.title });
    title.position.set(24, 20);
    this.panel.content.addChild(title);

    this.codeLabel = new Label("Code: —", { size: 20 });
    this.codeLabel.position.set(24, 70);
    this.panel.content.addChild(this.codeLabel);

    this.seatsLabel = new Label("", { size: 16 });
    this.seatsLabel.position.set(24, 110);
    this.panel.content.addChild(this.seatsLabel);

    const armyHint = new Label("Army", { muted: true, size: 14 });
    armyHint.position.set(24, 170);
    this.panel.content.addChild(armyHint);

    this.armyRow.position.set(24, 194);
    this.panel.content.addChild(this.armyRow);

    ARMIES.forEach((army, i) => {
      const btn = new Button(army, () => this.session.setArmy(army), {
        width: 90,
        height: 40,
        fill: colors.army[army],
        fillHover: colors.army[army],
      });
      btn.position.set(i * 98, 0);
      this.armyRow.addChild(btn);
      this.armyButtons.set(army, btn);
    });

    this.readyBtn = new Button("Ready", () => {
      const ready = !(this.session.room?.you.ready ?? false);
      this.session.setReady(ready);
    });
    this.readyBtn.position.set(24, 250);
    this.panel.content.addChild(this.readyBtn);

    this.startBtn = new Button("Start", () => this.session.start(), {
      disabled: true,
    });
    this.startBtn.position.set(236, 250);
    this.panel.content.addChild(this.startBtn);

    const leaveBtn = new Button(
      "Leave",
      () => {
        this.session.leaveRoom();
        this.scenes.show("hub");
      },
      { width: 120 },
    );
    leaveBtn.position.set(24, 306);
    this.panel.content.addChild(leaveBtn);

    this.statusLabel = new Label("", { muted: true, size: 14 });
    this.statusLabel.position.set(160, 318);
    this.panel.content.addChild(this.statusLabel);

    this.errorLabel = new Label("", { size: 14, color: 0xcc6666 });
    this.errorLabel.position.set(24, 340);
    this.panel.content.addChild(this.errorLabel);
  }

  onEnter(): void {
    this.sync();
  }

  onSessionUpdate(): void {
    this.sync();
  }

  onResize(width: number, height: number): void {
    this.panel.position.set((width - 440) / 2, (height - 380) / 2);
  }

  private sync(): void {
    const room = this.session.room;
    if (!room) {
      this.codeLabel.setText("Code: —");
      this.seatsLabel.setText("No room");
      return;
    }

    this.codeLabel.setText(`Code: ${room.code}`);
    this.seatsLabel.setText(
      room.seats
        .map((s) => {
          const bits = [
            s.name,
            s.host ? "host" : "guest",
            s.army ?? "no army",
            s.ready ? "ready" : "not ready",
            s.connected ? "" : "offline",
          ].filter(Boolean);
          return bits.join(" · ");
        })
        .join("\n") || "Waiting…",
    );

    const taken = new Set(
      room.seats
        .filter((s) => s.id !== room.you.id && s.army)
        .map((s) => s.army),
    );
    for (const [army, btn] of this.armyButtons) {
      btn.disabled = taken.has(army);
      btn.setSelected(room.you.army === army);
    }

    this.readyBtn.setLabel(room.you.ready ? "Unready" : "Ready");
    this.readyBtn.disabled = !room.you.army;
    this.startBtn.disabled = !(room.you.host && room.canStart);

    this.statusLabel.setText(
      room.seats.length < 2
        ? "Waiting for opponent…"
        : room.canStart
          ? "Ready to start"
          : "",
    );
    this.errorLabel.setText(this.session.error ?? "");
  }
}
