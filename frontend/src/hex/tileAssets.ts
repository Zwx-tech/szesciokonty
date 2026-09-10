import { Assets, type Texture } from "pixi.js";

/** Circumradius of the hex body in generated tile SVGs (`gentiles` hexSize). */
export const TILE_SVG_HEX_SIZE = 40;

/** Full SVG canvas size (viewBox width/height from gentiles). */
export const TILE_SVG_VIEW = 104;

export function tileSvgUrl(defId: string): string {
  return `/tiles/${encodeURIComponent(defId)}.svg`;
}

const textureLoads = new Map<string, Promise<Texture>>();

/**
 * Load a tile SVG as a raster texture.
 * Prefer texture over GraphicsContext: Pixi's SVG polygon parser uses parseInt and
 * corrupts decimal coordinates in our generated tiles.
 */
export function loadTileTexture(defId: string): Promise<Texture> {
  let pending = textureLoads.get(defId);
  if (!pending) {
    pending = Assets.load({
      alias: `tile-tex:${defId}`,
      src: tileSvgUrl(defId),
      data: { resolution: 3 },
    }) as Promise<Texture>;
    textureLoads.set(defId, pending);
  }
  return pending;
}
