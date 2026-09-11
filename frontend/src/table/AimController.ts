import type { Hex } from "../hex/coords";
import type { HandTile, MatchState } from "../protocol";
import type { Session } from "../session";
import { aimFor, type Aim } from "../rules/instants";
import { tilesAt, transportPassengers } from "../rules/abilities";
import type { InspectTarget } from "./SidePanel";

export type InstantHexResult =
  | { action: "noop" }
  | { action: "sync"; inspect?: InspectTarget }
  | { action: "played" };

export class AimController {
  aim: Aim = null;
  selectedTileId: string | null = null;
  /** When true, next discard-pile click recycles via quartermaster. */
  quartermasterPick = false;

  selectHandTile(tileId: string | null, tile?: HandTile): InspectTarget | null {
    this.quartermasterPick = false;
    if (tileId == null || !tile) {
      this.clear();
      return null;
    }
    if (this.selectedTileId === tileId) {
      this.clear();
      return null;
    }
    this.selectedTileId = tileId;
    this.aim = aimFor(tile);
    return { source: "hand", defId: tile.defId };
  }

  clear(): void {
    this.selectedTileId = null;
    this.aim = null;
    this.quartermasterPick = false;
  }

  startMobility(unitId: string): InspectTarget {
    this.selectedTileId = null;
    this.quartermasterPick = false;
    this.aim = { kind: "mobility", unitId };
    return { source: "board", tileId: unitId };
  }

  aimUnitHex(match: MatchState): Hex | null {
    const id =
      this.aim && "unitId" in this.aim ? this.aim.unitId : undefined;
    if (!id) return null;
    const t = match.board.find((b) => b.id === id);
    return t ? { q: t.q, r: t.r } : null;
  }

  private maybePassenger(
    session: Session,
    via: "move" | "mobility",
    handTileId: string | undefined,
    unitId: string,
    q: number,
    r: number,
    facing: number,
  ): InstantHexResult {
    const match = session.match!;
    const youId = session.youId!;
    const candidates = transportPassengers(match, youId, unitId).map(
      (t) => t.id,
    );
    if (candidates.length === 0) {
      this.commitMoveOrMobility(session, via, handTileId, unitId, q, r, facing);
      return { action: "played" };
    }
    this.aim = {
      kind: "passenger",
      via,
      handTileId,
      unitId,
      q,
      r,
      facing,
      candidates,
    };
    return { action: "sync" };
  }

  commitMoveOrMobility(
    session: Session,
    via: "move" | "mobility",
    handTileId: string | undefined,
    unitId: string,
    q: number,
    r: number,
    facing: number,
    passengerTileId?: string,
  ): void {
    if (via === "move") {
      session.playInstant(handTileId!, {
        targetTileId: unitId,
        q,
        r,
        facing,
        passengerTileId,
      });
    } else {
      session.useMobility(unitId, { q, r, facing, passengerTileId });
    }
    this.clear();
  }

  skipPassenger(session: Session): boolean {
    if (this.aim?.kind !== "passenger") return false;
    const a = this.aim;
    this.commitMoveOrMobility(
      session,
      a.via,
      a.handTileId,
      a.unitId,
      a.q,
      a.r,
      a.facing,
    );
    return true;
  }

  onInstantHex(hex: Hex, session: Session, facing: number): InstantHexResult {
    const match = session.match!;
    const youId = session.youId!;
    const tileId = this.selectedTileId;
    const stack = tilesAt(match, hex.q, hex.r);

    switch (this.aim?.kind) {
      case "move": {
        const moveAim = this.aim;
        if (!moveAim.unitId) {
          const mine = stack.find(
            (t) => t.ownerId === youId && t.kind !== "hq",
          );
          if (!mine) {
            session.setError("Move: click your unit first");
            return { action: "noop" };
          }
          this.aim = { kind: "move", unitId: mine.id };
          return {
            action: "sync",
            inspect: { source: "board", tileId: mine.id },
          };
        }
        const moverId = moveAim.unitId;
        const mover = match.board.find((t) => t.id === moverId);
        if (mover && hex.q === mover.q && hex.r === mover.r) {
          session.playInstant(tileId!, {
            targetTileId: mover.id,
            facing,
          });
          this.clear();
          return { action: "played" };
        }
        return this.maybePassenger(
          session,
          "move",
          tileId!,
          moverId,
          hex.q,
          hex.r,
          facing,
        );
      }
      case "mobility": {
        const mobilityAim = this.aim;
        const mover = match.board.find((t) => t.id === mobilityAim.unitId);
        if (!mover) {
          this.clear();
          return { action: "sync" };
        }
        // Same hex = rotate / stay with current facing.
        if (hex.q === mover.q && hex.r === mover.r) {
          session.useMobility(mover.id, { facing });
          this.clear();
          return { action: "played" };
        }
        return this.maybePassenger(
          session,
          "mobility",
          undefined,
          mover.id,
          hex.q,
          hex.r,
          facing,
        );
      }
      case "passenger": {
        const passengerAim = this.aim;
        const pass = stack.find((t) =>
          passengerAim.candidates.includes(t.id),
        );
        if (!pass) {
          session.setError("Transport: click a highlighted passenger");
          return { action: "noop" };
        }
        this.commitMoveOrMobility(
          session,
          passengerAim.via,
          passengerAim.handTileId,
          passengerAim.unitId,
          passengerAim.q,
          passengerAim.r,
          passengerAim.facing,
          pass.id,
        );
        return { action: "played" };
      }
      case "push": {
        if (!this.aim.unitId) {
          const enemy = stack.find((t) => t.ownerId !== youId);
          if (!enemy) {
            session.setError("Push: click an enemy unit");
            return { action: "noop" };
          }
          this.aim = { kind: "push", unitId: enemy.id };
          return {
            action: "sync",
            inspect: { source: "board", tileId: enemy.id },
          };
        }
        session.playInstant(tileId!, {
          targetTileId: this.aim.unitId,
          q: hex.q,
          r: hex.r,
        });
        this.clear();
        return { action: "played" };
      }
      case "sniper":
      case "grenade": {
        const enemy = stack.find(
          (t) => t.ownerId !== youId && t.kind !== "hq",
        );
        if (!enemy) {
          session.setError("Click an enemy unit (not HQ)");
          return { action: "noop" };
        }
        session.playInstant(tileId!, { targetTileId: enemy.id });
        this.clear();
        return { action: "played" };
      }
      case "airstrike": {
        session.playInstant(tileId!, { q: hex.q, r: hex.r });
        this.clear();
        return { action: "played" };
      }
      default:
        return { action: "noop" };
    }
  }
}
