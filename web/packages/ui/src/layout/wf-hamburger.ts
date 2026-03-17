import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

export type HamburgerPosition = 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';

/**
 * `<wf-hamburger>` -- A hamburger button that opens a slide-out panel.
 * Position is configurable to any corner of the viewport.
 *
 * @element wf-hamburger
 * @slot - Default slot for panel content.
 * @fires wf-toggle -- When the button is clicked (detail: { open: boolean }).
 */
export class WfHamburger extends WfElement {
  @property({ type: String, reflect: true }) position: HamburgerPosition = 'top-right';
  @property({ type: Boolean, reflect: true }) open = false;

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _userContent: Node[] = [];
  private _didSetup = false;
  private _childObserver: MutationObserver | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-hamburger');

    if (!this._didSetup) {
      this._userContent = Array.from(this.childNodes);
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
    if (this._childObserver) {
      this._childObserver.disconnect();
      this._childObserver = null;
    }
  }

  protected override updated(changed: Map<string, unknown>): void {
    super.updated(changed);

    if (!this._didSetup) {
      this._didSetup = true;
      const body = this.querySelector('.wf-hamburger__body');
      if (body) {
        for (const node of this._userContent) {
          body.appendChild(node);
        }
      }
      this._observeChildren();
    }

    if (changed.has('open')) {
      if (this.open) {
        this._setup();
      } else {
        this._teardown();
      }
    }
  }

  private _toggle(): void {
    const next = !this.open;
    this.dispatchEvent(
      new CustomEvent('wf-toggle', {
        bubbles: true,
        composed: true,
        detail: { open: next },
      }),
    );
    this.open = next;
  }

  private _close(): void {
    this.open = false;
    this.dispatchEvent(
      new CustomEvent('wf-toggle', {
        bubbles: true,
        composed: true,
        detail: { open: false },
      }),
    );
  }

  private _setup(): void {
    this._backdrop = createBackdrop(() => this._close());
    this._cleanupEscape = onEscape(this, () => this._close());
  }

  private _teardown(): void {
    if (this._backdrop) {
      removeBackdrop(this._backdrop);
      this._backdrop = null;
    }
    if (this._cleanupEscape) {
      this._cleanupEscape();
      this._cleanupEscape = null;
    }
  }

  /** Watch for children appended directly to the host and adopt them into the panel body. */
  private _observeChildren(): void {
    const body = this.querySelector('.wf-hamburger__body');
    if (!body) return;

    this._childObserver = new MutationObserver((mutations) => {
      for (const m of mutations) {
        for (const node of m.addedNodes) {
          if (node instanceof Element && (
            node.classList.contains('wf-hamburger__button') ||
            node.classList.contains('wf-hamburger__panel')
          )) continue;
          // Skip Lit's own template marker nodes (comments)
          if (node.nodeType === Node.COMMENT_NODE) continue;
          body.appendChild(node);
        }
      }
    });

    this._childObserver.observe(this, { childList: true });
  }

  /** Determine the panel slide direction from position. */
  private _panelSide(): 'left' | 'right' {
    return this.position.includes('left') ? 'left' : 'right';
  }

  render() {
    return html`
      <button
        class="wf-hamburger__button wf-hamburger__button--${this.position}"
        aria-label="Menu"
        aria-expanded=${String(this.open)}
        @click=${() => this._toggle()}
      >&#9776;</button>
      <div
        class="wf-hamburger__panel wf-hamburger__panel--${this._panelSide()}"
        ?hidden=${!this.open}
      >
        <div class="wf-hamburger__body"></div>
      </div>
    `;
  }
}

customElements.define('wf-hamburger', WfHamburger);

declare global {
  interface HTMLElementTagNameMap {
    'wf-hamburger': WfHamburger;
  }
}
