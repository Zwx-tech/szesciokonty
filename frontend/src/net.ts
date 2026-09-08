import { type ClientMsg, parseServerMsg, type ServerMsg } from "./protocol";

export type ServerHandler = (msg: ServerMsg) => void;

export class GameSocket {
  private ws: WebSocket | null = null;
  private onMessage: ServerHandler | null = null;

  connect(url = defaultWsURL()): Promise<void> {
    this.close();
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(url);
      this.ws = ws;
      ws.onopen = () => resolve();
      ws.onerror = () => reject(new Error("websocket error"));
      ws.onmessage = (ev) => {
        if (typeof ev.data !== "string" || !this.onMessage) return;
        this.onMessage(parseServerMsg(ev.data));
      };
      ws.onclose = () => {
        this.ws = null;
      };
    });
  }

  on(handler: ServerHandler): void {
    this.onMessage = handler;
  }

  send(msg: ClientMsg): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error("socket not open");
    }
    this.ws.send(JSON.stringify(msg));
  }

  close(): void {
    this.ws?.close();
    this.ws = null;
  }
}

function defaultWsURL(): string {
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${location.host}/ws`;
}
