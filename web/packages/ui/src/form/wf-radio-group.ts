// src/form/wf-radio-group.ts
import { LionRadioGroup } from '@lion/ui/radio-group.js';

/**
 * `<wf-radio-group>` — Container for multiple wf-radio elements.
 * Handles group validation and mutual exclusion.
 * Extends LionRadioGroup with light DOM rendering and WorkFort CSS classes.
 *
 * @element wf-radio-group
 */
export class WfRadioGroup extends LionRadioGroup {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-radio-group');
  }
}

customElements.define('wf-radio-group', WfRadioGroup);

declare global {
  interface HTMLElementTagNameMap {
    'wf-radio-group': WfRadioGroup;
  }
}
