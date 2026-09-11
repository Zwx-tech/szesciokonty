export const ProtocolVersion = 1 as const;

import type { Hex } from "./hex/coords";
export type { Hex };

export type Army = "red" | "blue" | "green" | "yellow";

export type RoomPhase = "lobby" | "match";
export type MatchPhase = "place_hq" | "turn" | "battle" | "ended";

export type ClientMsg =
  | { v: 1; type: "create_room"; name: string }
  | { v: 1; type: "join_room"; name: string; code: string }
  | { v: 1; type: "leave" }
  | { v: 1; type: "set_army"; army: Army }
  | { v: 1; type: "ready"; ready: boolean }
  | { v: 1; type: "start" }
  | { v: 1; type: "discard"; tileId: string }
  | {
      v: 1;
      type: "place";
      tileId: string;
      q: number;
      r: number;
      facing: number;
    }
  | {
      v: 1;
      type: "play_instant";
      tileId: string;
      q?: number;
      r?: number;
      facing?: number;
      targetTileId?: string;
      passengerTileId?: string;
    }
  | {
      v: 1;
      type: "use_mobility";
      tileId: string;
      q?: number;
      r?: number;
      facing?: number;
      passengerTileId?: string;
    }
  | { v: 1; type: "use_recon" }
  | { v: 1; type: "use_quartermaster"; tileId: string }
  | { v: 1; type: "end_turn" }
  | { v: 1; type: "redraw_unlucky" }
  | { v: 1; type: "rematch" }
  | { v: 1; type: "reconnect"; code: string; playerToken: string };

export type Seat = {
  id: string;
  name: string;
  army?: Army;
  ready: boolean;
  connected: boolean;
  host: boolean;
};

export type BoardTile = {
  id: string;
  defId: string;
  kind: string;
  ownerId: string;
  q: number;
  r: number;
  facing: number;
  wounds: number;
  hp?: number;
  maxHp?: number;
  netted?: boolean;
  effectiveInitiatives?: number[];
  edges?: EdgeMark[];
  mobilityAvailable?: boolean;
};

export type EdgeMark = {
  dir: number;
  kind: "melee" | "ranged" | "net" | "armor" | "module_link" | string;
  strength?: number;
};

export type HandTile = { id: string; defId: string; kind: string };

export type PlayerView = {
  id: string;
  army: Army;
  hqHp: number;
  hand?: HandTile[];
  handCount: number;
  deckCount: number;
  discardCount: number;
  discard?: HandTile[];
};

export type MatchResult = {
  winnerId?: string;
  draw?: boolean;
};

export type BattleStep = {
  initiative: number;
  label: string;
  board: BoardTile[];
  hqHp: Record<string, number>;
  log: string[];
};

export type BattleReplay = {
  steps: BattleStep[];
};

export type RoomState = {
  v: 1;
  type: "room_state";
  code: string;
  phase: RoomPhase;
  you: Seat;
  token: string;
  seats: Seat[];
  canStart: boolean;
};

export type MatchState = {
  v: 1;
  type: "match_state";
  phase: MatchPhase;
  turnPlayerId?: string;
  mustDiscard: boolean;
  unluckyAvailable: boolean;
  board: BoardTile[];
  players: PlayerView[];
  legalHexes?: Hex[];
  result?: MatchResult;
  log?: string[];
  endMode?: string;
  tieTurnsLeft?: number;
  replay?: BattleReplay;
  reconPeek?: string[];
  reconAvailable?: boolean;
  quartermasterAvailable?: boolean;
};

export type ErrorMsg = {
  v: 1;
  type: "error";
  code: string;
  message: string;
};

export type ServerMsg = RoomState | MatchState | ErrorMsg;

export function parseServerMsg(data: string): ServerMsg {
  const msg = JSON.parse(data) as ServerMsg;
  if (msg.v !== ProtocolVersion) {
    throw new Error(
      `unsupported version ${String((msg as { v?: unknown }).v)}`,
    );
  }
  return msg;
}
