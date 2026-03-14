// src/form/wf-input.ts
import { LionInput } from '@lion/ui/input.js';
import { applyFieldHostClasses, applyFieldClasses } from './field-styles.js';

/**
 * `<wf-input>` — Text input with validation, formatting, and interaction states.
 * Extends LionInput (not WfField) because LionInput adds NativeTextFieldMixin.
 * Uses Lion's default slot-based rendering in light DOM, with WorkFort CSS
 * classes applied to the generated elements.
 *
 * @element wf-input
 */
export class WfInput extends LionInput {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    applyFieldHostClasses(this as any, 'wf-input');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    applyFieldClasses(this as any);
  }
}

customElements.define('wf-input', WfInput);

declare global {
  interface HTMLElementTagNameMap {
    'wf-input': WfInput;
  }
}
