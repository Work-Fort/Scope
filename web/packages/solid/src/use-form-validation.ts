import { createSignal, onMount, onCleanup } from 'solid-js';

export function useFormValidation(getEl: () => HTMLElement | null) {
  const [dirty, setDirty] = createSignal(false);
  const [touched, setTouched] = createSignal(false);
  const [submitted, setSubmitted] = createSignal(false);
  const [hasFeedbackFor, setHasFeedbackFor] = createSignal<string[]>([]);
  const [modelValue, setModelValue] = createSignal<unknown>(undefined);

  const update = () => {
    const el = getEl() as any;
    if (!el) return;
    setDirty(el.dirty ?? false);
    setTouched(el.touched ?? false);
    setSubmitted(el.submitted ?? false);
    setHasFeedbackFor(el.hasFeedbackFor ?? []);
    setModelValue(el.modelValue);
  };

  onMount(() => {
    const el = getEl();
    if (!el) return;
    el.addEventListener('model-value-changed', update);
    el.addEventListener('focusout', update);
  });

  onCleanup(() => {
    const el = getEl();
    if (!el) return;
    el.removeEventListener('model-value-changed', update);
    el.removeEventListener('focusout', update);
  });

  return { dirty, touched, submitted, hasFeedbackFor, modelValue };
}
