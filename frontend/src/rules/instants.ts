import { HEX_DIRS, hexAdd, hexKey, onBoard, type Hex } from "../hex/coords";
import type { HandTile, MatchState } from "../protocol";
import { shortDefName } from "../tiles/format";
import { isBlockerTile } from "./abilities";

export type InstantKind =
  | "battle"
  | "move"
  | "push"
  | "sniper"
  | "airstrike"
  | "grenade";

export type Aim =
  | null
  | { kind: "battle" }
  | { kind: "move"; unitId?: string }
  | { kind: "push"; unitId?: string }
  | { kind: "sniper" }
  | { kind: "airstrike" }
  | { kind: "grenade" }
  | { kind: "mobility"; unitId: string }
  | {
      kind: "passenger";
      via: "move" | "mobility";
      handTileId?: string;
      unitId: string;
      q: number;
      r: number;
      facing: number;
      candidates: string[];
    };

export function aimLegalHexes(
  aim: Aim,
  match: MatchState,
  youId: string,
): Hex[] {
  if (!aim) return [];
  const occupied = new Set(match.board.map((t) => hexKey({ q: t.q, r: t.r })));
  const emptyNeighbors = (q: number, r: number): Hex[] => {
    const out: Hex[] = [];
    for (const d of HEX_DIRS) {
      const n = hexAdd({ q, r }, d);
      if (onBoard(n) && !occupied.has(hexKey(n))) out.push(n);
    }
    return out;
  };

  switch (aim.kind) {
    case "move": {
      if (!aim.unitId) {
        return match.board
          .filter((t) => t.ownerId === youId && !isBlockerTile(t))
          .map((t) => ({ q: t.q, r: t.r }));
      }
      const unit = match.board.find((t) => t.id === aim.unitId);
      if (!unit) return [];
      return [{ q: unit.q, r: unit.r }, ...emptyNeighbors(unit.q, unit.r)];
    }
    case "mobility": {
      const unit = match.board.find((t) => t.id === aim.unitId);
      if (!unit) return [];
      return [{ q: unit.q, r: unit.r }, ...emptyNeighbors(unit.q, unit.r)];
    }
    case "passenger": {
      return match.board
        .filter((t) => aim.candidates.includes(t.id))
        .map((t) => ({ q: t.q, r: t.r }));
    }
    case "push": {
      if (!aim.unitId) {
        return match.board
          .filter(
            (t) =>
              t.ownerId !== youId &&
              !isBlockerTile(t) &&
              adjacentToFriendly(match, youId, t.q, t.r),
          )
          .map((t) => ({ q: t.q, r: t.r }));
      }
      const unit = match.board.find((t) => t.id === aim.unitId);
      if (!unit) return [];
      return emptyNeighbors(unit.q, unit.r);
    }
    case "sniper":
      return match.board
        .filter((t) => t.ownerId !== youId && t.kind !== "hq")
        .map((t) => ({ q: t.q, r: t.r }));
    case "grenade": {
      const hq = match.board.find(
        (t) => t.ownerId === youId && t.kind === "hq",
      );
      if (!hq) return [];
      return match.board
        .filter(
          (t) =>
            t.ownerId !== youId &&
            t.kind !== "hq" &&
            hexDist(hq.q, hq.r, t.q, t.r) === 1,
        )
        .map((t) => ({ q: t.q, r: t.r }));
    }
    case "airstrike": {
      const out: Hex[] = [];
      for (let q = -2; q <= 2; q++) {
        for (let r = -2; r <= 2; r++) {
          const center = { q, r };
          if (!onBoard(center)) continue;
          let ok = true;
          for (const d of HEX_DIRS) {
            if (!onBoard(hexAdd(center, d))) {
              ok = false;
              break;
            }
          }
          if (ok) out.push(center);
        }
      }
      return out;
    }
    default:
      return [];
  }
}

export function adjacentToFriendly(
  match: MatchState,
  youId: string,
  q: number,
  r: number,
): boolean {
  for (const d of HEX_DIRS) {
    const n = hexAdd({ q, r }, d);
    if (
      match.board.some((t) => t.ownerId === youId && t.q === n.q && t.r === n.r)
    ) {
      return true;
    }
  }
  return false;
}

export function hexDist(
  aq: number,
  ar: number,
  bq: number,
  br: number,
): number {
  const as_ = -aq - ar;
  const bs = -bq - br;
  return Math.max(Math.abs(aq - bq), Math.abs(ar - br), Math.abs(as_ - bs));
}

export function aimFor(tile: HandTile): Aim {
  if (tile.kind !== "instant") return null;
  const k = instantKind(tile.defId);
  if (!k) return null;
  if (k === "battle") return { kind: "battle" };
  if (k === "move") return { kind: "move" };
  if (k === "push") return { kind: "push" };
  if (k === "sniper") return { kind: "sniper" };
  if (k === "airstrike") return { kind: "airstrike" };
  if (k === "grenade") return { kind: "grenade" };
  return null;
}

export function instantKind(defId: string): InstantKind | null {
  const name = shortDefName(defId);
  if (name === "battle") return "battle";
  if (name === "move") return "move";
  if (name === "push") return "push";
  if (name === "shot") return "sniper";
  if (name === "blast") return "airstrike";
  if (name === "grenade") return "grenade";
  return null;
}

export function aimHint(aim: Aim): string {
  if (!aim) return "";
  switch (aim.kind) {
    case "battle":
      return "Battle: press Play";
    case "move":
      return aim.unitId
        ? "Move: click destination hex"
        : "Move: click your unit";
    case "mobility":
      return "Mobility: click adjacent hex or same hex to rotate";
    case "passenger":
      return "Transport: click passenger (or Skip passenger)";
    case "push":
      return aim.unitId
        ? "Push: click destination hex"
        : "Push: click enemy unit";
    case "sniper":
      return "Shot: click enemy unit";
    case "grenade":
      return "Grenade: click enemy next to your HQ";
    case "airstrike":
      return "Blast: click center hex (must fit on board)";
  }
}
