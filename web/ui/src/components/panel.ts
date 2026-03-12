import { html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfPanel extends WfElement {
  @property({ type: String }) label = '';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-panel');
  }

  render() {
    return this.label
      ? html`<div class="wf-panel__label">${this.label}</div>`
      : nothing;
  }
}

customElements.define('wf-panel', WfPanel);

declare global {
  interface HTMLElementTagNameMap {
    'wf-panel': WfPanel;
  }
}
