export function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  opts?: {
    className?: string;
    text?: string;
    attrs?: Record<string, string>;
  },
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (opts?.className) node.className = opts.className;
  if (opts?.text != null) node.textContent = opts.text;
  if (opts?.attrs) {
    for (const [k, v] of Object.entries(opts.attrs)) {
      node.setAttribute(k, v);
    }
  }
  return node;
}

export function clear(node: HTMLElement): void {
  node.replaceChildren();
}

export function setDisabled(btn: HTMLButtonElement, disabled: boolean): void {
  btn.disabled = disabled;
}
