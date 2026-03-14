import { createSignal } from 'solid-js';

export interface ToastEntry {
  id: string;
  variant: 'error' | 'warning' | 'info' | 'success';
  message: string;
  sticky: boolean;
  duration: number;
}

let nextId = 0;
const [toasts, setToasts] = createSignal<ToastEntry[]>([]);

export { toasts };

export interface ToastOptions {
  sticky?: boolean;
  duration?: number;
}

export function addToast(
  variant: ToastEntry['variant'],
  message: string,
  options: ToastOptions = {},
): string {
  const id = `toast-${++nextId}`;
  const entry: ToastEntry = {
    id,
    variant,
    message,
    sticky: options.sticky ?? false,
    duration: options.duration ?? 5000,
  };
  setToasts((prev) => [...prev, entry]);
  return id;
}

export function dismissToast(id: string): void {
  setToasts((prev) => prev.filter((t) => t.id !== id));
}
