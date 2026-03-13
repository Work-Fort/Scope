import { WfElement } from '../base.js';

export class WfDivider extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-divider');
    this.setAttribute('role', 'separator');
  }
}

customElements.define('wf-divider', WfDivider);

declare global {
  interface HTMLElementTagNameMap {
    'wf-divider': WfDivider;
  }
}
