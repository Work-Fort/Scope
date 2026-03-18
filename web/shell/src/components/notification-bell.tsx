import { Show, For, type Component } from 'solid-js';
import { notifications, unreadCount, markAllRead, notificationPanelOpen, toggleNotificationPanel, closeNotificationPanel } from '../stores/notifications';
import { closeSidebar } from '../stores/theme';
import { useNavigate, useParams } from '@solidjs/router';

const NotificationBell: Component = () => {
  const navigate = useNavigate();
  const params = useParams<{ fort: string }>();

  const handleClick = (route?: string) => {
    if (route && params.fort) {
      navigate(`/forts/${params.fort}${route}`);
    }
    closeNotificationPanel();
  };

  const toggleDropdown = () => {
    toggleNotificationPanel();
    if (notificationPanelOpen()) {
      closeSidebar();
      markAllRead();
    }
  };

  return (
    <div class="notification-bell-wrapper">
      <button
        class="notification-bell"
        onClick={toggleDropdown}
        aria-label="Notifications"
        aria-expanded={notificationPanelOpen()}
      >
        <svg class="notification-bell__icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
        <Show when={unreadCount() > 0}>
          <span class="notification-badge">
            {unreadCount() > 99 ? '99+' : unreadCount()}
          </span>
        </Show>
      </button>

      <Show when={notificationPanelOpen()}>
        <div class="notification-dropdown">
          <div class="notification-dropdown__header">Notifications</div>
          <Show
            when={notifications().length > 0}
            fallback={<div class="notification-empty">No notifications</div>}
          >
            <For each={notifications().slice(0, 20)}>
              {(n) => (
                <button
                  class="notification-item"
                  classList={{ 'notification-item--unread': !n.read }}
                  onClick={() => handleClick(n.route)}
                >
                  <div class="notification-item__title">{n.title}</div>
                  <Show when={n.body}>
                    <div class="notification-item__body">{n.body}</div>
                  </Show>
                  <div class="notification-item__meta">{n.service}</div>
                </button>
              )}
            </For>
          </Show>
        </div>
      </Show>
    </div>
  );
};

export default NotificationBell;
