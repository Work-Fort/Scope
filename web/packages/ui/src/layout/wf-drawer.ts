import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { trapFocus, createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

/**
 * `<wf-drawer>` -- Slide-in panel from screen edge. Supports left, right, top, bottom.
 *
 * @element wf-drawer
 * @slot - Default slot for drawer content.
 * @fires wf-close -- When the drawer is closed.
 */
export class WfDrawer extends WfElement {
  @property({ type: Boolean, reflect: true }) open = false;
  @property({ type: String, reflect: true }) header = '';
  @property({ type: String, reflect: true }) position: 'left' | 'right' | 'top' | 'bottom' = 'right';

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupFocus: (() => void) | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _previousFocus: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-drawer');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  show(): void {
    this._previousFocus = document.activeElement as HTMLElement | null;
    this.open = true;
    this.classList.add('wf-drawer--open');

    this._backdrop = createBackdrop(() => this.hide());
    this._cleanupEscape = onEscape(this, () => this.hide());

    requestAnimationFrame(() => {
      const panel = this.querySelector('.wf-drawer__panel') as HTMLElement;
      if (panel) {
        this._cleanupFocus = trapFocus(panel);
      }
    });
  }

  hide(): void {
    this.open = false;
    this.classList.remove('wf-drawer--open');
    this._teardown();

    this.dispatchEvent(
      new CustomEvent('wf-close', {
        bubbles: true,
        composed: true,
      }),
    );

    if (this._previousFocus) {
      this._previousFocus.focus();
      this._previousFocus = null;
    }
  }

  private _teardown(): void {
    if (this._backdrop) {
      removeBackdrop(this._backdrop);
      this._backdrop = null;
    }
    if (this._cleanupFocus) {
      this._cleanupFocus();
      this._cleanupFocus = null;
    }
    if (this._cleanupEscape) {
      this._cleanupEscape();
      this._cleanupEscape = null;
    }
  }

  render() {
    return html`
      <div
        class="wf-drawer__panel wf-drawer__panel--${this.position}"
        role="dialog"
        aria-modal="true"
        aria-label=${this.header || 'Drawer'}
        ?hidden=${!this.open}
      >
        <div class="wf-drawer__header" ?hidden=${!this.header}>
          <span class="wf-drawer__title">${this.header}</span>
          <button
            class="wf-drawer__close"
            aria-label="Close"
            @click=${() => this.hide()}
          >&times;</button>
        </div>
        <div class="wf-drawer__body">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define('wf-drawer', WfDrawer);

declare global {
  interface HTMLElementTagNameMap {
    'wf-drawer': WfDrawer;
  }
}
