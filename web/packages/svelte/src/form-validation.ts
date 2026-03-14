import { writable } from 'svelte/store';

interface FormValidationState {
  dirty: boolean;
  touched: boolean;
  submitted: boolean;
  hasFeedbackFor: string[];
  modelValue: unknown;
}

export function createFormValidation(node: HTMLElement) {
  const { subscribe, set } = writable<FormValidationState>({
    dirty: false, touched: false, submitted: false,
    hasFeedbackFor: [], modelValue: undefined,
  });

  const update = () => {
    const el = node as any;
    set({
      dirty: el.dirty ?? false,
      touched: el.touched ?? false,
      submitted: el.submitted ?? false,
      hasFeedbackFor: el.hasFeedbackFor ?? [],
      modelValue: el.modelValue,
    });
  };

  node.addEventListener('model-value-changed', update);
  node.addEventListener('focusout', update);

  return {
    subscribe,
    destroy() {
      node.removeEventListener('model-value-changed', update);
      node.removeEventListener('focusout', update);
    },
  };
}
