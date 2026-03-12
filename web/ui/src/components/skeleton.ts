import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfSkeleton extends WfElement {
  @property({ type: String }) width = '100%';
  @property({ type: String }) height = '1em';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-skeleton');
    this.setAttribute('aria-hidden', 'true');
    this._applyDimensions();
  }

  updated(): void {
    this._applyDimensions();
  }

  private _applyDimensions(): void {
    this.style.width = this.width;
    this.style.height = this.height;
  }
}

customElements.define('wf-skeleton', WfSkeleton);

declare global {
  interface HTMLElementTagNameMap {
    'wf-skeleton': WfSkeleton;
  }
}
