import { Application } from "pixi.js";
import { el, clear, setDisabled } from "../dom/dom";
import { View } from "../app/View";
import type { ViewRouter } from "../app/ViewRouter";
import type { Hex } from "../hex/coords";
import { HexBoard } from "../hex/HexBoard";
import type { HandTile } from "../protocol";
import { tileSvgUrl } from "../hex/tileAssets";
import { colors } from "../ui/theme";

type InstantKind =
  "battle" | "move" | "push" | "sniper" | "airstrike" | "grenade";

type Aim =
  | null
  | { kind: "battle" }
  | { kind: "move"; unitId?: string }
  | { kind: "push"; unitId?: string }
  | { kind: "sniper" }
  | { kind: "airstrike" }
  | { kind: "grenade" };

export class TableView extends View {
  private banner: HTMLElement;
  private status: HTMLElement;
  private errorLabel: HTMLElement;
  private oppInfo: HTMLElement;
  private myInfo: HTMLElement;
  private oppHand: HTMLElement;
  private myHand: HTMLElement;
  private boardHost: HTMLElement;
  private facingLabel: HTMLElement;
  private discardBtn: HTMLButtonElement;
  private playBtn: HTMLButtonElement;
  private endBtn: HTMLButtonElement;
  private unluckyBtn: HTMLButtonElement;
  private rotLeft: HTMLButtonElement;
  private rotRight: HTMLButtonElement;

  private pixi: Application | null = null;
  private board: HexBoard | null = null;
  private resizeObserver: ResizeObserver | null = null;

  private selectedTileId: string | null = null;
  private aim: Aim = null;
  private facing = 0;

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

    this.oppInfo = el("div", { className: "muted" });
    this.oppHand = el("div", { className: "hand" });
    this.root.append(this.oppInfo, this.oppHand);

    this.boardHost = el("div", { attrs: { id: "board-host" } });
    this.root.append(this.boardHost);

    this.myInfo = el("div", { className: "muted" });
    this.myHand = el("div", { className: "hand" });
    this.root.append(this.myInfo, this.myHand);

    const controls = el("div", { className: "controls" });
    this.rotLeft = el("button", { text: "◀" });
    this.rotLeft.addEventListener("click", () => this.rotate(-1));
    this.facingLabel = el("span", { className: "facing", text: "Facing 0" });
    this.rotRight = el("button", { text: "▶" });
    this.rotRight.addEventListener("click", () => this.rotate(1));
    this.discardBtn = el("button", { text: "Discard" });
    this.discardBtn.addEventListener("click", () => this.doDiscard());
    this.playBtn = el("button", { text: "Play" });
    this.playBtn.addEventListener("click", () => this.doPlayBattle());
    this.unluckyBtn = el("button", { text: "Unlucky" });
    this.unluckyBtn.addEventListener("click", () =>
      this.session.redrawUnlucky(),
    );
    this.endBtn = el("button", { text: "End turn" });
    this.endBtn.addEventListener("click", () => this.session.endTurn());
    controls.append(
      this.rotLeft,
      this.facingLabel,
      this.rotRight,
      this.discardBtn,
      this.playBtn,
      this.unluckyBtn,
      this.endBtn,
    );
    this.root.append(controls);
  }

  onEnter(): void {
    this.selectedTileId = null;
    this.aim = null;
    this.facing = 0;
    void this.mountBoard().then(() => {
      this.sync();
      this.layoutBoard();
    });
  }

  onLeave(): void {
    this.destroyBoard();
  }

  onSessionUpdate(): void {
    const me = this.session.match?.players.find(
      (p) => p.id === this.session.youId,
    );
    if (
      this.selectedTileId &&
      !me?.hand.some((t) => t.id === this.selectedTileId)
    ) {
      this.selectedTileId = null;
      this.aim = null;
    }
    this.sync();
  }

  onResize(): void {
    this.layoutBoard();
  }

  private layoutBoard(): void {
    if (!this.board || !this.pixi) return;
    const w = Math.max(1, Math.floor(this.boardHost.clientWidth));
    const h = Math.max(1, Math.floor(this.boardHost.clientHeight));
    this.pixi.renderer.resize(w, h);
    this.board.fit(w, h);
    this.board.position.set(w / 2, h / 2);
  }

  private async mountBoard(): Promise<void> {
    this.destroyBoard();
    // Wait two frames so flex layout assigns the board host height.
    await new Promise<void>((r) =>
      requestAnimationFrame(() => requestAnimationFrame(() => r())),
    );

    const w = Math.max(280, this.boardHost.clientWidth || 280);
    const h = Math.max(280, this.boardHost.clientHeight || 280);
    const app = new Application();
    await app.init({
      background: colors.bg,
      antialias: true,
      resolution: window.devicePixelRatio || 1,
      autoDensity: true,
      width: w,
      height: h,
    });
    this.boardHost.appendChild(app.canvas);
    const board = new HexBoard({ size: 36 });
    board.onPick = (hex) => this.onHex(hex);
    app.stage.addChild(board);
    this.pixi = app;
    this.board = board;

    this.resizeObserver = new ResizeObserver(() => this.layoutBoard());
    this.resizeObserver.observe(this.boardHost);
    this.layoutBoard();
  }

  private destroyBoard(): void {
    this.resizeObserver?.disconnect();
    this.resizeObserver = null;
    if (this.pixi) {
      this.pixi.destroy(true, { children: true });
      this.pixi = null;
    }
    this.board = null;
    clear(this.boardHost);
  }

  private rotate(delta: number): void {
    this.facing = (((this.facing + delta) % 6) + 6) % 6;
    this.facingLabel.textContent = `Facing ${this.facing}`;
    this.sync();
  }

  private sync(): void {
    const match = this.session.match;
    const room = this.session.room;
    const youId = this.session.youId;
    this.errorLabel.textContent = this.session.error ?? "";

    if (!match || !youId) {
      this.banner.textContent = "Waiting…";
      return;
    }

    const me = match.players.find((p) => p.id === youId);
    const opp = match.players.find((p) => p.id !== youId);
    const myTurn = match.turnPlayerId === youId;
    const selected = me?.hand.find((t) => t.id === this.selectedTileId);

    this.banner.textContent = phaseTitle(match.phase, myTurn);
    this.status.textContent = [
      room ? `Room ${room.code}` : "",
      match.mustDiscard ? "Discard one tile" : "",
      aimHint(this.aim),
      `Facing ${this.facing}`,
    ]
      .filter(Boolean)
      .join(" · ");
    this.facingLabel.textContent = `Facing ${this.facing}`;

    this.oppInfo.textContent = opp
      ? `Opponent · HQ ${opp.hqHp} · deck ${opp.deckCount} · discard ${opp.discardCount}`
      : "";
    this.myInfo.textContent = me
      ? `You · HQ ${me.hqHp} · deck ${me.deckCount} · discard ${me.discardCount}`
      : "";

    this.renderHand(this.oppHand, opp?.hand ?? [], false);
    this.renderHand(this.myHand, me?.hand ?? [], myTurn);

    if (this.board) {
      this.board.setOccupants(
        match.board.map((t) => ({
          q: t.q,
          r: t.r,
          defId: t.defId,
          facing: t.facing,
        })),
      );

      const canAct =
        myTurn && match.phase !== "battle" && match.phase !== "ended";
      this.updateBoardLegal(canAct, match.phase, selected);
    }

    setDisabled(
      this.discardBtn,
      !(myTurn && match.phase === "turn" && !!this.selectedTileId),
    );
    setDisabled(
      this.playBtn,
      !(
        myTurn &&
        match.phase === "turn" &&
        !match.mustDiscard &&
        selected?.kind === "instant" &&
        instantKind(selected.defId) === "battle"
      ),
    );
    setDisabled(
      this.endBtn,
      !(myTurn && match.phase === "turn" && !match.mustDiscard),
    );
    setDisabled(this.unluckyBtn, !(myTurn && match.unluckyAvailable));
    const canAct =
      myTurn && match.phase !== "battle" && match.phase !== "ended";
    setDisabled(this.rotLeft, !canAct);
    setDisabled(this.rotRight, !canAct);
  }

  private updateBoardLegal(
    canAct: boolean,
    phase: string,
    selected: HandTile | undefined,
  ): void {
    const board = this.board;
    const match = this.session.match;
    if (!board || !match) return;
    if (!canAct) {
      board.clearLegal();
      return;
    }
    if (phase === "place_hq") {
      board.setLegal(match.legalHexes ?? []);
      return;
    }
    if (phase === "turn" && !match.mustDiscard) {
      if (
        selected &&
        (selected.kind === "warrior" || selected.kind === "module")
      ) {
        board.setLegal(match.legalHexes ?? []);
        return;
      }
      if (this.aim) {
        board.clearLegal();
        return;
      }
    }
    board.clearLegal();
  }

  private renderHand(
    root: HTMLElement,
    hand: HandTile[],
    interactive: boolean,
  ): void {
    clear(root);
    for (const tile of hand) {
      const label = shortDef(tile.defId);
      const btn = el("button", { className: "tile" });
      const img = el("img", {
        attrs: {
          src: tileSvgUrl(tile.defId),
          alt: label,
          draggable: "false",
        },
      });
      btn.append(img, el("span", { className: "tile-label", text: label }));
      if (!interactive) btn.disabled = true;
      if (interactive && this.selectedTileId === tile.id) {
        btn.classList.add("selected");
      }
      btn.addEventListener("click", () => {
        if (!interactive) return;
        if (this.selectedTileId === tile.id) {
          this.selectedTileId = null;
          this.aim = null;
        } else {
          this.selectedTileId = tile.id;
          this.aim = aimFor(tile);
        }
        this.sync();
      });
      root.append(btn);
    }
  }

  private onHex(hex: Hex): void {
    const match = this.session.match;
    const youId = this.session.youId;
    if (!match || !youId || match.turnPlayerId !== youId) return;

    if (match.phase === "place_hq") {
      this.session.place("", hex.q, hex.r, this.facing);
      return;
    }
    if (match.phase !== "turn" || match.mustDiscard) return;

    if (this.aim) {
      this.onInstantHex(hex);
      return;
    }

    if (!this.selectedTileId) {
      this.session.setError("Select a tile first");
      return;
    }
    const me = match.players.find((p) => p.id === youId);
    const tile = me?.hand.find((t) => t.id === this.selectedTileId);
    if (!tile || (tile.kind !== "warrior" && tile.kind !== "module")) {
      this.session.setError("Select a unit, or target with an instant");
      return;
    }
    this.session.place(this.selectedTileId, hex.q, hex.r, this.facing);
    this.selectedTileId = null;
    this.aim = null;
  }

  private onInstantHex(hex: Hex): void {
    const match = this.session.match!;
    const youId = this.session.youId!;
    const tileId = this.selectedTileId!;
    const unit = match.board.find((t) => t.q === hex.q && t.r === hex.r);

    switch (this.aim?.kind) {
      case "move": {
        if (!this.aim.unitId) {
          if (!unit || unit.ownerId !== youId) {
            this.session.setError("Move: click your unit first");
            return;
          }
          this.aim = { kind: "move", unitId: unit.id };
          this.sync();
          return;
        }
        this.session.playInstant(tileId, {
          targetTileId: this.aim.unitId,
          q: hex.q,
          r: hex.r,
          facing: this.facing,
        });
        this.clearAim();
        return;
      }
      case "push": {
        if (!this.aim.unitId) {
          if (!unit || unit.ownerId === youId) {
            this.session.setError("Push: click an enemy unit");
            return;
          }
          this.aim = { kind: "push", unitId: unit.id };
          this.sync();
          return;
        }
        this.session.playInstant(tileId, {
          targetTileId: this.aim.unitId,
          q: hex.q,
          r: hex.r,
        });
        this.clearAim();
        return;
      }
      case "sniper":
      case "grenade": {
        if (!unit || unit.ownerId === youId || unit.kind === "hq") {
          this.session.setError("Click an enemy unit (not HQ)");
          return;
        }
        this.session.playInstant(tileId, { targetTileId: unit.id });
        this.clearAim();
        return;
      }
      case "airstrike": {
        this.session.playInstant(tileId, { q: hex.q, r: hex.r });
        this.clearAim();
        return;
      }
      default:
        return;
    }
  }

  private doPlayBattle(): void {
    if (!this.selectedTileId) return;
    this.session.playInstant(this.selectedTileId);
    this.clearAim();
  }

  private clearAim(): void {
    this.selectedTileId = null;
    this.aim = null;
  }

  private doDiscard(): void {
    if (!this.selectedTileId) {
      this.session.setError("Select a tile to discard");
      return;
    }
    this.session.discard(this.selectedTileId);
    this.clearAim();
  }
}

function aimFor(tile: HandTile): Aim {
  if (tile.kind !== "instant") return null;
  const k = instantKind(tile.defId);
  if (!k) return null;
  if (k === "battle") return { kind: "battle" };
  if (k === "move") return { kind: "move" };
  if (k === "push") return { kind: "push" };
  if (k === "sniper") return { kind: "sniper" };
  if (k === "airstrike") return { kind: "airstrike" };
  if (k === "grenade") return { kind: "grenade" };
  return null;
}

function instantKind(defId: string): InstantKind | null {
  const name = shortDef(defId);
  if (name === "battle") return "battle";
  if (name === "move") return "move";
  if (name === "push") return "push";
  if (name === "shot") return "sniper";
  if (name === "blast") return "airstrike";
  if (name === "grenade") return "grenade";
  return null;
}

function aimHint(aim: Aim): string {
  if (!aim) return "";
  switch (aim.kind) {
    case "battle":
      return "Battle: press Play";
    case "move":
      return aim.unitId
        ? "Move: click destination hex"
        : "Move: click your unit";
    case "push":
      return aim.unitId
        ? "Push: click destination hex"
        : "Push: click enemy unit";
    case "sniper":
      return "Shot: click enemy unit";
    case "grenade":
      return "Grenade: click enemy next to your HQ";
    case "airstrike":
      return "Blast: click center hex (must fit on board)";
  }
}

function phaseTitle(phase: string, myTurn: boolean): string {
  const turn = myTurn ? "Your turn" : "Opponent's turn";
  switch (phase) {
    case "place_hq":
      return `${turn} · Place HQ`;
    case "turn":
      return turn;
    case "battle":
      return "Battle";
    case "ended":
      return "Match over";
    default:
      return turn;
  }
}

function shortDef(defId: string): string {
  return defId.replace(/^(red|blue|green|yellow)_/, "");
}
