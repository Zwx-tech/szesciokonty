import { GameSocket } from "./net";
import type {
  Army,
  BattleReplay,
  ErrorMsg,
  MatchState,
  RoomState,
  ServerMsg,
} from "./protocol";

const KEY_NAME = "playerName";
const KEY_CODE = "roomCode";
const KEY_TOKEN = "playerToken";

type Listener = () => void;

export class Session {
  readonly socket = new GameSocket();
  name = sessionStorage.getItem(KEY_NAME) ?? "";
  room: RoomState | null = null;
  match: MatchState | null = null;
  error: string | null = null;
  private listeners = new Set<Listener>();

  constructor() {
    this.socket.on((msg) => this.handle(msg));
  }

  get youId(): string | null {
    return this.room?.you.id ?? null;
  }

  subscribe(fn: Listener): () => void {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  setName(name: string): void {
    this.name = name.trim();
    sessionStorage.setItem(KEY_NAME, this.name);
  }

  async ensureConnected(): Promise<void> {
    if (this.socket.isOpen) return;
    await this.socket.connect();
  }

  async tryReconnect(): Promise<boolean> {
    const code = sessionStorage.getItem(KEY_CODE);
    const token = sessionStorage.getItem(KEY_TOKEN);
    if (!code || !token || !this.name) return false;
    await this.ensureConnected();
    this.socket.send({ v: 1, type: "reconnect", code, playerToken: token });
    return true;
  }

  createRoom(): void {
    this.clearError();
    this.socket.send({ v: 1, type: "create_room", name: this.name });
  }

  joinRoom(code: string): void {
    this.clearError();
    this.socket.send({
      v: 1,
      type: "join_room",
      name: this.name,
      code: code.trim().toUpperCase(),
    });
  }

  leaveRoom(): void {
    this.clearError();
    if (this.socket.isOpen) {
      this.socket.send({ v: 1, type: "leave" });
    }
    this.clearRoom();
    this.emit();
  }

  setError(message: string): void {
    this.error = message;
    this.emit();
  }

  setArmy(army: Army): void {
    this.clearError();
    this.socket.send({ v: 1, type: "set_army", army });
  }

  setReady(ready: boolean): void {
    this.clearError();
    this.socket.send({ v: 1, type: "ready", ready });
  }

  start(): void {
    this.clearError();
    this.socket.send({ v: 1, type: "start" });
  }

  place(tileId: string, q: number, r: number, facing: number): void {
    this.clearError();
    this.socket.send({ v: 1, type: "place", tileId, q, r, facing });
  }

  discard(tileId: string): void {
    this.clearError();
    this.socket.send({ v: 1, type: "discard", tileId });
  }

  endTurn(): void {
    this.clearError();
    this.socket.send({ v: 1, type: "end_turn" });
  }

  redrawUnlucky(): void {
    this.clearError();
    this.socket.send({ v: 1, type: "redraw_unlucky" });
  }

  playInstant(
    tileId: string,
    opts?: {
      q?: number;
      r?: number;
      facing?: number;
      targetTileId?: string;
      passengerTileId?: string;
    },
  ): void {
    this.clearError();
    this.socket.send({
      v: 1,
      type: "play_instant",
      tileId,
      q: opts?.q,
      r: opts?.r,
      facing: opts?.facing,
      targetTileId: opts?.targetTileId,
      passengerTileId: opts?.passengerTileId,
    });
  }

  useMobility(
    tileId: string,
    opts?: {
      q?: number;
      r?: number;
      facing?: number;
      passengerTileId?: string;
    },
  ): void {
    this.clearError();
    this.socket.send({
      v: 1,
      type: "use_mobility",
      tileId,
      q: opts?.q,
      r: opts?.r,
      facing: opts?.facing,
      passengerTileId: opts?.passengerTileId,
    });
  }

  useRecon(): void {
    this.clearError();
    this.socket.send({ v: 1, type: "use_recon" });
  }

  useQuartermaster(tileId: string): void {
    this.clearError();
    this.socket.send({ v: 1, type: "use_quartermaster", tileId });
  }

  rematch(): void {
    this.clearError();
    this.socket.send({ v: 1, type: "rematch" });
  }

  /**
   * Take battle replay from the current match once.
   * Clears `match.replay` so the same snapshot cannot start playback again.
   */
  consumeReplay(): BattleReplay | undefined {
    const replay = this.match?.replay;
    if (!this.match || !replay?.steps?.length) return undefined;
    this.match = { ...this.match, replay: undefined };
    return replay;
  }

  /** Apply a server message (used by the socket; also handy in tests). */
  receive(msg: ServerMsg): void {
    this.handle(msg);
  }

  private handle(msg: ServerMsg): void {
    if (msg.type === "room_state") {
      this.room = msg;
      this.error = null;
      sessionStorage.setItem(KEY_CODE, msg.code);
      sessionStorage.setItem(KEY_TOKEN, msg.token);
      if (msg.phase === "lobby") this.match = null;
      this.emit();
      return;
    }
    if (msg.type === "match_state") {
      this.match = msg;
      this.error = null;
      this.emit();
      return;
    }
    if (msg.type === "error") {
      this.onError(msg);
      return;
    }
    this.emit();
  }

  private onError(msg: ErrorMsg): void {
    if (msg.code === "room_closed") {
      this.clearRoom();
      this.error = msg.message;
      this.emit();
      return;
    }
    this.error = msg.message;
    this.emit();
  }

  clearRoom(): void {
    this.room = null;
    this.match = null;
    sessionStorage.removeItem(KEY_CODE);
    sessionStorage.removeItem(KEY_TOKEN);
  }

  clearError(): void {
    if (this.error == null) return;
    this.error = null;
    this.emit();
  }

  private emit(): void {
    for (const fn of this.listeners) fn();
  }
}
