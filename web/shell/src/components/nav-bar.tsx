import { For, type Component } from 'solid-js';
import { useNavigate, useLocation, useParams } from '@solidjs/router';
import { services, fortName, clearAuthRequired } from '../stores/services';
import { toggleTheme, toggleHandedness, handedness } from '../stores/theme';
import { useTheme } from '@workfort/ui-solid';

const NavBar: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams<{ fort: string }>();
  const theme = useTheme();

  const visibleServices = () => services().filter((s) => s.enabled && s.ui);

  async function handleLogout() {
    // Clear the session cookie via the auth proxy.
    await fetch(`/forts/${params.fort}/api/auth/v1/sign-out`, { method: 'POST' }).catch(() => {});
    // Force sign-in on next load.
    window.location.reload();
  }

  return (
    <wf-nav-bar hamburger-position={handedness() === 'left' ? 'top-left' : 'top-right'}>
      <span slot="brand" class="shell-nav__brand">{fortName() || 'WorkFort'}</span>

      <For each={visibleServices()}>
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

      <div slot="menu">
        <wf-list>
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
