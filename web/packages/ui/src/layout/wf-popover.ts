import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { onEscape } from '../utils/overlay.js';

/**
 * `<wf-popover>` -- Click-triggered popup with rich content.
 * Wrap the trigger element as the default slot. Use the `content` slot
 * or set innerHTML in the popover body.
 *
 * @element wf-popover
 * @slot - Default slot for the trigger element.
 * @fires wf-close -- When the popover is closed.
 */
export class WfPopover extends WfElement {
  @property({ type: Boolean, reflect: true }) open = false;
  @property({ type: String, reflect: true }) position: 'top' | 'bottom' | 'left' | 'right' = 'bottom';

  private _cleanupEscape: (() => void) | null = null;
  private _boundDocClick: ((e: Event) => void) | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-popover');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  toggle(): void {
    if (this.open) {
      this.hide();
    } else {
      this.show();
    }
  }

  show(): void {
    this.open = true;
    this.classList.add('wf-popover--open');

    this._cleanupEscape = onEscape(this, () => this.hide());

    // Close on outside click — use microtask to avoid catching the opening click
    this._boundDocClick = (e: Event) => {
      if (!this.contains(e.target as Node)) {
        this.hide();
      }
    };
    queueMicrotask(() => {
      if (this._boundDocClick) {
        document.addEventListener('click', this._boundDocClick);
      }
    });
  }

  hide(): void {
    this.open = false;
    this.classList.remove('wf-popover--open');
    this._teardown();

    this.dispatchEvent(
      new CustomEvent('wf-close', {
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _teardown(): void {
    if (this._cleanupEscape) {
      this._cleanupEscape();
      this._cleanupEscape = null;
    }
    if (this._boundDocClick) {
      document.removeEventListener('click', this._boundDocClick);
      this._boundDocClick = null;
    }
  }

  render() {
    return html`
      <div class="wf-popover__trigger" @click=${(e: Event) => { e.stopPropagation(); this.toggle(); }}>
        <slot></slot>
      </div>
      <div
        class="wf-popover__content wf-popover__content--${this.position}"
        ?hidden=${!this.open}
      >
        <slot name="content"></slot>
      </div>
    `;
  }
}

customElements.define('wf-popover', WfPopover);

declare global {
  interface HTMLElementTagNameMap {
    'wf-popover': WfPopover;
  }
}
