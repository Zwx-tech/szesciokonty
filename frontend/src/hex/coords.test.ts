import { describe, expect, it } from "vitest";
import {
  BOARD_CELLS,
  HEX_DIRS,
  boardCells,
  hexAdd,
  hexEq,
  hexKey,
  hexNeighbor,
  hexS,
  onBoard,
} from "./coords";

describe("hexKey / hexEq", () => {
  it("formats and compares axial coords", () => {
    expect(hexKey({ q: 1, r: -2 })).toBe("1,-2");
    expect(hexEq({ q: 0, r: 0 }, { q: 0, r: 0 })).toBe(true);
    expect(hexEq({ q: 0, r: 0 }, { q: 1, r: 0 })).toBe(false);
  });
});

describe("hexS / hexAdd / hexNeighbor", () => {
  it("keeps cube constraint s = -q - r", () => {
    expect(hexS({ q: 1, r: -1 })).toBe(0);
    expect(hexS({ q: 2, r: -1 })).toBe(-1);
  });

  it("adds vectors and walks facing neighbors", () => {
    expect(hexAdd({ q: 1, r: 0 }, { q: 0, r: 1 })).toEqual({ q: 1, r: 1 });
    expect(hexNeighbor({ q: 0, r: 0 }, 0)).toEqual(HEX_DIRS[0]);
    expect(hexNeighbor({ q: 0, r: 0 }, 6)).toEqual(HEX_DIRS[0]);
    expect(hexNeighbor({ q: 0, r: 0 }, -1)).toEqual(HEX_DIRS[5]);
  });
});

describe("onBoard / boardCells", () => {
  it("radius-2 board has 19 cells", () => {
    expect(boardCells(2)).toHaveLength(19);
    expect(BOARD_CELLS).toHaveLength(19);
  });

  it("includes origin and excludes cells outside radius", () => {
    expect(onBoard({ q: 0, r: 0 })).toBe(true);
    expect(onBoard({ q: 2, r: 0 })).toBe(true);
    expect(onBoard({ q: 2, r: -2 })).toBe(true);
    expect(onBoard({ q: 3, r: 0 })).toBe(false);
    expect(onBoard({ q: 2, r: 1 })).toBe(false);
  });

  it("every board cell is onBoard and unique", () => {
    const keys = BOARD_CELLS.map(hexKey);
    expect(new Set(keys).size).toBe(keys.length);
    for (const h of BOARD_CELLS) {
      expect(onBoard(h)).toBe(true);
    }
  });
});
