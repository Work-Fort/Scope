import { createSignal } from 'solid-js';

type Theme = 'dark' | 'light';
type Handedness = 'left' | 'right';

const STORAGE_KEY = 'wf-theme';
const HANDEDNESS_KEY = 'wf-handedness';

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

// --- Handedness preference ---

const [handedness, setHandedness] = createSignal<Handedness>(
  (typeof window !== 'undefined'
    ? (localStorage.getItem(HANDEDNESS_KEY) as Handedness)
    : null) || 'right',
);

export function toggleHandedness(): void {
  const next = handedness() === 'right' ? 'left' : 'right';
  setHandedness(next);
  localStorage.setItem(HANDEDNESS_KEY, next);
}

export { handedness };

// --- Mobile sidebar toggle ---
const [sidebarOpen, setSidebarOpen] = createSignal(false);
export function toggleSidebar(): void {
  setSidebarOpen(!sidebarOpen());
}
export function closeSidebar(): void {
  setSidebarOpen(false);
}
export { sidebarOpen };

// Apply saved theme on load (overrides the static "dark" in index.html if user chose light).
applyTheme(getInitialTheme());
