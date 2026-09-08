/** Axial hex coordinates — same q/r fields as the wire protocol. */
export type Hex = { q: number; r: number };

const SQRT3 = Math.sqrt(3);

export function hexKey(h: Hex): string {
  return `${h.q},${h.r}`;
}

export function hexEq(a: Hex, b: Hex): boolean {
  return a.q === b.q && a.r === b.r;
}

/** Cube constraint: s = -q - r */
export function hexS(h: Hex): number {
  return -h.q - h.r;
}

/** Pointy-top neighbor order matches facing 0..5 (E, NE, NW, W, SW, SE). */
export const HEX_DIRS: readonly Hex[] = [
  { q: 1, r: 0 },
  { q: 1, r: -1 },
  { q: 0, r: -1 },
  { q: -1, r: 0 },
  { q: -1, r: 1 },
  { q: 0, r: 1 },
];

export function hexAdd(a: Hex, b: Hex): Hex {
  return { q: a.q + b.q, r: a.r + b.r };
}

export function hexNeighbor(h: Hex, facing: number): Hex {
  return hexAdd(h, HEX_DIRS[((facing % 6) + 6) % 6]!);
}

/** All cells on the radius-2 (19-hex) board. */
export function boardCells(radius = 2): Hex[] {
  const cells: Hex[] = [];
  for (let q = -radius; q <= radius; q++) {
    const r1 = Math.max(-radius, -q - radius);
    const r2 = Math.min(radius, -q + radius);
    for (let r = r1; r <= r2; r++) {
      cells.push({ q, r });
    }
  }
  return cells;
}

export const BOARD_CELLS: readonly Hex[] = boardCells(2);

export function onBoard(h: Hex, radius = 2): boolean {
  return (
    Math.abs(h.q) <= radius &&
    Math.abs(h.r) <= radius &&
    Math.abs(hexS(h)) <= radius
  );
}

export type Pixel = { x: number; y: number };

/** Pointy-top: center of hex in local pixels. */
export function hexToPixel(h: Hex, size: number): Pixel {
  return {
    x: size * (SQRT3 * h.q + (SQRT3 / 2) * h.r),
    y: size * ((3 / 2) * h.r),
  };
}

export function pixelToHex(p: Pixel, size: number): Hex {
  const q = ((SQRT3 / 3) * p.x - (1 / 3) * p.y) / size;
  const r = ((2 / 3) * p.y) / size;
  return axialRound(q, r);
}

export function axialRound(q: number, r: number): Hex {
  const s = -q - r;
  let rq = Math.round(q);
  let rr = Math.round(r);
  const rs = Math.round(s);
  const dq = Math.abs(rq - q);
  const dr = Math.abs(rr - r);
  const ds = Math.abs(rs - s);
  if (dq > dr && dq > ds) rq = -rr - rs;
  else if (dr > ds) rr = -rq - rs;
  return { q: rq, r: rr };
}

/** Six corner points of a pointy-top hex centered at origin. */
export function hexCorners(size: number): Pixel[] {
  const out: Pixel[] = [];
  for (let i = 0; i < 6; i++) {
    const angle = (Math.PI / 180) * (60 * i - 30);
    out.push({ x: size * Math.cos(angle), y: size * Math.sin(angle) });
  }
  return out;
}
