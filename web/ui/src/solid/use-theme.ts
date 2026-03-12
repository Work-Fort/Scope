import { createSignal, onCleanup } from 'solid-js';

type Theme = 'dark' | 'light';

function getCurrentTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
}

export function useTheme() {
  const [theme, setTheme] = createSignal<Theme>(getCurrentTheme());
  if (typeof document !== 'undefined') {
    const obs = new MutationObserver(() => setTheme(getCurrentTheme()));
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    onCleanup(() => obs.disconnect());
  }
  return theme;
}
