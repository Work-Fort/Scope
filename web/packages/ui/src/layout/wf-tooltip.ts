import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-tooltip>` -- Hover/focus popup that shows text content.
 * Wrap the trigger element as a child. Set `content` for the tooltip text.
 *
 * @element wf-tooltip
 * @slot - Default slot for the trigger element.
 */
export class WfTooltip extends WfElement {
  @property({ type: String, reflect: true }) content = '';
  @property({ type: String, reflect: true }) position: 'top' | 'bottom' | 'left' | 'right' = 'top';
  @property({ type: Boolean }) _visible = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-tooltip');
    this.addEventListener('mouseenter', this._show);
    this.addEventListener('mouseleave', this._hide);
    this.addEventListener('focusin', this._show);
    this.addEventListener('focusout', this._hide);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('mouseenter', this._show);
    this.removeEventListener('mouseleave', this._hide);
    this.removeEventListener('focusin', this._show);
    this.removeEventListener('focusout', this._hide);
  }

  private _show = (): void => {
    this._visible = true;
  };

  private _hide = (): void => {
    this._visible = false;
  };

  render() {
    return html`
      <slot></slot>
      <span
        class="wf-tooltip__content wf-tooltip__content--${this.position}"
        role="tooltip"
        aria-hidden=${!this._visible}
      >
        ${this.content}
      </span>
    `;
  }
}

customElements.define('wf-tooltip', WfTooltip);

declare global {
  interface HTMLElementTagNameMap {
    'wf-tooltip': WfTooltip;
  }
}
