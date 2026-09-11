import { hexNeighbor } from "../hex/coords";
import type { BoardTile, MatchState } from "../protocol";
import { cachedTileDef, type TileDef } from "../tiles/catalog";

export function hasSpecial(def: TileDef | null, id: string): boolean {
  return (
    def?.components?.some((c) => c.type === "special" && c.id === id) ?? false
  );
}

export function isBlockerTile(t: BoardTile): boolean {
  return hasSpecial(cachedTileDef(t.defId), "blocker");
}

export function isDualStackHand(defId: string): boolean {
  return hasSpecial(cachedTileDef(defId), "dual_stack");
}

export function hasAuraEffect(def: TileDef | null, kind: string): boolean {
  return (
    def?.components?.some(
      (c) =>
        (c.type === "module_aura" || c.type === "hq_aura") &&
        (c.effects ?? []).some((e) => e.kind === kind),
    ) ?? false
  );
}

export function moduleLinksHex(
  mod: BoardTile,
  def: TileDef,
  q: number,
  r: number,
): boolean {
  for (const c of def.components ?? []) {
    if (c.type !== "module_link") continue;
    for (const d of c.dirs ?? []) {
      const n = hexNeighbor({ q: mod.q, r: mod.r }, mod.facing + d);
      if (n.q === q && n.r === r) return true;
    }
  }
  return false;
}

/** Friendly warriors that share a transport module link with the mover (at current hex). */
export function transportPassengers(
  match: MatchState,
  youId: string,
  moverId: string,
): BoardTile[] {
  const mover = match.board.find((t) => t.id === moverId);
  if (!mover) return [];
  const out: BoardTile[] = [];
  const seen = new Set<string>();

  for (const mod of match.board) {
    if (mod.ownerId !== youId || mod.kind !== "module" || mod.netted) continue;
    const def = cachedTileDef(mod.defId);
    if (!def || !hasAuraEffect(def, "transport")) continue;
    if (!moduleLinksHex(mod, def, mover.q, mover.r)) continue;
    for (const t of match.board) {
      if (t.id === moverId || t.ownerId !== youId || t.kind !== "warrior") {
        continue;
      }
      if (seen.has(t.id)) continue;
      if (!moduleLinksHex(mod, def, t.q, t.r)) continue;
      seen.add(t.id);
      out.push(t);
    }
  }
  return out;
}

export function tilesAt(
  match: MatchState,
  q: number,
  r: number,
): BoardTile[] {
  return match.board.filter((t) => t.q === q && t.r === r);
}
