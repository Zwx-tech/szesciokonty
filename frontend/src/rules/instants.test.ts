import { describe, expect, it } from "vitest";
import { HEX_DIRS, hexAdd, hexKey, onBoard, type Hex } from "../hex/coords";
import type { BoardTile, HandTile, MatchState, PlayerView } from "../protocol";
import {
  aimFor,
  aimHint,
  aimLegalHexes,
  hexDist,
  instantKind,
  type Aim,
} from "./instants";

const YOU = "p1";
const OPP = "p2";

function tile(
  partial: Pick<BoardTile, "id" | "defId" | "kind" | "ownerId" | "q" | "r"> &
    Partial<BoardTile>,
): BoardTile {
  return {
    facing: 0,
    wounds: 0,
    ...partial,
  };
}

function player(id: string, army: "red" | "blue" = "red"): PlayerView {
  return {
    id,
    army: id === YOU ? army : "blue",
    hqHp: 20,
    handCount: 0,
    deckCount: 0,
    discardCount: 0,
  };
}

/** Fixture: your HQ at 0,0; your unit at 1,0; enemy unit at -1,0; enemy HQ at 0,1. */
function match(board: BoardTile[]): MatchState {
  return {
    v: 1,
    type: "match_state",
    phase: "turn",
    turnPlayerId: YOU,
    mustDiscard: false,
    unluckyAvailable: false,
    board,
    players: [player(YOU), player(OPP, "blue")],
  };
}

const fixtureBoard: BoardTile[] = [
  tile({ id: "hq1", defId: "red_hq", kind: "hq", ownerId: YOU, q: 0, r: 0 }),
  tile({
    id: "u1",
    defId: "red_soldier",
    kind: "warrior",
    ownerId: YOU,
    q: 1,
    r: 0,
  }),
  tile({
    id: "e1",
    defId: "blue_soldier",
    kind: "warrior",
    ownerId: OPP,
    q: -1,
    r: 0,
  }),
  tile({ id: "hq2", defId: "blue_hq", kind: "hq", ownerId: OPP, q: 0, r: 1 }),
];

function keys(hexes: Hex[]): string[] {
  return hexes.map(hexKey).sort();
}

describe("instantKind / aimFor / aimHint", () => {
  it("maps defId short names to kinds", () => {
    expect(instantKind("red_battle")).toBe("battle");
    expect(instantKind("blue_shot")).toBe("sniper");
    expect(instantKind("green_blast")).toBe("airstrike");
    expect(instantKind("yellow_move")).toBe("move");
    expect(instantKind("red_soldier")).toBeNull();
  });

  it("aimFor only aims instants", () => {
    const warrior: HandTile = {
      id: "h1",
      defId: "red_soldier",
      kind: "warrior",
    };
    const battle: HandTile = { id: "h2", defId: "red_battle", kind: "instant" };
    const shot: HandTile = { id: "h3", defId: "red_shot", kind: "instant" };
    expect(aimFor(warrior)).toBeNull();
    expect(aimFor(battle)).toEqual({ kind: "battle" });
    expect(aimFor(shot)).toEqual({ kind: "sniper" });
  });

  it("aimHint describes each aim stage", () => {
    expect(aimHint(null)).toBe("");
    expect(aimHint({ kind: "battle" })).toContain("Play");
    expect(aimHint({ kind: "move" })).toContain("your unit");
    expect(aimHint({ kind: "move", unitId: "u1" })).toContain("destination");
  });
});

describe("hexDist", () => {
  it("computes cube distance", () => {
    expect(hexDist(0, 0, 0, 0)).toBe(0);
    expect(hexDist(0, 0, 1, 0)).toBe(1);
    expect(hexDist(0, 0, 2, -1)).toBe(2);
  });
});

describe("aimLegalHexes", () => {
  const m = match(fixtureBoard);

  it("returns empty for null / battle aim", () => {
    expect(aimLegalHexes(null, m, YOU)).toEqual([]);
    expect(aimLegalHexes({ kind: "battle" }, m, YOU)).toEqual([]);
  });

  it("move: first click is your units; second is self + empty neighbors", () => {
    expect(keys(aimLegalHexes({ kind: "move" }, m, YOU))).toEqual(
      keys([
        { q: 0, r: 0 },
        { q: 1, r: 0 },
      ]),
    );
    const after: Aim = { kind: "move", unitId: "u1" };
    const legal = aimLegalHexes(after, m, YOU);
    expect(keys(legal)).toContain("1,0");
    expect(keys(legal)).not.toContain("0,0"); // occupied by HQ
    expect(keys(legal)).not.toContain("-1,0"); // occupied by enemy
    expect(legal.length).toBeGreaterThan(1);
  });

  it("push: first click is enemies adjacent to friendlies", () => {
    // e1 (-1,0) touches your HQ; enemy HQ (0,1) touches your unit at (1,0)
    expect(keys(aimLegalHexes({ kind: "push" }, m, YOU))).toEqual(
      keys([
        { q: -1, r: 0 },
        { q: 0, r: 1 },
      ]),
    );
    const after: Aim = { kind: "push", unitId: "e1" };
    const dest = aimLegalHexes(after, m, YOU);
    expect(keys(dest)).not.toContain("-1,0");
    expect(keys(dest)).not.toContain("0,0");
    expect(
      dest.every((h) => !fixtureBoard.some((t) => t.q === h.q && t.r === h.r)),
    ).toBe(true);
  });

  it("sniper: enemy non-HQ units only", () => {
    expect(keys(aimLegalHexes({ kind: "sniper" }, m, YOU))).toEqual(["-1,0"]);
  });

  it("grenade: enemy units adjacent to your HQ", () => {
    expect(keys(aimLegalHexes({ kind: "grenade" }, m, YOU))).toEqual(["-1,0"]);
  });

  it("grenade: empty when you have no HQ on board", () => {
    const noHq = match(fixtureBoard.filter((t) => t.id !== "hq1"));
    expect(aimLegalHexes({ kind: "grenade" }, noHq, YOU)).toEqual([]);
  });

  it("airstrike: only centers whose full neighborhood is on the board", () => {
    const centers = aimLegalHexes({ kind: "airstrike" }, m, YOU);
    expect(centers.length).toBeGreaterThan(0);
    expect(keys(centers)).toContain("0,0");
    expect(keys(centers)).not.toContain("2,0");
    expect(keys(centers)).not.toContain("-2,0");
    for (const c of centers) {
      expect(onBoard(c)).toBe(true);
      for (const d of HEX_DIRS) {
        expect(onBoard(hexAdd(c, d))).toBe(true);
      }
    }
  });
});
