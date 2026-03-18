import { createSignal } from 'solid-js';

export interface ShellNotification {
  id: number;
  service: string;
  title: string;
  body?: string;
  urgency: 'passive' | 'active';
  route?: string;
  read: boolean;
  created_at: string;
}

const [notificationList, setNotificationList] = createSignal<ShellNotification[]>([]);
const [unreadTotal, setUnreadTotal] = createSignal(0);
const [panelOpen, setPanelOpen] = createSignal(false);

export const notificationPanelOpen = panelOpen;
export function openNotificationPanel(): void { setPanelOpen(true); }
export function closeNotificationPanel(): void { setPanelOpen(false); }
export function toggleNotificationPanel(): void { setPanelOpen(!panelOpen()); }

export const notifications = notificationList;
export const unreadCount = unreadTotal;

export function addNotification(n: ShellNotification): void {
  setNotificationList((prev) => [n, ...prev]);
  if (!n.read) {
    setUnreadTotal((c) => c + 1);
  }
}

export function markAllRead(): void {
  setUnreadTotal(0);
  setNotificationList((prev) => prev.map((n) => ({ ...n, read: true })));

  // Tell the server to mark all as read (fire-and-forget).
  const latest = notificationList()[0];
  if (latest?.id) {
    fetch('/api/notifications/read', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ up_to_id: latest.id }),
    }).catch(() => {});
  }
}

/** Fetch initial notifications from the scope-server API. */
export async function fetchNotifications(): Promise<void> {
  try {
    const resp = await fetch('/api/notifications?limit=20');
    if (resp.ok) {
      const data = await resp.json();
      setNotificationList(data.notifications ?? []);
      setUnreadTotal(data.unread ?? 0);
    }
  } catch {
    // Silently ignore — WS will deliver new ones.
  }
}
