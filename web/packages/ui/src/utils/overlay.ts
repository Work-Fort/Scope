/**
 * Lightweight overlay utility for focus trapping and backdrop management.
 * Used by wf-dialog, wf-drawer, wf-tooltip, and wf-popover.
 */

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

/**
 * Trap focus within a container element.
 * Returns a cleanup function that removes the event listener.
 */
export function trapFocus(container: HTMLElement): () => void {
  const handler = (e: KeyboardEvent) => {
    if (e.key !== 'Tab') return;

    const focusable = Array.from(
      container.querySelectorAll(FOCUSABLE_SELECTOR),
    ) as HTMLElement[];
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (e.shiftKey) {
      if (document.activeElement === first) {
        e.preventDefault();
        last.focus();
      }
    } else {
      if (document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }
  };

  container.addEventListener('keydown', handler);
  return () => container.removeEventListener('keydown', handler);
}

/**
 * Create and append a backdrop element to document.body.
 * Returns the backdrop element. Caller is responsible for removal.
 */
export function createBackdrop(onClick?: () => void): HTMLDivElement {
  const backdrop = document.createElement('div');
  backdrop.classList.add('wf-overlay-backdrop');
  if (onClick) {
    backdrop.addEventListener('click', onClick);
  }
  document.body.appendChild(backdrop);
  return backdrop;
}

/**
 * Remove a backdrop element from the DOM.
 */
export function removeBackdrop(backdrop: HTMLDivElement): void {
  backdrop.remove();
}

/**
 * Listen for Escape key on a target element.
 * Returns a cleanup function.
 */
export function onEscape(
  target: HTMLElement | Document,
  callback: () => void,
): () => void {
  const handler = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.stopPropagation();
      callback();
    }
  };
  target.addEventListener('keydown', handler as EventListener);
  return () => target.removeEventListener('keydown', handler as EventListener);
}
