import type {
  BattleReplay,
  BattleStep,
  MatchState,
  RoomState,
} from "../protocol";
import { prefetchTileDefs } from "../tiles/catalog";
import type { HexBoard } from "../hex/HexBoard";

export type BattlePlaybackDeps = {
  getMatch: () => MatchState | null | undefined;
  getRoom: () => RoomState | null | undefined;
  getYouId: () => string | null | undefined;
  getBoard: () => HexBoard | null;
  onBannerStatus: (banner: string, status: string) => void;
  onHqStrips: (oppText: string, myText: string) => void;
  onEventLog: (lines: string[]) => void;
  onLockControls: (locked: boolean) => void;
  onSkipVisible: (visible: boolean) => void;
  onPlaybackClass: (active: boolean) => void;
  onFinished: () => void;
};

export class BattlePlayback {
  private _playing = false;
  private steps: BattleStep[] = [];
  private timer: ReturnType<typeof setTimeout> | null = null;
  private log: string[] = [];
  private static readonly PLAYBACK_MS = 900;
  private readonly deps: BattlePlaybackDeps;

  constructor(deps: BattlePlaybackDeps) {
    this.deps = deps;
  }

  get playing(): boolean {
    return this._playing;
  }

  start(replay: BattleReplay): void {
    const match = this.deps.getMatch();
    if (!match || !replay.steps.length) {
      this.deps.onFinished();
      return;
    }

    this.stop(false);
    this._playing = true;
    this.steps = replay.steps;
    this.log = ["Battle begins"];
    this.deps.onPlaybackClass(true);
    this.deps.onSkipVisible(true);
    this.deps.onLockControls(true);
    this.deps.onEventLog(this.log);
    this.showStep(0);
  }

  skip(): void {
    if (!this._playing) return;
    this.finish();
  }

  stop(runFinished: boolean): void {
    if (this.timer != null) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.steps = [];
    this._playing = false;
    if (runFinished) this.deps.onFinished();
  }

  private finish(): void {
    this.stop(false);
    this._playing = false;
    this.deps.onPlaybackClass(false);
    this.deps.onSkipVisible(false);
    this.deps.onFinished();
  }

  private showStep(index: number): void {
    if (!this._playing) return;
    if (index >= this.steps.length) {
      this.finish();
      return;
    }
    const step = this.steps[index]!;
    const match = this.deps.getMatch();

    this.deps.onBannerStatus(
      `Battle · ${step.label}`,
      `Playing initiative phases… (${index + 1}/${this.steps.length})`,
    );

    this.log = [...this.log, ...step.log];
    this.deps.onEventLog(this.log);

    if (match) {
      this.applyBoard(step, match);
      this.applyHqStrips(step, match);
    }

    this.timer = setTimeout(() => {
      this.timer = null;
      this.showStep(index + 1);
    }, BattlePlayback.PLAYBACK_MS);
  }

  private applyBoard(step: BattleStep, match: MatchState): void {
    const board = this.deps.getBoard();
    if (!board) return;
    const hqHp = step.hqHp;
    board.setOccupants(
      step.board.map((t) => ({
        id: t.id,
        q: t.q,
        r: t.r,
        defId: t.defId,
        facing: t.facing,
        wounds: t.wounds,
        hp: t.hp,
        maxHp: t.maxHp,
        netted: t.netted,
        isHQ: t.kind === "hq",
        hqHp: t.kind === "hq" ? hqHp[t.ownerId] : undefined,
      })),
    );
    board.clearLegal();
    board.setPlacementPreview(null);
    board.setSelected(null);
    prefetchTileDefs([
      ...step.board.map((t) => t.defId),
      ...match.board.map((t) => t.defId),
    ]);
  }

  private applyHqStrips(step: BattleStep, match: MatchState): void {
    const room = this.deps.getRoom();
    const youId = this.deps.getYouId();
    const seatName = (id: string) =>
      room?.seats.find((s) => s.id === id)?.name ?? "Player";
    const me = match.players.find((p) => p.id === youId);
    const opp = match.players.find((p) => p.id !== youId);
    let oppText = "";
    let myText = "";
    if (opp) {
      const hq = step.hqHp[opp.id] ?? opp.hqHp;
      oppText = `${seatName(opp.id)} · ${opp.army} · HQ ${hq} · deck ${opp.deckCount} · discard ${opp.discardCount} · hand ${opp.handCount}`;
    }
    if (me) {
      const hq = step.hqHp[me.id] ?? me.hqHp;
      myText = `You (${seatName(me.id)}) · ${me.army} · HQ ${hq} · deck ${me.deckCount} · discard ${me.discardCount}`;
    }
    this.deps.onHqStrips(oppText, myText);
  }
}
