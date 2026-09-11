/** Single palette source for DOM CSS vars and Pixi numeric colors. */

export const palette = {
  bg: "#1a1a1a",
  panel: "#2a2a2a",
  button: "#3a3a3a",
  buttonHover: "#4a4a4a",
  buttonDisabled: "#252525",
  border: "#555555",
  focus: "#888888",
  text: "#ffffff",
  textMuted: "#aaaaaa",
  error: "#cc6666",
  army: {
    red: "#aa3333",
    blue: "#3366aa",
    green: "#339933",
    yellow: "#bbaa33",
  },
  edge: {
    melee: "#f0f0f0",
    ranged: "#66ccff",
    net: "#e8c84a",
    armor: "#888888",
    moduleLink: "#66cc88",
  },
  hex: {
    cell: "#2a2a2a",
    stroke: "#666666",
    legal: "#3a553a",
    selected: "#55553a",
    hover: "#3a3a55",
    blocked: "#222222",
  },
  font: 'Georgia, "Times New Roman", serif',
} as const;

function cssToPixi(css: string): number {
  return Number.parseInt(css.slice(1), 16);
}

/** Pixi-friendly numeric colors derived from `palette`. */
export const colors = {
  bg: cssToPixi(palette.bg),
  panel: cssToPixi(palette.panel),
  button: cssToPixi(palette.button),
  buttonHover: cssToPixi(palette.buttonHover),
  buttonDisabled: cssToPixi(palette.buttonDisabled),
  border: cssToPixi(palette.border),
  focus: cssToPixi(palette.focus),
  text: cssToPixi(palette.text),
  textMuted: cssToPixi(palette.textMuted),
  error: cssToPixi(palette.error),
  army: {
    red: cssToPixi(palette.army.red),
    blue: cssToPixi(palette.army.blue),
    green: cssToPixi(palette.army.green),
    yellow: cssToPixi(palette.army.yellow),
  },
  edge: {
    melee: cssToPixi(palette.edge.melee),
    ranged: cssToPixi(palette.edge.ranged),
    net: cssToPixi(palette.edge.net),
    armor: cssToPixi(palette.edge.armor),
    module_link: cssToPixi(palette.edge.moduleLink),
  },
  hex: {
    cell: cssToPixi(palette.hex.cell),
    stroke: cssToPixi(palette.hex.stroke),
    legal: cssToPixi(palette.hex.legal),
    selected: cssToPixi(palette.hex.selected),
    hover: cssToPixi(palette.hex.hover),
    blocked: cssToPixi(palette.hex.blocked),
  },
} as const;

export function edgeColor(kind: string): number {
  switch (kind) {
    case "melee":
      return colors.edge.melee;
    case "ranged":
      return colors.edge.ranged;
    case "net":
      return colors.edge.net;
    case "armor":
      return colors.edge.armor;
    case "module_link":
      return colors.edge.module_link;
    default:
      return colors.focus;
  }
}

/** Push palette values into CSS custom properties on `:root`. */
export function applyTheme(root: HTMLElement = document.documentElement): void {
  const s = root.style;
  s.setProperty("--bg", palette.bg);
  s.setProperty("--panel", palette.panel);
  s.setProperty("--button", palette.button);
  s.setProperty("--button-hover", palette.buttonHover);
  s.setProperty("--button-disabled", palette.buttonDisabled);
  s.setProperty("--border", palette.border);
  s.setProperty("--focus", palette.focus);
  s.setProperty("--text", palette.text);
  s.setProperty("--text-muted", palette.textMuted);
  s.setProperty("--error", palette.error);
  s.setProperty("--army-red", palette.army.red);
  s.setProperty("--army-blue", palette.army.blue);
  s.setProperty("--army-green", palette.army.green);
  s.setProperty("--army-yellow", palette.army.yellow);
  s.setProperty("--font", palette.font);
}
