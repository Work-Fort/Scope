import { readable } from 'svelte/store';

type Theme = 'dark' | 'light';

function getCurrentTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
}

export const theme = readable<Theme>(getCurrentTheme(), (set) => {
  if (typeof document === 'undefined') return;
  const obs = new MutationObserver(() => set(getCurrentTheme()));
  obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
  return () => obs.disconnect();
});
