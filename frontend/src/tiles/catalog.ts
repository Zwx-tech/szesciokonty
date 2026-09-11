/** Tile def shape returned by GET /api/tiles/{id}. */
export type TileEffect = {
  kind: string;
  value?: number;
};

export type TileComponent = {
  type: string;
  attackType?: string;
  strength?: number;
  dirs?: number[];
  instant?: string;
  id?: string;
  effects?: TileEffect[];
  params?: Record<string, number>;
};

export type TileDef = {
  id: string;
  kind: string;
  initiatives?: number[];
  toughness: number;
  components: TileComponent[];
};

const cache = new Map<string, TileDef>();
const pending = new Map<string, Promise<TileDef | null>>();

export async function fetchTileDef(defId: string): Promise<TileDef | null> {
  const hit = cache.get(defId);
  if (hit) return hit;
  let p = pending.get(defId);
  if (!p) {
    p = (async () => {
      try {
        const res = await fetch(`/api/tiles/${encodeURIComponent(defId)}`);
        if (!res.ok) return null;
        const def = (await res.json()) as TileDef;
        cache.set(defId, def);
        return def;
      } catch {
        return null;
      } finally {
        pending.delete(defId);
      }
    })();
    pending.set(defId, p);
  }
  return p;
}

export function prefetchTileDefs(defIds: Iterable<string>): void {
  for (const id of defIds) {
    if (!id || cache.has(id)) continue;
    void fetchTileDef(id);
  }
}

export function cachedTileDef(defId: string): TileDef | null {
  return cache.get(defId) ?? null;
}
