// src/form/wf-checkbox.ts
import { LionCheckbox } from '@lion/ui/checkbox-group.js';
import { applyFieldHostClasses, applyFieldClasses } from './field-styles.js';

/**
 * `<wf-checkbox>` — Checkbox input with validation and interaction states.
 * Extends LionCheckbox (ChoiceInputMixin + LionInput) with light DOM rendering
 * and WorkFort CSS classes.
 *
 * @element wf-checkbox
 */
export class WfCheckbox extends LionCheckbox {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    applyFieldHostClasses(this as any, 'wf-checkbox');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    applyFieldClasses(this as any);
  }
}

customElements.define('wf-checkbox', WfCheckbox);

declare global {
  interface HTMLElementTagNameMap {
    'wf-checkbox': WfCheckbox;
  }
}
