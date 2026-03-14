// src/form/wf-radio.ts
import { LionRadio } from '@lion/ui/radio-group.js';
import { applyFieldHostClasses, applyFieldClasses } from './field-styles.js';

/**
 * `<wf-radio>` — Radio input for use inside a wf-radio-group.
 * Extends LionRadio (ChoiceInputMixin + LionInput) with light DOM rendering
 * and WorkFort CSS classes.
 *
 * @element wf-radio
 */
export class WfRadio extends LionRadio {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    applyFieldHostClasses(this as any, 'wf-radio');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    applyFieldClasses(this as any);
  }
}

customElements.define('wf-radio', WfRadio);

declare global {
  interface HTMLElementTagNameMap {
    'wf-radio': WfRadio;
  }
}
