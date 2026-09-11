import type { TileComponent, TileDef, TileEffect } from "./catalog";

const DIR_LABELS = ["E", "SE", "SW", "W", "NW", "NE"] as const;

export function shortDefName(defId: string): string {
  return defId.replace(/^(red|blue|green|yellow)_/, "");
}

export function facingLabel(facing: number): string {
  const f = ((facing % 6) + 6) % 6;
  return `${f} (${DIR_LABELS[f]})`;
}

export function dirList(dirs: number[] | undefined): string {
  if (!dirs?.length) return "";
  return dirs.map((d) => DIR_LABELS[((d % 6) + 6) % 6]).join("/");
}

function effectLine(e: TileEffect): string {
  const v =
    e.value && e.value !== 0 ? ` ${e.value > 0 ? "+" : ""}${e.value}` : "";
  switch (e.kind) {
    case "init_bonus":
      return `+initiative${v || " +1"}`;
    case "melee_bonus":
      return `+melee${v || " +1"}`;
    case "ranged_bonus":
      return `+ranged${v || " +1"}`;
    case "medic":
      return "medic (absorb one wound)";
    case "mother":
      return "extra attack after last initiative";
    case "saboteur":
      return `saboteur (enemy −initiative${v || " −1"})`;
    case "extra_attack":
      return "HQ extra attack after last initiative";
    case "recon":
      return "recon (peek 3 from enemy deck once/turn)";
    case "scoper":
      return "scoper (ranged/sniper ignore armor)";
    case "transport":
      return "transport (bring passenger into vacated hex)";
    case "quartermaster":
      return "quartermaster (recycle discard → deck bottom)";
    default:
      return e.kind + v;
  }
}

function specialLine(c: TileComponent): string {
  switch (c.id) {
    case "line_wound":
      return `pierce line (str ${c.params?.strength ?? 1})`;
    case "detonate":
      return "detonate on death (wound all adjacent)";
    case "blocker":
      return "blocker (immune to Move/Push)";
    case "dual_stack":
      return "dual stack (place on friendly warrior)";
    default:
      return c.id ? `special: ${c.id}` : "special";
  }
}

export function describeComponent(c: TileComponent): string | null {
  switch (c.type) {
    case "attack": {
      const kind = c.attackType === "ranged" ? "Ranged" : "Melee";
      const dirs = dirList(c.dirs);
      return `${kind} ${c.strength ?? 1}${dirs ? ` → ${dirs}` : ""}`;
    }
    case "armor": {
      const dirs = dirList(c.dirs);
      return `Armor${dirs ? ` on ${dirs}` : ""}`;
    }
    case "net": {
      const dirs = dirList(c.dirs);
      return `Net${dirs ? ` → ${dirs}` : ""}`;
    }
    case "mobility":
      return "Mobility (move/rotate once per turn)";
    case "module_link": {
      const dirs = dirList(c.dirs);
      return `Links${dirs ? ` → ${dirs}` : ""}`;
    }
    case "module_aura":
    case "hq_aura": {
      const effects = (c.effects ?? []).map(effectLine).join(", ");
      return effects || "Aura";
    }
    case "instant":
      return `Instant: ${c.instant ?? "?"}`;
    case "special":
      return specialLine(c);
    default:
      return c.type || null;
  }
}

export function describeDef(def: TileDef): string[] {
  const lines: string[] = [];
  lines.push(`Kind: ${def.kind}`);
  if (def.initiatives?.length) {
    lines.push(`Initiative: ${def.initiatives.join(", ")}`);
  }
  if (def.kind !== "hq") {
    const maxHp = 1 + (def.toughness ?? 0);
    lines.push(`Toughness: ${def.toughness ?? 0} (HP ${maxHp})`);
  } else {
    lines.push("HQ (separate HP track)");
  }
  for (const c of def.components ?? []) {
    const line = describeComponent(c);
    if (line) lines.push(line);
  }
  return lines;
}
