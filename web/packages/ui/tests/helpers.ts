/**
 * Create and connect a Custom Element for testing.
 * Waits for Lit's updateComplete before returning.
 */
export async function fixture<T extends HTMLElement>(
  tag: string,
  attrs?: Record<string, string | number | boolean>,
): Promise<T> {
  const el = document.createElement(tag) as T;
  if (attrs) {
    for (const [key, val] of Object.entries(attrs)) {
      if (typeof val === 'boolean') {
        if (val) el.setAttribute(key, '');
      } else {
        el.setAttribute(key, String(val));
      }
    }
  }
  document.body.appendChild(el);
  if ('updateComplete' in el) {
    await (el as any).updateComplete;
  }
  return el;
}

/** Remove all children from document.body between tests. */
export function cleanup(): void {
  while (document.body.firstChild) {
    document.body.removeChild(document.body.firstChild);
  }
}
