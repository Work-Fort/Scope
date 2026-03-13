import { ref, readonly, onUnmounted } from 'vue';

type Theme = 'dark' | 'light';

export function useTheme() {
  const theme = ref<Theme>(
    (typeof document !== 'undefined'
      ? (document.documentElement.getAttribute('data-theme') as Theme) : null) ?? 'dark',
  );
  let observer: MutationObserver | null = null;
  if (typeof document !== 'undefined') {
    observer = new MutationObserver(() => {
      theme.value = (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
  }
  try { onUnmounted(() => observer?.disconnect()); } catch { /* not in component setup */ }
  return readonly(theme);
}
