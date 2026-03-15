import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { trapFocus, createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

/**
 * `<wf-dialog>` -- Modal dialog with backdrop, focus trap, and Escape dismissal.
 *
 * @element wf-dialog
 * @slot - Default slot for dialog body content.
 * @fires wf-close -- When the dialog is closed.
 */
export class WfDialog extends WfElement {
  @property({ type: Boolean, reflect: true }) open = false;
  @property({ type: String, reflect: true }) header = '';

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupFocus: (() => void) | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _previousFocus: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-dialog');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  show(): void {
    this._previousFocus = document.activeElement as HTMLElement | null;
    this.open = true;
    this.classList.add('wf-dialog--open');

    this._backdrop = createBackdrop(() => this.hide());
    this._cleanupEscape = onEscape(this, () => this.hide());

    // Defer focus trap setup to after render
    requestAnimationFrame(() => {
      const panel = this.querySelector('.wf-dialog__panel') as HTMLElement;
      if (panel) {
        this._cleanupFocus = trapFocus(panel);
        // Focus the first focusable element or the panel itself
        const firstFocusable = panel.querySelector(
          'button, [tabindex]:not([tabindex="-1"])',
        ) as HTMLElement | null;
        if (firstFocusable) firstFocusable.focus();
      }
    });
  }

  hide(): void {
    this.open = false;
    this.classList.remove('wf-dialog--open');
    this._teardown();

    this.dispatchEvent(
      new CustomEvent('wf-close', {
        bubbles: true,
        composed: true,
      }),
    );

    // Restore focus
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
        class="wf-dialog__panel"
        role="dialog"
        aria-modal="true"
        aria-label=${this.header || 'Dialog'}
        ?hidden=${!this.open}
      >
        <div class="wf-dialog__header" ?hidden=${!this.header}>
          <span class="wf-dialog__title">${this.header}</span>
          <button
            class="wf-dialog__close"
            aria-label="Close"
            @click=${() => this.hide()}
          >&times;</button>
        </div>
        <div class="wf-dialog__body">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define('wf-dialog', WfDialog);

declare global {
  interface HTMLElementTagNameMap {
    'wf-dialog': WfDialog;
  }
}
