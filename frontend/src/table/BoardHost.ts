import { Application } from "pixi.js";
import { clear } from "../dom/dom";
import { HexBoard } from "../hex/HexBoard";
import type { Hex } from "../hex/coords";
import { colors } from "../ui/theme";

export type BoardHostCallbacks = {
  onPick: (hex: Hex) => void;
  onHover: (hex: Hex | null) => void;
  onRotate: (delta: number) => void;
};

export class BoardHost {
  readonly element: HTMLElement;
  private pixi: Application | null = null;
  private hexBoard: HexBoard | null = null;
  private resizeObserver: ResizeObserver | null = null;
  private wheelHandler: ((e: WheelEvent) => void) | null = null;

  constructor() {
    this.element = document.createElement("div");
    this.element.id = "board-host";
  }

  get board(): HexBoard | null {
    return this.hexBoard;
  }

  async mount(callbacks: BoardHostCallbacks): Promise<void> {
    this.destroy();
    await new Promise<void>((r) =>
      requestAnimationFrame(() => requestAnimationFrame(() => r())),
    );

    const w = Math.max(280, this.element.clientWidth || 280);
    const h = Math.max(280, this.element.clientHeight || 280);
    const app = new Application();
    await app.init({
      background: colors.bg,
      antialias: true,
      resolution: window.devicePixelRatio || 1,
      autoDensity: true,
      width: w,
      height: h,
    });
    this.element.appendChild(app.canvas);
    const board = new HexBoard({ size: 36 });
    board.onPick = (hex) => callbacks.onPick(hex);
    board.onHover = (hex) => callbacks.onHover(hex);
    app.stage.addChild(board);
    this.pixi = app;
    this.hexBoard = board;

    this.wheelHandler = (e: WheelEvent) => {
      e.preventDefault();
      if (e.deltaY === 0) return;
      callbacks.onRotate(e.deltaY > 0 ? 1 : -1);
    };
    this.element.addEventListener("wheel", this.wheelHandler, {
      passive: false,
    });

    this.resizeObserver = new ResizeObserver(() => this.layout());
    this.resizeObserver.observe(this.element);
    this.layout();
  }

  destroy(): void {
    this.resizeObserver?.disconnect();
    this.resizeObserver = null;
    if (this.wheelHandler) {
      this.element.removeEventListener("wheel", this.wheelHandler);
      this.wheelHandler = null;
    }
    if (this.pixi) {
      this.pixi.destroy(true, { children: true });
      this.pixi = null;
    }
    this.hexBoard = null;
    clear(this.element);
  }

  layout(): void {
    if (!this.hexBoard || !this.pixi) return;
    const w = Math.max(1, Math.floor(this.element.clientWidth));
    const h = Math.max(1, Math.floor(this.element.clientHeight));
    this.pixi.renderer.resize(w, h);
    this.hexBoard.fit(w, h);
    this.hexBoard.position.set(w / 2, h / 2);
  }
}
