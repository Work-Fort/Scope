import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfStatusDot extends WfElement {
  @property({ type: String }) status: 'online' | 'offline' | 'away' = 'offline';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-status-dot');
    this._applyStatus();
  }

  updated(): void {
    this._applyStatus();
  }

  private _applyStatus(): void {
    this.classList.remove('wf-status-dot--online', 'wf-status-dot--away', 'wf-status-dot--offline');
    this.classList.add(`wf-status-dot--${this.status}`);
    this.setAttribute('aria-label', this.status);
  }
}

customElements.define('wf-status-dot', WfStatusDot);

declare global {
  interface HTMLElementTagNameMap {
    'wf-status-dot': WfStatusDot;
  }
}
