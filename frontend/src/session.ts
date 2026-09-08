import { GameSocket } from "./net";
import type { Army, ErrorMsg, RoomState, ServerMsg } from "./protocol";

const KEY_NAME = "playerName";
const KEY_CODE = "roomCode";
const KEY_TOKEN = "playerToken";

type Listener = () => void;

export class Session {
  readonly socket = new GameSocket();
  name = sessionStorage.getItem(KEY_NAME) ?? "";
  room: RoomState | null = null;
  error: string | null = null;
  private listeners = new Set<Listener>();

  constructor() {
    this.socket.on((msg) => this.handle(msg));
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

  private handle(msg: ServerMsg): void {
    if (msg.type === "room_state") {
      this.room = msg;
      this.error = null;
      sessionStorage.setItem(KEY_CODE, msg.code);
      sessionStorage.setItem(KEY_TOKEN, msg.token);
      this.emit();
      return;
    }
    if (msg.type === "error") {
      this.onError(msg);
      return;
    }
    // match_state arrives in later tasks
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
    sessionStorage.removeItem(KEY_CODE);
    sessionStorage.removeItem(KEY_TOKEN);
  }

  clearError(): void {
    this.error = null;
  }

  private emit(): void {
    for (const fn of this.listeners) fn();
  }
}
