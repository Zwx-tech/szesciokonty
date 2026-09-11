import { beforeEach, describe, expect, it, vi } from "vitest";
import type { BattleReplay, MatchState, RoomState } from "./protocol";

const store = new Map<string, string>();

vi.stubGlobal("sessionStorage", {
  getItem: (k: string) => store.get(k) ?? null,
  setItem: (k: string, v: string) => {
    store.set(k, v);
  },
  removeItem: (k: string) => {
    store.delete(k);
  },
  clear: () => store.clear(),
});

const { Session } = await import("./session");

function baseMatch(over: Partial<MatchState> = {}): MatchState {
  return {
    v: 1,
    type: "match_state",
    phase: "turn",
    mustDiscard: false,
    unluckyAvailable: false,
    board: [],
    players: [
      {
        id: "p1",
        army: "red",
        hqHp: 20,
        handCount: 0,
        deckCount: 0,
        discardCount: 0,
      },
    ],
    ...over,
  };
}

function lobbyRoom(): RoomState {
  const you = {
    id: "p1",
    name: "Alice",
    ready: false,
    connected: true,
    host: true,
  };
  return {
    v: 1,
    type: "room_state",
    code: "ABCD",
    phase: "lobby",
    you,
    token: "tok",
    seats: [you],
    canStart: false,
  };
}

describe("Session", () => {
  beforeEach(() => {
    store.clear();
  });

  it("receive(match_state) stores match and clears error", () => {
    const session = new Session();
    session.error = "old";
    let n = 0;
    session.subscribe(() => {
      n++;
    });

    session.receive(baseMatch({ phase: "place_hq" }));
    expect(session.match?.phase).toBe("place_hq");
    expect(session.error).toBeNull();
    expect(n).toBe(1);
  });

  it("receive(room_state) in lobby clears match", () => {
    const session = new Session();
    session.match = baseMatch();
    session.receive(lobbyRoom());
    expect(session.room?.code).toBe("ABCD");
    expect(session.match).toBeNull();
    expect(sessionStorage.getItem("roomCode")).toBe("ABCD");
    expect(sessionStorage.getItem("playerToken")).toBe("tok");
  });

  it("consumeReplay returns replay once and clears it", () => {
    const session = new Session();
    const replay: BattleReplay = {
      steps: [
        {
          initiative: 3,
          label: "Init 3",
          board: [],
          hqHp: {},
          log: ["hit"],
        },
      ],
    };
    session.match = baseMatch({ replay });

    const first = session.consumeReplay();
    expect(first?.steps).toHaveLength(1);
    expect(session.match?.replay).toBeUndefined();
    expect(session.consumeReplay()).toBeUndefined();
  });

  it("consumeReplay ignores empty steps", () => {
    const session = new Session();
    session.match = baseMatch({ replay: { steps: [] } });
    expect(session.consumeReplay()).toBeUndefined();
    expect(session.match?.replay?.steps).toEqual([]);
  });

  it("clearError emits only when an error was set", () => {
    const session = new Session();
    let n = 0;
    session.subscribe(() => {
      n++;
    });

    session.clearError();
    expect(n).toBe(0);

    session.setError("boom");
    expect(n).toBe(1);
    expect(session.error).toBe("boom");

    session.clearError();
    expect(session.error).toBeNull();
    expect(n).toBe(2);
  });
});
