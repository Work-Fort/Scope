import { html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfPanel extends WfElement {
  @property({ type: String }) label = '';

  private _userContent: Node[] = [];
  private _didSetup = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-panel');

    // Capture user-provided children before Lit renders
    if (!this._didSetup) {
      this._userContent = Array.from(this.childNodes);
    }
  }

  protected override updated(_changed: Map<string, unknown>): void {
    super.updated(_changed);

    // After first render, re-append children so they come after the label
    if (!this._didSetup) {
      this._didSetup = true;
      for (const node of this._userContent) {
        this.appendChild(node);
      }
    }
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
