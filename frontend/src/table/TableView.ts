import { el } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import type { Hex } from "../hex/coords";
import type { HandTile, MatchState } from "../protocol";
import { prefetchTileDefs } from "../tiles/catalog";
import { facingLabel } from "../tiles/format";
import { tilesAt } from "../rules/abilities";
import { aimHint, aimLegalHexes, instantKind } from "../rules/instants";
import { endModeHint, hqDefId, phaseTitle } from "../rules/matchUi";
import { AimController } from "./AimController";
import { ActionDock } from "./ActionDock";
import { BattlePlayback } from "./BattlePlayback";
import { BoardHost } from "./BoardHost";
import { HandPanel } from "./HandPanel";
import { SidePanel } from "./SidePanel";

export class TableView extends View {
  private banner: HTMLElement;
  private status: HTMLElement;
  private errorLabel: HTMLElement;
  private oppInfo: HTMLElement;
  private myInfo: HTMLElement;
  private midRow: HTMLElement;

  private readonly boardHost: BoardHost;
  private readonly side: SidePanel;
  private readonly hand: HandPanel;
  private readonly dock: ActionDock;
  private readonly aimCtrl = new AimController();
  private readonly playback: BattlePlayback;

  private facing = 0;
  private stackCycle = 0;

  constructor(router: ViewRouter) {
    super(router);
    this.root.classList.add("table-view");

    const top = el("div", { className: "table-top" });
    const left = el("div", { className: "stack" });
    this.banner = el("h1", { text: "" });
    this.status = el("div", { className: "muted" });
    this.errorLabel = el("div", { className: "error" });
    left.append(this.banner, this.status, this.errorLabel);
    const leaveBtn = el("button", { text: "Leave" });
    leaveBtn.addEventListener("click", () => {
      this.session.leaveRoom();
      this.router.show("hub");
    });
    top.append(left, leaveBtn);
    this.root.append(top);

    const oppZone = el("div", { className: "player-zone" });
    this.oppInfo = el("div", { className: "muted player-strip" });
    oppZone.append(this.oppInfo);
    this.root.append(oppZone);

    this.midRow = el("div", { className: "table-mid" });
    this.boardHost = new BoardHost();
    this.side = new SidePanel();
    this.side.onCollapseChange = (collapsed) => {
      this.midRow.classList.toggle("side-collapsed", collapsed);
      requestAnimationFrame(() => this.boardHost.layout());
    };
    this.side.onDiscardPick = (tileId) => {
      this.session.useQuartermaster(tileId);
      this.aimCtrl.quartermasterPick = false;
      this.sync();
    };
    this.midRow.append(this.boardHost.element, this.side.element);
    this.root.append(this.midRow);

    const youZone = el("div", { className: "player-zone" });
    this.myInfo = el("div", { className: "muted player-strip" });
    this.hand = new HandPanel();
    this.hand.onSelect = (tileId, tile) => {
      const inspect = this.aimCtrl.selectHandTile(tileId, tile);
      this.side.setInspect(inspect);
      this.sync();
    };
    const myShelf = el("div", { className: "hand-shelf" });
    myShelf.append(this.hand.element);

    this.dock = new ActionDock();
    this.dock.onRotate = (delta) => this.rotate(delta);
    this.dock.onDiscard = () => this.doDiscard();
    this.dock.onPlay = () => this.doPlayBattle();
    this.dock.onUnlucky = () => this.session.redrawUnlucky();
    this.dock.onEndTurn = () => this.session.endTurn();
    this.dock.onSkip = () => this.playback.skip();
    this.dock.onRecon = () => this.session.useRecon();
    this.dock.onQuartermaster = () => {
      this.aimCtrl.quartermasterPick = true;
      this.sync();
    };
    this.dock.onSkipPassenger = () => {
      if (this.aimCtrl.skipPassenger(this.session)) this.sync();
    };

    const bottom = el("div", { className: "table-bottom" });
    bottom.append(myShelf, this.dock.element);
    youZone.append(this.myInfo, bottom);
    this.root.append(youZone);

    this.playback = new BattlePlayback({
      getMatch: () => this.session.match,
      getRoom: () => this.session.room,
      getYouId: () => this.session.youId,
      getBoard: () => this.boardHost.board,
      onBannerStatus: (banner, status) => {
        this.banner.textContent = banner;
        this.status.textContent = status;
      },
      onHqStrips: (oppText, myText) => {
        this.oppInfo.textContent = oppText;
        this.myInfo.textContent = myText;
      },
      onEventLog: (lines) => this.side.renderEventLog(lines),
      onLockControls: (locked) => this.dock.setLocked(locked),
      onSkipVisible: (visible) => this.dock.setSkipVisible(visible),
      onPlaybackClass: (active) => {
        this.root.classList.toggle("battle-playback", active);
      },
      onFinished: () => this.sync(),
    });
  }

  onEnter(): void {
    this.aimCtrl.clear();
    this.facing = 0;
    this.side.clearInspect();
    this.playback.stop(false);
    void this.boardHost
      .mount({
        onPick: (hex) => this.onHex(hex),
        onHover: (hex) => this.onBoardHover(hex),
        onRotate: (delta) => this.rotate(delta),
      })
      .then(() => {
        this.sync();
        this.boardHost.layout();
      });
  }

  onLeave(): void {
    this.playback.stop(false);
    this.boardHost.destroy();
  }

  onSessionUpdate(): void {
    const me = this.session.match?.players.find(
      (p) => p.id === this.session.youId,
    );
    if (
      this.aimCtrl.selectedTileId &&
      !me?.hand?.some((t) => t.id === this.aimCtrl.selectedTileId)
    ) {
      this.aimCtrl.clear();
    }

    const replay = this.session.consumeReplay();
    if (replay && !this.playback.playing) {
      this.playback.start(replay);
      return;
    }
    if (this.playback.playing) {
      // Keep playing; final match is already stored on session.
      return;
    }
    this.sync();
  }

  onResize(): void {
    this.boardHost.layout();
  }

  private rotate(delta: number): void {
    if (this.playback.playing) return;
    this.facing = (((this.facing + delta) % 6) + 6) % 6;
    this.dock.setFacingLabel(facingLabel(this.facing));
    this.status.textContent = (this.status.textContent ?? "").replace(
      /Facing \d+ \([^)]+\)/,
      facingLabel(this.facing),
    );
    this.boardHost.board?.setPlacementPreview(this.currentPlacementPreview());
  }

  private currentPlacementPreview(): {
    defId: string;
    facing: number;
  } | null {
    const match = this.session.match;
    const youId = this.session.youId;
    if (!match || !youId || match.turnPlayerId !== youId) return null;
    if (match.phase === "battle" || match.phase === "ended") return null;

    const me = match.players.find((p) => p.id === youId);
    if (!me) return null;

    if (match.phase === "place_hq") {
      return { defId: hqDefId(me.army), facing: this.facing };
    }
    if (match.phase !== "turn" || match.mustDiscard) return null;

    const selected = me.hand?.find((t) => t.id === this.aimCtrl.selectedTileId);
    if (
      selected &&
      (selected.kind === "warrior" || selected.kind === "module")
    ) {
      return { defId: selected.defId, facing: this.facing };
    }
    return null;
  }

  private onBoardHover(hex: Hex | null): void {
    const match = this.session.match;
    if (!hex || !match) return;
    const stack = tilesAt(match, hex.q, hex.r);
    const tile = stack[0];
    if (tile) {
      this.side.setInspect({ source: "board", tileId: tile.id });
      void this.side.refreshInspect(match);
    }
  }

  private sync(): void {
    const match = this.session.match;
    const room = this.session.room;
    const youId = this.session.youId;
    this.errorLabel.textContent = this.session.error ?? "";

    if (!match || !youId) {
      this.banner.textContent = "Waiting…";
      this.status.textContent = "";
      return;
    }

    if (this.playback.playing) {
      this.dock.setLocked(true);
      this.dock.setSkipVisible(true);
      this.dock.setSkipPassengerVisible(false);
      return;
    }
    this.dock.setSkipVisible(false);
    this.root.classList.remove("battle-playback");

    const me = match.players.find((p) => p.id === youId);
    const opp = match.players.find((p) => p.id !== youId);
    const myTurn = match.turnPlayerId === youId;
    const selected = me?.hand?.find(
      (t) => t.id === this.aimCtrl.selectedTileId,
    );
    const seatName = (id: string) =>
      room?.seats.find((s) => s.id === id)?.name ?? "Player";

    this.root.classList.toggle("must-discard", !!match.mustDiscard && myTurn);

    this.banner.textContent = phaseTitle(match.phase, myTurn, match.endMode);
    this.status.textContent = [
      room ? `Room ${room.code}` : "",
      match.mustDiscard ? "Discard one tile" : "",
      endModeHint(match),
      aimHint(this.aimCtrl.aim),
      this.aimCtrl.quartermasterPick ? "Recycle: pick a discard tile" : "",
      facingLabel(this.facing),
    ]
      .filter(Boolean)
      .join(" · ");
    this.dock.setFacingLabel(facingLabel(this.facing));

    this.oppInfo.textContent = opp
      ? `${seatName(opp.id)} · ${opp.army} · HQ ${opp.hqHp} · deck ${opp.deckCount} · discard ${opp.discardCount} · hand ${opp.handCount}`
      : "";
    this.myInfo.textContent = me
      ? `You (${seatName(me.id)}) · ${me.army} · HQ ${me.hqHp} · deck ${me.deckCount} · discard ${me.discardCount}`
      : "";

    this.hand.setHand(me?.hand ?? [], {
      interactive: myTurn,
      selectedId: this.aimCtrl.selectedTileId,
    });

    prefetchTileDefs([
      ...(me?.hand ?? []).map((t) => t.defId),
      ...(me?.discard ?? []).map((t) => t.defId),
      ...(match.reconPeek ?? []),
      ...match.board.map((t) => t.defId),
      me ? hqDefId(me.army) : "",
    ]);

    this.side.setReconPeek(match.reconPeek ?? []);
    this.side.setDiscard(me?.discard ?? [], this.aimCtrl.quartermasterPick);

    const board = this.boardHost.board;
    if (board) {
      const hqHpByOwner = new Map<string, number>();
      for (const p of match.players) hqHpByOwner.set(p.id, p.hqHp);
      board.setOccupants(
        match.board.map((t) => ({
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
          hqHp: t.kind === "hq" ? hqHpByOwner.get(t.ownerId) : undefined,
        })),
      );

      const canAct =
        myTurn && match.phase !== "battle" && match.phase !== "ended";
      this.updateBoardLegal(canAct, match.phase, selected, match, youId);
      board.setPlacementPreview(this.currentPlacementPreview());
      board.setSelected(this.aimCtrl.aimUnitHex(match));
    }

    this.side.renderEventLog(match.log ?? []);
    void this.side.refreshInspect(match);

    const canAct =
      myTurn && match.phase !== "battle" && match.phase !== "ended";
    const passengerAim = this.aimCtrl.aim?.kind === "passenger";
    this.dock.setSkipPassengerVisible(passengerAim);
    this.dock.setDisabledStates({
      discard: !(
        myTurn &&
        match.phase === "turn" &&
        !!this.aimCtrl.selectedTileId
      ),
      play: !(
        myTurn &&
        match.phase === "turn" &&
        !match.mustDiscard &&
        selected?.kind === "instant" &&
        instantKind(selected.defId) === "battle"
      ),
      end: !(myTurn && match.phase === "turn" && !match.mustDiscard),
      unlucky: !(myTurn && match.unluckyAvailable),
      rotLeft: !canAct,
      rotRight: !canAct,
      recon: !(myTurn && !!match.reconAvailable),
      quartermaster: !(myTurn && !!match.quartermasterAvailable),
      skipPassenger: !passengerAim,
    });
    this.dock.setUnluckyTitle(
      match.unluckyAvailable
        ? "Redraw this turn’s draw once"
        : "Unlucky redraw not available",
    );

    if (selected) {
      this.side.setInspect({ source: "hand", defId: selected.defId });
    }
  }

  private updateBoardLegal(
    canAct: boolean,
    phase: string,
    selected: HandTile | undefined,
    match: MatchState,
    youId: string,
  ): void {
    const board = this.boardHost.board;
    if (!board) return;
    if (!canAct) {
      board.clearLegal();
      return;
    }
    if (phase === "place_hq") {
      board.setLegal(match.legalHexes ?? []);
      return;
    }
    if (phase === "turn" && match.mustDiscard) {
      board.clearLegal();
      return;
    }
    if (phase === "turn") {
      if (
        selected &&
        (selected.kind === "warrior" || selected.kind === "module")
      ) {
        board.setLegal(match.legalHexes ?? []);
        return;
      }
      if (this.aimCtrl.aim) {
        board.setLegal(aimLegalHexes(this.aimCtrl.aim, match, youId));
        return;
      }
      // Mobility units are clickable via their hexes.
      const mobilityHexes = match.board
        .filter((t) => t.ownerId === youId && t.mobilityAvailable)
        .map((t) => ({ q: t.q, r: t.r }));
      if (mobilityHexes.length) {
        board.setLegal(mobilityHexes);
        return;
      }
    }
    board.clearLegal();
  }

  private onHex(hex: Hex): void {
    if (this.playback.playing) return;
    const match = this.session.match;
    const youId = this.session.youId;
    if (!match || !youId || match.turnPlayerId !== youId) return;

    const stack = tilesAt(match, hex.q, hex.r);
    const occupied = stack[0];

    if (occupied && !this.aimCtrl.aim) {
      this.stackCycle = (this.stackCycle + 1) % stack.length;
      const pick = stack[this.stackCycle % stack.length]!;
      this.side.setInspect({ source: "board", tileId: pick.id });
      void this.side.refreshInspect(match);

      // Start mobility on own available unit when nothing selected from hand.
      if (
        !this.aimCtrl.selectedTileId &&
        match.phase === "turn" &&
        !match.mustDiscard
      ) {
        const mobile = stack.find(
          (t) => t.ownerId === youId && t.mobilityAvailable,
        );
        if (mobile) {
          const inspect = this.aimCtrl.startMobility(mobile.id);
          this.side.setInspect(inspect);
          this.sync();
          return;
        }
      }

      if (this.aimCtrl.selectedTileId) {
        const me = match.players.find((p) => p.id === youId);
        const tile = me?.hand?.find(
          (t) => t.id === this.aimCtrl.selectedTileId,
        );
        if (tile && (tile.kind === "warrior" || tile.kind === "module")) {
          const legal = (match.legalHexes ?? []).some(
            (h) => h.q === hex.q && h.r === hex.r,
          );
          if (!legal) return; // dual-stack legal hexes still allow place
        }
      }
    }

    if (match.phase === "place_hq") {
      this.session.place("", hex.q, hex.r, this.facing);
      return;
    }
    if (match.phase !== "turn" || match.mustDiscard) return;

    if (this.aimCtrl.aim) {
      const result = this.aimCtrl.onInstantHex(hex, this.session, this.facing);
      if (result.action === "sync") {
        if (result.inspect) this.side.setInspect(result.inspect);
        this.sync();
      }
      return;
    }

    if (!this.aimCtrl.selectedTileId) {
      if (!occupied) this.session.setError("Select a tile first");
      return;
    }
    const me = match.players.find((p) => p.id === youId);
    const tile = me?.hand?.find((t) => t.id === this.aimCtrl.selectedTileId);
    if (!tile || (tile.kind !== "warrior" && tile.kind !== "module")) {
      this.session.setError("Select a unit, or target with an instant");
      return;
    }
    this.session.place(this.aimCtrl.selectedTileId, hex.q, hex.r, this.facing);
    this.aimCtrl.clear();
  }

  private doPlayBattle(): void {
    if (!this.aimCtrl.selectedTileId) return;
    this.session.playInstant(this.aimCtrl.selectedTileId);
    this.aimCtrl.clear();
  }

  private doDiscard(): void {
    if (!this.aimCtrl.selectedTileId) {
      this.session.setError("Select a tile to discard");
      return;
    }
    this.session.discard(this.aimCtrl.selectedTileId);
    this.aimCtrl.clear();
  }
}
