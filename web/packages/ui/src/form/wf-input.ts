// src/form/wf-input.ts
import { LionInput } from '@lion/ui/input.js';

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
    this.classList.add('wf-field', 'wf-input');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);

    // Sync error class
    this.classList.toggle(
      'wf-field--error',
      !!(this as any).showsFeedbackFor?.includes('error'),
    );

    // Sync disabled class
    this.classList.toggle('wf-field--disabled', !!this.disabled);

    // Apply WorkFort CSS classes to Lion's generated elements
    this._applyFieldClasses();
  }

  private _applyFieldClasses(): void {
    // Label
    const label = this.querySelector('[slot="label"]');
    if (label && !label.classList.contains('wf-field__label')) {
      label.classList.add('wf-field__label');
      // Add required indicator
      const isRequired = this.validators?.some(
        (v: any) => v.constructor?.validatorName === 'Required',
      );
      label.classList.toggle('wf-field__label--required', !!isRequired);
    }

    // Native input
    if (
      this._inputNode &&
      !this._inputNode.classList.contains('wf-field__input')
    ) {
      this._inputNode.classList.add('wf-field__input');
    }

    // Help text
    const helpText = this.querySelector('[slot="help-text"]');
    if (helpText && !helpText.classList.contains('wf-field__help-text')) {
      helpText.classList.add('wf-field__help-text');
    }

    // Feedback
    const feedback = this.querySelector('lion-validation-feedback');
    if (feedback && !feedback.classList.contains('wf-field__feedback')) {
      feedback.classList.add('wf-field__feedback');
    }
  }
}

customElements.define('wf-input', WfInput);

declare global {
  interface HTMLElementTagNameMap {
    'wf-input': WfInput;
  }
}
