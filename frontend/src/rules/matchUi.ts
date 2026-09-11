import type { Army, BoardTile, MatchState } from "../protocol";
import { facingLabel } from "../tiles/format";

export function hqDefId(army: Army): string {
  return `${army}_hq`;
}

export function phaseTitle(
  phase: string,
  myTurn: boolean,
  endMode?: string,
): string {
  const turn = myTurn ? "Your turn" : "Opponent's turn";
  switch (phase) {
    case "place_hq":
      return `${turn} · Place HQ`;
    case "turn":
      if (endMode === "tie_break") {
        return `${turn} · Tie-break`;
      }
      return turn;
    case "battle":
      return "Battle";
    case "ended":
      return "Match over";
    default:
      return turn;
  }
}

export function endModeHint(match: MatchState): string {
  switch (match.endMode) {
    case "await_opponent":
      return "Deck empty — finish turns";
    case "final_armed":
      return "Final battle armed";
    case "tie_break":
      return `Tie-break (${match.tieTurnsLeft ?? "?"} turns left)`;
    default:
      return "";
  }
}

export function kindChip(kind: string): string {
  if (kind === "module") return "mod";
  if (kind === "instant") return "act";
  if (kind === "hq") return "hq";
  return "unit";
}

export function boardRuntimeLines(
  tile: BoardTile,
  match: MatchState,
): string[] {
  const lines: string[] = [];
  const owner = match.players.find((p) => p.id === tile.ownerId);
  lines.push(`Owner: ${owner?.army ?? tile.ownerId}`);
  lines.push(`Facing: ${facingLabel(tile.facing)}`);
  if (tile.kind === "hq") {
    lines.push(`HQ HP: ${owner?.hqHp ?? "?"}`);
  } else if (tile.maxHp != null && tile.hp != null) {
    lines.push(`HP: ${tile.hp}/${tile.maxHp} (wounds ${tile.wounds})`);
  } else if (tile.wounds > 0) {
    lines.push(`Wounds: ${tile.wounds}`);
  }
  if (tile.netted) lines.push("Netted");
  if (tile.mobilityAvailable) lines.push("Mobility available");
  if (tile.effectiveInitiatives?.length) {
    lines.push(`Effective init: ${tile.effectiveInitiatives.join(", ")}`);
  }
  return lines;
}
