import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import { initials } from '../utils/initials.js';

export class WfAvatar extends WfElement {
  @property({ type: String }) username = '';
  @property({ type: String, reflect: true }) size: 'sm' | 'md' = 'md';
  @property({ type: String }) status?: string;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-avatar');
  }

  render() {
    return html`
      <span class="wf-avatar__initials">${initials(this.username)}</span>
      ${this.status
        ? html`<wf-status-dot class="wf-avatar__dot" status=${this.status}></wf-status-dot>`
        : ''}
    `;
  }
}

customElements.define('wf-avatar', WfAvatar);

declare global {
  interface HTMLElementTagNameMap {
    'wf-avatar': WfAvatar;
  }
}
