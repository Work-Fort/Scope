import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfToastContainer extends WfElement {
  @property({ type: String }) position: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' = 'top-right';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-toast-container');
    this._applyPosition();
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('position')) this._applyPosition();
  }

  private _applyPosition(): void {
    this.classList.remove(
      'wf-toast-container--top-right',
      'wf-toast-container--top-left',
      'wf-toast-container--bottom-right',
      'wf-toast-container--bottom-left',
    );
    this.classList.add(`wf-toast-container--${this.position}`);
  }
}

customElements.define('wf-toast-container', WfToastContainer);

declare global {
  interface HTMLElementTagNameMap {
    'wf-toast-container': WfToastContainer;
  }
}
