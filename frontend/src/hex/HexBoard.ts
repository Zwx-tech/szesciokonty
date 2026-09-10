import { Container, Graphics, Sprite } from "pixi.js";
import {
  BOARD_CELLS,
  hexCorners,
  hexEq,
  hexKey,
  hexToPixel,
  type Hex,
} from "./coords";
import { loadTileTexture, TILE_SVG_HEX_SIZE } from "./tileAssets";
import { colors } from "../ui/theme";

export type HexBoardOpts = {
  size?: number;
};

export type Occupant = {
  q: number;
  r: number;
  defId: string;
  facing: number;
};

type CellGfx = {
  hex: Hex;
  gfx: Graphics;
};

export class HexBoard extends Container {
  private size: number;
  private cells = new Map<string, CellGfx>();
  private tokens = new Container();
  private legal = new Set<string>();
  private selected: Hex | null = null;
  private hover: Hex | null = null;
  private occupants: Occupant[] = [];
  private paintGen = 0;
  onPick: ((hex: Hex) => void) | null = null;

  constructor(opts: HexBoardOpts = {}) {
    super();
    this.size = opts.size ?? 36;
    this.build();
    this.addChild(this.tokens);
  }

  setSize(size: number): void {
    this.size = size;
    this.rebuild();
    this.paintTokens();
  }

  fit(maxW: number, maxH: number, padding = 16): void {
    const extent = boardPixelExtent(2);
    const size = Math.min(
      (maxW - padding * 2) / extent.w,
      (maxH - padding * 2) / extent.h,
    );
    this.setSize(Math.max(12, size));
  }

  setLegal(hexes: readonly Hex[]): void {
    this.legal = new Set(hexes.map(hexKey));
    this.redrawAll();
  }

  clearLegal(): void {
    this.legal.clear();
    this.redrawAll();
  }

  setSelected(hex: Hex | null): void {
    this.selected = hex;
    this.redrawAll();
  }

  setOccupants(occupants: Occupant[]): void {
    this.occupants = occupants;
    this.paintTokens();
  }

  private build(): void {
    this.cells.clear();
    this.removeChildren();
    this.tokens = new Container();
    for (const hex of BOARD_CELLS) {
      const gfx = new Graphics();
      gfx.eventMode = "static";
      gfx.cursor = "pointer";
      gfx.on("pointertap", () => {
        if (this.legal.size > 0 && !this.legal.has(hexKey(hex))) return;
        this.onPick?.(hex);
      });
      gfx.on("pointerover", () => {
        this.hover = hex;
        this.paint(hex);
      });
      gfx.on("pointerout", () => {
        if (this.hover && hexEq(this.hover, hex)) this.hover = null;
        this.paint(hex);
      });
      this.addChild(gfx);
      this.cells.set(hexKey(hex), { hex, gfx });
      this.paint(hex);
    }
    this.addChild(this.tokens);
  }

  private rebuild(): void {
    for (const { hex } of this.cells.values()) this.paint(hex);
  }

  private redrawAll(): void {
    for (const { hex } of this.cells.values()) this.paint(hex);
  }

  private paint(hex: Hex): void {
    const cell = this.cells.get(hexKey(hex));
    if (!cell) return;
    const { x, y } = hexToPixel(hex, this.size);
    const corners = hexCorners(this.size);
    const key = hexKey(hex);
    const isLegal = this.legal.size === 0 || this.legal.has(key);
    const isSelected = this.selected !== null && hexEq(this.selected, hex);
    const isHover = this.hover !== null && hexEq(this.hover, hex);

    let fill: number = colors.hex.cell;
    if (isSelected) fill = colors.hex.selected;
    else if (isHover && isLegal) fill = colors.hex.hover;
    else if (this.legal.size > 0 && this.legal.has(key))
      fill = colors.hex.legal;
    else if (this.legal.size > 0 && !this.legal.has(key))
      fill = colors.hex.blocked;

    const g = cell.gfx;
    g.clear();
    g.poly(corners.flatMap((c) => [c.x, c.y]));
    g.fill(fill);
    g.stroke({ width: 1, color: colors.hex.stroke });
    g.position.set(x, y);
    g.cursor = isLegal ? "pointer" : "default";
  }

  private paintTokens(): void {
    const gen = ++this.paintGen;
    const occupants = this.occupants;
    void this.renderTokens(gen, occupants);
  }

  private clearTokenChildren(): void {
    for (const child of this.tokens.removeChildren()) {
      child.destroy({ children: true, texture: false });
    }
  }

  private async renderTokens(
    gen: number,
    occupants: Occupant[],
  ): Promise<void> {
    if (occupants.length === 0) {
      if (gen === this.paintGen) this.clearTokenChildren();
      return;
    }

    let textures;
    try {
      textures = await Promise.all(
        occupants.map((o) => loadTileTexture(o.defId)),
      );
    } catch (err) {
      console.warn("Failed to load tile SVGs", err);
      return;
    }
    if (gen !== this.paintGen) return;

    this.clearTokenChildren();
    const tokenSize = this.size * 0.82;
    const scale = tokenSize / TILE_SVG_HEX_SIZE;

    for (let i = 0; i < occupants.length; i++) {
      const o = occupants[i]!;
      const texture = textures[i]!;
      const { x, y } = hexToPixel({ q: o.q, r: o.r }, this.size);
      const sprite = new Sprite({
        texture,
        anchor: 0.5,
        eventMode: "none",
      });
      sprite.scale.set(scale);
      sprite.rotation = (((o.facing % 6) + 6) % 6) * (Math.PI / 3);
      sprite.position.set(x, y);
      this.tokens.addChild(sprite);
    }
  }
}

function boardPixelExtent(radius: number): { w: number; h: number } {
  const sqrt3 = Math.sqrt(3);
  return {
    w: sqrt3 * (2 * radius + 1),
    h: 2 + 1.5 * 2 * radius,
  };
}
