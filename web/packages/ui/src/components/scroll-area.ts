import { WfElement } from '../base.js';

export class WfScrollArea extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-scroll-area');
  }
}

customElements.define('wf-scroll-area', WfScrollArea);

declare global {
  interface HTMLElementTagNameMap {
    'wf-scroll-area': WfScrollArea;
  }
}
