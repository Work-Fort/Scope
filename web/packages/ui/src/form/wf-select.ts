// src/form/wf-select.ts
import { LionSelect } from '@lion/ui/select.js';
import { applyFieldHostClasses, applyFieldClasses } from './field-styles.js';

/**
 * `<wf-select>` — Native select dropdown with validation and interaction states.
 * Extends LionSelect with light DOM rendering and WorkFort CSS classes.
 *
 * @element wf-select
 */
export class WfSelect extends LionSelect {
  createRenderRoot(): this {
    return this;
  }

  get slots() {
    return {
      ...super.slots,
      input: () => document.createElement('select'),
    };
  }

  connectedCallback(): void {
    super.connectedCallback();
    applyFieldHostClasses(this as any, 'wf-select');
    this._moveOptionsToSelect();
  }

  /** Move <option> children from the host into the native <select>. */
  private _moveOptionsToSelect(): void {
    const select = this._inputNode as HTMLSelectElement;
    if (!select) return;
    const options = Array.from(this.querySelectorAll(':scope > option'));
    for (const opt of options) {
      select.appendChild(opt);
    }
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    applyFieldClasses(this as any);
  }
}

customElements.define('wf-select', WfSelect);

declare global {
  interface HTMLElementTagNameMap {
    'wf-select': WfSelect;
  }
}
