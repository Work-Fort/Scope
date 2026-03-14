// src/form/wf-field.ts
import { LionField } from '@lion/ui/form-core.js';

/**
 * Base class for WorkFort form fields.
 * Extends LionField with light DOM rendering and WorkFort CSS classes.
 * Uses Lion's default slot-based rendering, applying WorkFort CSS classes
 * to the generated elements.
 *
 * Subclasses override:
 * - `_wfTag` for the component-specific CSS class (e.g., 'wf-input')
 */
export class WfField extends LionField {
  createRenderRoot(): this {
    return this;
  }

  /** CSS class added to host. Override in subclasses. */
  protected get _wfTag(): string {
    return '';
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field');
    if (this._wfTag) this.classList.add(this._wfTag);
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

  /** Apply WorkFort CSS classes to Lion's slot-based elements. */
  protected _applyFieldClasses(): void {
    // Label
    const label = this.querySelector('[slot="label"]');
    if (label && !label.classList.contains('wf-field__label')) {
      label.classList.add('wf-field__label');
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
