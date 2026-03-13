type Theme = 'dark' | 'light';

const STORAGE_KEY = 'wf-theme';

function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';
  return (localStorage.getItem(STORAGE_KEY) as Theme) ?? 'dark';
}

export function applyTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem(STORAGE_KEY, theme);
}

export function toggleTheme(): void {
  const current = (document.documentElement.getAttribute('data-theme') ?? 'dark') as Theme;
  applyTheme(current === 'dark' ? 'light' : 'dark');
}

// Apply saved theme on load (overrides the static "dark" in index.html if user chose light).
applyTheme(getInitialTheme());
