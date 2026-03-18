import { For, Show, type Component } from 'solid-js';
import { useNavigate, useLocation, useParams } from '@solidjs/router';
import { services, fortName, clearAuthRequired } from '../stores/services';
import { toggleTheme, toggleHandedness, handedness, toggleSidebar, sidebarOpen, closeSidebar } from '../stores/theme';
import { closeNotificationPanel } from '../stores/notifications';
import { useTheme } from '@workfort/ui-solid';
import NotificationBell from './notification-bell';

const NavBar: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams<{ fort: string }>();
  const theme = useTheme();

  /** Services that show as tabs in the nav bar. */
  const navServices = () =>
    services().filter((s) => s.enabled && s.ui && (s.display ?? 'nav') === 'nav');

  /** Services that show as links in the hamburger menu. */
  const menuServices = () =>
    services().filter((s) => s.enabled && s.ui && s.display === 'menu');

  async function handleLogout() {
    // Clear the session cookie via the auth proxy.
    await fetch(`/forts/${params.fort}/api/auth/v1/sign-out`, { method: 'POST' }).catch(() => {});
    // Force sign-in on next load.
    window.location.reload();
  }

  return (
    <wf-nav-bar hamburger-position={handedness() === 'left' ? 'top-left' : 'top-right'}>
      <span slot="brand" class="shell-nav__brand">{fortName() || 'WorkFort'}</span>

      <For each={navServices()}>
        {(svc) => (
          <wf-list-item
            active={location.pathname.includes(svc.route)}
            on:wf-select={() => navigate(`/forts/${params.fort}${svc.route}`)}
          >
            <wf-status-dot status={svc.connected ? 'online' : 'offline'} />
            {svc.label}
          </wf-list-item>
        )}
      </For>

      <span slot="controls">
        <button
          class="shell-sidebar-toggle"
          aria-label="Toggle sidebar"
          onClick={() => { closeNotificationPanel(); toggleSidebar(); }}
        >
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
            <rect x="1" y="1" width="16" height="16" rx="2" />
            <line x1="7" y1="1" x2="7" y2="17" />
          </svg>
        </button>
        <NotificationBell />
      </span>

      <div slot="menu">
        <wf-list>
          <For each={menuServices()}>
            {(svc) => (
              <wf-list-item on:wf-select={() => navigate(`/forts/${params.fort}${svc.route}`)}>
                {svc.label}
              </wf-list-item>
            )}
          </For>
          <Show when={menuServices().length > 0}>
            <wf-divider />
          </Show>
          <wf-list-item on:wf-select={() => toggleTheme()}>
            {theme() === 'dark' ? '☀ Light mode' : '☾ Dark mode'}
          </wf-list-item>
          <wf-list-item on:wf-select={() => toggleHandedness()}>
            {handedness() === 'right' ? '← Left-handed' : '→ Right-handed'}
          </wf-list-item>
          <wf-divider />
          <wf-list-item on:wf-select={handleLogout}>
            Sign out
          </wf-list-item>
        </wf-list>
      </div>
    </wf-nav-bar>
  );
};

export default NavBar;
