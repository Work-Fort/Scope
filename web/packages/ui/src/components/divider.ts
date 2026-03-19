import { html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfDivider extends WfElement {
  @property({ type: String }) label?: string;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-divider');
    this.setAttribute('role', 'separator');
  }

  render() {
    if (!this.label) return nothing;
    return html`
      <span class="wf-divider__line"></span>
      <span class="wf-divider__text">${this.label}</span>
      <span class="wf-divider__line"></span>
    `;
  }
}

customElements.define('wf-divider', WfDivider);

declare global {
  interface HTMLElementTagNameMap {
    'wf-divider': WfDivider;
  }
}
