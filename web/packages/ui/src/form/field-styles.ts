// src/form/field-styles.ts
// Shared CSS class application logic for WorkFort form field components.
// Eliminates duplication across WfField, WfInput, WfTextarea, WfSelect, etc.

import type { LitElement } from 'lit';

type LionFieldLike = LitElement & {
  classList: DOMTokenList;
  _inputNode: HTMLElement;
  disabled: boolean;
  validators?: Array<{ constructor?: { validatorName?: string } }>;
  querySelector(selector: string): Element | null;
  showsFeedbackFor?: string[];
};

/**
 * Add the standard `wf-field` host class plus a component-specific class.
 * Call from `connectedCallback()`.
 */
export function applyFieldHostClasses(
  host: LionFieldLike,
  componentClass: string,
): void {
  host.classList.add('wf-field', componentClass);
}

/**
 * Toggle error/disabled state classes and apply CSS classes to Lion's
 * generated slot elements (label, input, help-text, feedback).
 * Call from `updated()`.
 */
export function applyFieldClasses(host: LionFieldLike): void {
  // State classes on host
  host.classList.toggle(
    'wf-field--error',
    !!(host as any).showsFeedbackFor?.includes('error'),
  );
  host.classList.toggle('wf-field--disabled', !!host.disabled);

  // Label
  const label = host.querySelector('[slot="label"]');
  if (label && !label.classList.contains('wf-field__label')) {
    label.classList.add('wf-field__label');
    const isRequired = host.validators?.some(
      (v: any) => v.constructor?.validatorName === 'Required',
    );
    label.classList.toggle('wf-field__label--required', !!isRequired);
  }

  // Native input
  if (
    host._inputNode &&
    !host._inputNode.classList.contains('wf-field__input')
  ) {
    host._inputNode.classList.add('wf-field__input');
  }

  // Help text
  const helpText = host.querySelector('[slot="help-text"]');
  if (helpText && !helpText.classList.contains('wf-field__help-text')) {
    helpText.classList.add('wf-field__help-text');
  }

  // Feedback
  const feedback = host.querySelector('lion-validation-feedback');
  if (feedback && !feedback.classList.contains('wf-field__feedback')) {
    feedback.classList.add('wf-field__feedback');
  }
}
