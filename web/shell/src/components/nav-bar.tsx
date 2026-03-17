import { For, type Component } from 'solid-js';
import { useNavigate, useLocation, useParams } from '@solidjs/router';
import { services, fortName } from '../stores/services';
import { toggleTheme, toggleHandedness, handedness } from '../stores/theme';
import { useTheme } from '@workfort/ui-solid';

const NavBar: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const params = useParams<{ fort: string }>();
  const theme = useTheme();

  // Only show services that have a UI.
  const visibleServices = () => services().filter((s) => s.enabled && s.ui);

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

      <div slot="actions">
        <wf-button variant="text" on:wf-click={() => toggleHandedness()}>
          {handedness() === 'right' ? '\u21E4 Left' : 'Right \u21E5'}
        </wf-button>
        <wf-button variant="text" on:wf-click={() => toggleTheme()}>
          {theme() === 'dark' ? 'Light' : 'Dark'}
        </wf-button>
      </div>
    </wf-nav-bar>
  );
};

export default NavBar;
