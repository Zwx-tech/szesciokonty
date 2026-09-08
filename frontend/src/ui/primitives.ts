import { Container, Graphics, Rectangle, Text } from "pixi.js";
import { colors, fonts } from "./theme";

export class Label extends Text {
  constructor(
    text: string,
    opts?: { size?: number; color?: number; muted?: boolean },
  ) {
    super({
      text,
      style: {
        fontFamily: fonts.family,
        fontSize: opts?.size ?? fonts.size,
        fill: opts?.muted ? colors.textMuted : (opts?.color ?? colors.text),
      },
    });
  }

  setText(text: string): void {
    this.text = text;
  }
}

type ButtonOpts = {
  width?: number;
  height?: number;
  disabled?: boolean;
  fill?: number;
  fillHover?: number;
};

export class Button extends Container {
  private bg: Graphics;
  private caption: Label;
  private w: number;
  private h: number;
  private _disabled: boolean;
  private fill: number;
  private fillHover: number;
  private onClick: () => void;

  constructor(text: string, onClick: () => void, opts: ButtonOpts = {}) {
    super();
    this.w = opts.width ?? 200;
    this.h = opts.height ?? 44;
    this._disabled = opts.disabled ?? false;
    this.fill = opts.fill ?? colors.button;
    this.fillHover = opts.fillHover ?? colors.buttonHover;
    this.onClick = onClick;

    this.bg = new Graphics();
    this.addChild(this.bg);

    this.caption = new Label(text);
    this.caption.anchor.set(0.5);
    this.caption.position.set(this.w / 2, this.h / 2);
    this.addChild(this.caption);

    this.eventMode = "static";
    this.cursor = "pointer";
    this.on("pointertap", () => {
      if (!this._disabled) this.onClick();
    });
    this.on("pointerover", () => this.redraw(true));
    this.on("pointerout", () => this.redraw(false));

    this.redraw(false);
  }

  set disabled(v: boolean) {
    this._disabled = v;
    this.cursor = v ? "default" : "pointer";
    this.redraw(false);
  }

  get disabled(): boolean {
    return this._disabled;
  }

  setLabel(text: string): void {
    this.caption.setText(text);
  }

  setSelected(selected: boolean): void {
    this.bg
      .clear()
      .rect(0, 0, this.w, this.h)
      .fill(selected ? this.fillHover : this.fill)
      .stroke({
        width: selected ? 2 : 1,
        color: selected ? colors.focus : colors.border,
      });
  }

  private redraw(hover: boolean): void {
    const fill = this._disabled
      ? colors.buttonDisabled
      : hover
        ? this.fillHover
        : this.fill;
    this.bg
      .clear()
      .rect(0, 0, this.w, this.h)
      .fill(fill)
      .stroke({ width: 1, color: colors.border });
  }
}

type PanelOpts = { width: number; height: number };

export class Panel extends Container {
  private bg: Graphics;
  readonly content: Container;

  constructor(opts: PanelOpts) {
    super();
    this.bg = new Graphics();
    this.bg
      .rect(0, 0, opts.width, opts.height)
      .fill(colors.panel)
      .stroke({ width: 1, color: colors.border });
    this.addChild(this.bg);
    this.content = new Container();
    this.addChild(this.content);
  }
}

type TextFieldOpts = {
  width?: number;
  height?: number;
  placeholder?: string;
  maxLength?: number;
};

export class TextField extends Container {
  private bg: Graphics;
  private valueText: Label;
  private placeholderText: Label;
  private _value = "";
  private focused = false;
  private maxLength: number;
  private w: number;
  private h: number;
  private onChange?: (value: string) => void;
  private keyHandler = (e: KeyboardEvent) => this.onKey(e);

  constructor(opts: TextFieldOpts = {}) {
    super();
    this.w = opts.width ?? 280;
    this.h = opts.height ?? 44;
    this.maxLength = opts.maxLength ?? 24;

    this.bg = new Graphics();
    this.addChild(this.bg);

    this.placeholderText = new Label(opts.placeholder ?? "", { muted: true });
    this.placeholderText.position.set(
      12,
      (this.h - this.placeholderText.height) / 2,
    );
    this.addChild(this.placeholderText);

    this.valueText = new Label("");
    this.valueText.position.set(12, (this.h - fonts.size) / 2);
    this.addChild(this.valueText);

    this.eventMode = "static";
    this.cursor = "text";
    this.hitArea = new Rectangle(0, 0, this.w, this.h);

    this.on("pointertap", () => this.focus());
    this.redraw();
  }

  get value(): string {
    return this._value;
  }

  set value(v: string) {
    this._value = v.slice(0, this.maxLength);
    this.syncText();
    this.onChange?.(this._value);
  }

  setChangeHandler(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  focus(): void {
    if (this.focused) return;
    this.focused = true;
    window.addEventListener("keydown", this.keyHandler);
    this.redraw();
  }

  blur(): void {
    if (!this.focused) return;
    this.focused = false;
    window.removeEventListener("keydown", this.keyHandler);
    this.redraw();
  }

  private onKey(e: KeyboardEvent): void {
    if (!this.focused) return;
    if (e.key === "Backspace") {
      this.value = this._value.slice(0, -1);
      e.preventDefault();
      return;
    }
    if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
      this.value = this._value + e.key;
      e.preventDefault();
    }
  }

  private syncText(): void {
    this.valueText.setText(this._value);
    this.placeholderText.visible = this._value.length === 0;
  }

  private redraw(): void {
    this.bg
      .clear()
      .rect(0, 0, this.w, this.h)
      .fill(colors.bg)
      .stroke({ width: 1, color: this.focused ? colors.focus : colors.border });
  }

  override destroy(options?: boolean | { children?: boolean }): void {
    this.blur();
    super.destroy(options);
  }
}
