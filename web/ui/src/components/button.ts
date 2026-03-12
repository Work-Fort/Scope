import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfButton extends WfElement {
  @property({ type: String }) variant: 'text' | 'filled' = 'text';
  @property({ type: Boolean, reflect: true }) disabled = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-button');
    this.setAttribute('role', 'button');
    this.setAttribute('tabindex', '0');
    this.addEventListener('click', this._handleClick);
    this.addEventListener('keydown', this._handleKeydown);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('click', this._handleClick);
    this.removeEventListener('keydown', this._handleKeydown);
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('variant')) {
      this.classList.toggle('wf-button--filled', this.variant === 'filled');
    }
    if (changed.has('disabled')) {
      this.setAttribute('aria-disabled', String(this.disabled));
      this.setAttribute('tabindex', this.disabled ? '-1' : '0');
    }
  }

  private _handleClick = (e: Event): void => {
    if (this.disabled) {
      e.stopImmediatePropagation();
      return;
    }
    this.dispatchEvent(new CustomEvent('wf-click', { bubbles: true, composed: true }));
  };

  private _handleKeydown = (e: KeyboardEvent): void => {
    if ((e.key === 'Enter' || e.key === ' ') && !this.disabled) {
      e.preventDefault();
      this.dispatchEvent(new CustomEvent('wf-click', { bubbles: true, composed: true }));
    }
  };
}

customElements.define('wf-button', WfButton);

declare global {
  interface HTMLElementTagNameMap {
    'wf-button': WfButton;
  }
}
