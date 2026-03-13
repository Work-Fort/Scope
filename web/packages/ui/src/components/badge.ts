import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfBadge extends WfElement {
  @property({ type: Number }) count = 0;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-badge');
    this._sync();
  }

  updated(): void {
    this._sync();
  }

  private _sync(): void {
    this.textContent = this.count > 0 ? String(this.count) : '';
    this.style.display = this.count > 0 ? '' : 'none';
  }
}

customElements.define('wf-badge', WfBadge);

declare global {
  interface HTMLElementTagNameMap {
    'wf-badge': WfBadge;
  }
}
