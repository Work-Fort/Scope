import { WfElement } from '../base.js';

export class WfScrollArea extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-scroll-area');
    if (!this.hasAttribute('tabindex')) {
      this.setAttribute('tabindex', '0');
    }
    if (!this.hasAttribute('role')) {
      this.setAttribute('role', 'region');
    }
    if (!this.hasAttribute('aria-label')) {
      this.setAttribute('aria-label', 'Scrollable region');
    }
  }
}

customElements.define('wf-scroll-area', WfScrollArea);

declare global {
  interface HTMLElementTagNameMap {
    'wf-scroll-area': WfScrollArea;
  }
}
