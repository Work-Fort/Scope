import { useSyncExternalStore, useCallback } from 'react';

type Theme = 'dark' | 'light';

function getTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
}

export function useTheme(): Theme {
  const subscribe = useCallback((cb: () => void) => {
    if (typeof document === 'undefined') return () => {};
    const obs = new MutationObserver(() => cb());
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    return () => obs.disconnect();
  }, []);
  return useSyncExternalStore(subscribe, getTheme, () => 'dark' as Theme);
}
