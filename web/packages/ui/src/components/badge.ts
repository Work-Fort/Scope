import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfBadge extends WfElement {
  @property({ type: Number }) count = 0;
  @property({ type: String, reflect: true }) size: 'sm' | 'md' | 'lg' = 'md';
  @property({ type: Boolean, reflect: true }) hidden = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-badge');
    this._sync();
  }

  updated(): void {
    this._sync();
  }

  private _sync(): void {
    this.textContent = String(this.count);
    this.style.display = this.hidden ? 'none' : '';
    this.classList.remove('wf-badge--sm', 'wf-badge--md', 'wf-badge--lg');
    this.classList.add(`wf-badge--${this.size}`);
  }
}

customElements.define('wf-badge', WfBadge);

declare global {
  interface HTMLElementTagNameMap {
    'wf-badge': WfBadge;
  }
}
