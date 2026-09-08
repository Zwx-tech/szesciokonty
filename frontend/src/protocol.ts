export const ProtocolVersion = 1 as const;

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
  | { v: 1; type: "place"; tileId: string; q: number; r: number; facing: number }
  | {
      v: 1;
      type: "play_instant";
      tileId: string;
      q?: number;
      r?: number;
      facing?: number;
      targetTileId?: string;
    }
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

export type Hex = { q: number; r: number };

export type BoardTile = {
  id: string;
  defId: string;
  ownerId: string;
  q: number;
  r: number;
  facing: number;
  wounds: number;
};

export type HandTile = { id: string; defId: string };

export type PlayerView = {
  id: string;
  army: Army;
  hqHp: number;
  hand: HandTile[];
  deckCount: number;
  discardCount: number;
};

export type MatchResult = {
  winnerId?: string;
  draw?: boolean;
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
    throw new Error(`unsupported version ${String((msg as { v?: unknown }).v)}`);
  }
  return msg;
}
