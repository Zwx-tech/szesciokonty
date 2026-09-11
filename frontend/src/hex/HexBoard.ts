import { Container, Graphics, Sprite, Text } from "pixi.js";
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
  id?: string;
  q: number;
  r: number;
  defId: string;
  facing: number;
  wounds?: number;
  hp?: number;
  maxHp?: number;
  netted?: boolean;
  isHQ?: boolean;
  hqHp?: number;
};

export type PlacementPreview = {
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
  private previewLayer = new Container();
  private legal = new Set<string>();
  private selected: Hex | null = null;
  private hover: Hex | null = null;
  private occupants: Occupant[] = [];
  private placement: PlacementPreview | null = null;
  private previewSprite: Sprite | null = null;
  private previewGen = 0;
  private paintGen = 0;
  onPick: ((hex: Hex) => void) | null = null;
  onHover: ((hex: Hex | null) => void) | null = null;

  constructor(opts: HexBoardOpts = {}) {
    super();
    this.size = opts.size ?? 36;
    this.build();
    this.addChild(this.tokens);
    this.previewLayer.eventMode = "none";
    this.addChild(this.previewLayer);
  }

  setSize(size: number): void {
    this.size = size;
    this.rebuild();
    this.paintTokens();
    void this.refreshPreview();
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
    void this.refreshPreview();
  }

  clearLegal(): void {
    this.legal.clear();
    this.redrawAll();
    void this.refreshPreview();
  }

  setSelected(hex: Hex | null): void {
    this.selected = hex;
    this.redrawAll();
  }

  setOccupants(occupants: Occupant[]): void {
    this.occupants = occupants;
    this.paintTokens();
  }

  setPlacementPreview(preview: PlacementPreview | null): void {
    this.placement = preview;
    void this.refreshPreview();
  }

  private build(): void {
    this.cells.clear();
    this.removeChildren();
    this.tokens = new Container();
    this.previewLayer = new Container();
    this.previewLayer.eventMode = "none";
    this.previewSprite = null;
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
        this.onHover?.(hex);
        void this.refreshPreview();
      });
      gfx.on("pointerout", () => {
        if (this.hover && hexEq(this.hover, hex)) this.hover = null;
        this.paint(hex);
        this.onHover?.(null);
        void this.refreshPreview();
      });
      this.addChild(gfx);
      this.cells.set(hexKey(hex), { hex, gfx });
      this.paint(hex);
    }
    this.addChild(this.tokens);
    this.addChild(this.previewLayer);
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
    const fontSize = Math.max(9, Math.round(this.size * 0.28));

    for (let i = 0; i < occupants.length; i++) {
      const o = occupants[i]!;
      const texture = textures[i]!;
      const { x, y } = hexToPixel({ q: o.q, r: o.r }, this.size);
      const wrap = new Container();
      const stackIndex = stackIndexAt(occupants, i);
      const stackCount = stackCountAt(occupants, o.q, o.r);
      const offset =
        stackCount > 1 ? (stackIndex - (stackCount - 1) / 2) * this.size * 0.22 : 0;
      wrap.position.set(x + offset, y - Math.abs(offset) * 0.15);
      wrap.eventMode = "none";

      const sprite = new Sprite({
        texture,
        anchor: 0.5,
        eventMode: "none",
      });
      const stackScale = stackCount > 1 ? 0.88 : 1;
      sprite.scale.set(scale * stackScale);
      sprite.rotation = (((o.facing % 6) + 6) % 6) * (Math.PI / 3);
      if (o.netted) sprite.tint = 0xcccc66;
      wrap.addChild(sprite);

      const badge = hpBadgeText(o);
      if (badge) {
        const label = new Text({
          text: badge,
          style: {
            fontFamily: "Arial, Helvetica, sans-serif",
            fontSize,
            fontWeight: "700",
            fill: 0xffffff,
            stroke: { color: 0x111111, width: 3 },
          },
        });
        label.anchor.set(0.5);
        label.position.set(0, this.size * 0.42);
        wrap.addChild(label);
      }

      if (stackCount > 1 && stackIndex === 0) {
        const stack = new Text({
          text: String(stackCount),
          style: {
            fontFamily: "Arial, Helvetica, sans-serif",
            fontSize: Math.max(8, fontSize - 1),
            fontWeight: "700",
            fill: 0xffe08a,
            stroke: { color: 0x111111, width: 3 },
          },
        });
        stack.anchor.set(0.5);
        stack.position.set(this.size * 0.38, -this.size * 0.38);
        wrap.addChild(stack);
      }

      if (o.netted) {
        const net = new Text({
          text: "NET",
          style: {
            fontFamily: "Arial, Helvetica, sans-serif",
            fontSize: Math.max(8, fontSize - 1),
            fontWeight: "700",
            fill: colors.edge.net,
            stroke: { color: 0x111111, width: 3 },
          },
        });
        net.anchor.set(0.5);
        net.position.set(0, -this.size * 0.4);
        wrap.addChild(net);
      }

      this.tokens.addChild(wrap);
    }
  }

  private clearPreviewSprite(): void {
    if (this.previewSprite) {
      this.previewSprite.destroy({ texture: false });
      this.previewSprite = null;
    }
    this.previewLayer.removeChildren();
  }

  private async refreshPreview(): Promise<void> {
    const gen = ++this.previewGen;
    const placement = this.placement;
    const hex = this.hover;

    const show =
      placement !== null && hex !== null && this.legal.has(hexKey(hex));

    if (!show) {
      if (gen === this.previewGen) this.clearPreviewSprite();
      return;
    }

    let texture;
    try {
      texture = await loadTileTexture(placement.defId);
    } catch (err) {
      console.warn("Failed to load preview tile SVG", err);
      return;
    }
    if (gen !== this.previewGen) return;

    const { x, y } = hexToPixel(hex, this.size);
    const tokenSize = this.size * 0.82;
    const scale = tokenSize / TILE_SVG_HEX_SIZE;
    const facing = ((placement.facing % 6) + 6) % 6;

    if (this.previewSprite && this.previewSprite.texture === texture) {
      this.previewSprite.position.set(x, y);
      this.previewSprite.scale.set(scale);
      this.previewSprite.rotation = facing * (Math.PI / 3);
      this.previewSprite.alpha = 0.55;
      return;
    }

    this.clearPreviewSprite();
    const sprite = new Sprite({
      texture,
      anchor: 0.5,
      eventMode: "none",
      alpha: 0.55,
    });
    sprite.scale.set(scale);
    sprite.rotation = facing * (Math.PI / 3);
    sprite.position.set(x, y);
    this.previewSprite = sprite;
    this.previewLayer.addChild(sprite);
  }
}

function hpBadgeText(o: Occupant): string | null {
  if (o.isHQ && o.hqHp != null) return `HQ ${o.hqHp}`;
  if (o.maxHp != null && o.maxHp > 0 && o.hp != null) {
    if (o.maxHp > 1 || (o.wounds ?? 0) > 0) return `${o.hp}/${o.maxHp}`;
  }
  return null;
}

function stackCountAt(occupants: Occupant[], q: number, r: number): number {
  let n = 0;
  for (const o of occupants) {
    if (o.q === q && o.r === r) n++;
  }
  return n;
}

function stackIndexAt(occupants: Occupant[], index: number): number {
  const o = occupants[index]!;
  let idx = 0;
  for (let i = 0; i < index; i++) {
    const other = occupants[i]!;
    if (other.q === o.q && other.r === o.r) idx++;
  }
  return idx;
}

function boardPixelExtent(radius: number): { w: number; h: number } {
  const sqrt3 = Math.sqrt(3);
  return {
    w: sqrt3 * (2 * radius + 1),
    h: 2 + 1.5 * 2 * radius,
  };
}
