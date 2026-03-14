import { ref, onMounted, onUnmounted, type Ref } from 'vue';

export function useFormValidation(elementRef: Ref<HTMLElement | null>) {
  const dirty = ref(false);
  const touched = ref(false);
  const submitted = ref(false);
  const hasFeedbackFor = ref<string[]>([]);
  const modelValue = ref<unknown>(undefined);

  const update = () => {
    const el = elementRef.value as any;
    if (!el) return;
    dirty.value = el.dirty ?? false;
    touched.value = el.touched ?? false;
    submitted.value = el.submitted ?? false;
    hasFeedbackFor.value = el.hasFeedbackFor ?? [];
    modelValue.value = el.modelValue;
  };

  let cleanup: (() => void) | null = null;

  onMounted(() => {
    const el = elementRef.value;
    if (!el) return;
    el.addEventListener('model-value-changed', update);
    el.addEventListener('focusout', update);
    cleanup = () => {
      el.removeEventListener('model-value-changed', update);
      el.removeEventListener('focusout', update);
    };
  });

  onUnmounted(() => cleanup?.());

  return { dirty, touched, submitted, hasFeedbackFor, modelValue };
}
