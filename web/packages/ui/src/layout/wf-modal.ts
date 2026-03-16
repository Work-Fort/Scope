import { html, noChange } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { trapFocus, createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

/**
 * `<wf-modal>` -- Modal dialog with backdrop, focus trap, and Escape dismissal.
 *
 * @element wf-modal
 * @slot - Default slot for dialog body content.
 * @fires wf-close -- When the dialog is closed.
 */
export class WfModal extends WfElement {
  @property({ type: Boolean, reflect: true }) open = false;
  @property({ type: String, reflect: true }) header = '';

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupFocus: (() => void) | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _previousFocus: HTMLElement | null = null;
  private _userContent: Node[] = [];
  private _didSetup = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-modal');

    // Capture user-provided children before Lit renders
    if (!this._didSetup) {
      this._userContent = Array.from(this.childNodes);
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  protected override updated(_changed: Map<string, unknown>): void {
    super.updated(_changed);

    // After first render, move captured children into the body div
    if (!this._didSetup) {
      this._didSetup = true;
      const body = this.querySelector('.wf-modal__body');
      if (body) {
        for (const node of this._userContent) {
          body.appendChild(node);
        }
      }
    }
  }

  private _handleOverlayClick = (e: MouseEvent): void => {
    // Only close if clicking directly on the overlay, not on the panel
    if (e.target === this) {
      this.hide();
    }
  };

  show(): void {
    this._previousFocus = document.activeElement as HTMLElement | null;
    this.open = true;
    this.classList.add('wf-modal--open');
    this.addEventListener('click', this._handleOverlayClick);

    this._backdrop = createBackdrop();
    this._cleanupEscape = onEscape(document, () => this.hide());

    // Defer focus trap setup to after render
    requestAnimationFrame(() => {
      const panel = this.querySelector('.wf-modal__panel') as HTMLElement;
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
    this.classList.remove('wf-modal--open');
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
    this.removeEventListener('click', this._handleOverlayClick);
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
        class="wf-modal__panel"
        role="dialog"
        aria-modal="true"
        aria-label=${this.header || 'Dialog'}
        ?hidden=${!this.open}
        @click=${(e: Event) => e.stopPropagation()}
      >
        <div class="wf-modal__header" ?hidden=${!this.header}>
          <span class="wf-modal__title">${this.header}</span>
          <button
            class="wf-modal__close"
            aria-label="Close"
            @click=${() => this.hide()}
          >&times;</button>
        </div>
        <div class="wf-modal__body"></div>
      </div>
    `;
  }
}

customElements.define('wf-modal', WfModal);

declare global {
  interface HTMLElementTagNameMap {
    'wf-modal': WfModal;
  }
}
