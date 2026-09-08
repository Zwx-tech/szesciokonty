import { Scene } from "../app/Scene";
import type { SceneManager } from "../app/SceneManager";
import { BOARD_CELLS, type Hex } from "../hex/coords";
import { HexBoard } from "../hex/HexBoard";
import { Button, Label } from "../ui/primitives";
import { fonts } from "../ui/theme";

export class TableScene extends Scene {
  private title = new Label("Table", { size: fonts.title });
  private info = new Label("", { muted: true, size: 14 });
  private pickLabel = new Label("Click a hex", { muted: true, size: 14 });
  private board = new HexBoard({ size: 36 });
  private endBtn: Button;
  private viewW = 0;
  private viewH = 0;

  constructor(scenes: SceneManager) {
    super(scenes);
    this.endBtn = new Button(
      "Leave",
      () => {
        this.session.leaveRoom();
        this.scenes.show("hub");
      },
      { width: 120 },
    );

    this.board.setLegal([...BOARD_CELLS]);
    this.board.onPick = (hex) => this.onPick(hex);

    this.addChild(
      this.title,
      this.info,
      this.pickLabel,
      this.board,
      this.endBtn,
    );
  }

  onEnter(): void {
    this.sync();
    this.layout();
  }

  onSessionUpdate(): void {
    this.sync();
  }

  onResize(width: number, height: number): void {
    this.viewW = width;
    this.viewH = height;
    this.layout();
  }

  private layout(): void {
    const w = this.viewW;
    const h = this.viewH;
    this.title.position.set(24, 20);
    this.info.position.set(24, 60);
    this.pickLabel.position.set(24, 84);
    this.endBtn.position.set(w - 144, 20);

    const top = 120;
    const boardH = Math.max(120, h - top - 24);
    this.board.fit(w - 48, boardH);
    this.board.position.set(w / 2, top + boardH / 2);
  }

  private sync(): void {
    const room = this.session.room;
    this.info.setText(room ? `Room ${room.code}` : "");
  }

  private onPick(hex: Hex): void {
    this.board.setSelected(hex);
    this.pickLabel.setText(`Selected q=${hex.q} r=${hex.r}`);
  }
}
