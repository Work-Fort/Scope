import { WfElement } from '../base.js';

export class WfList extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-list');
    this.setAttribute('role', 'list');
  }
}

customElements.define('wf-list', WfList);

declare global {
  interface HTMLElementTagNameMap {
    'wf-list': WfList;
  }
}
