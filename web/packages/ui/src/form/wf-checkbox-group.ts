// src/form/wf-checkbox-group.ts
import { LionCheckboxGroup } from '@lion/ui/checkbox-group.js';

/**
 * `<wf-checkbox-group>` — Container for multiple wf-checkbox elements.
 * Handles group validation (e.g., "select at least 2").
 * Extends LionCheckboxGroup with light DOM rendering and WorkFort CSS classes.
 *
 * @element wf-checkbox-group
 */
export class WfCheckboxGroup extends LionCheckboxGroup {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-checkbox-group');
  }
}

customElements.define('wf-checkbox-group', WfCheckboxGroup);

declare global {
  interface HTMLElementTagNameMap {
    'wf-checkbox-group': WfCheckboxGroup;
  }
}
