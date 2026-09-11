import { describe, expect, it } from "vitest";
import { applyTheme, colors, edgeColor, palette } from "./theme";

describe("theme", () => {
  it("derives Pixi colors from the CSS palette", () => {
    expect(colors.bg).toBe(0x1a1a1a);
    expect(colors.army.red).toBe(0xaa3333);
    expect(colors.edge.net).toBe(0xe8c84a);
  });

  it("maps edge kinds to colors", () => {
    expect(edgeColor("melee")).toBe(colors.edge.melee);
    expect(edgeColor("module_link")).toBe(colors.edge.module_link);
    expect(edgeColor("unknown")).toBe(colors.focus);
  });

  it("applyTheme writes CSS custom properties", () => {
    const props = new Map<string, string>();
    const el = {
      style: {
        setProperty: (k: string, v: string) => {
          props.set(k, v);
        },
      },
    } as unknown as HTMLElement;
    applyTheme(el);
    expect(props.get("--bg")).toBe(palette.bg);
    expect(props.get("--army-red")).toBe(palette.army.red);
    expect(props.get("--font")).toContain("Georgia");
  });
});
