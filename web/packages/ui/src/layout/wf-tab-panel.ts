import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-tab-panel>` -- A panel within a `<wf-tabs>` container.
 *
 * @element wf-tab-panel
 * @slot - Default slot for panel content.
 */
export class WfTabPanel extends WfElement {
  static get properties() {
    return {
      name: { type: String, reflect: true },
      label: { type: String, reflect: true },
      active: { type: Boolean, reflect: true },
    };
  }

  name = '';
  label = '';
  active = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-tab-panel');
    this.setAttribute('role', 'tabpanel');
    this._syncVisibility();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncVisibility();
  }

  private _syncVisibility(): void {
    this.style.display = this.active ? '' : 'none';
  }

  render() {
    return html`<slot></slot>`;
  }
}

customElements.define('wf-tab-panel', WfTabPanel);

declare global {
  interface HTMLElementTagNameMap {
    'wf-tab-panel': WfTabPanel;
  }
}
