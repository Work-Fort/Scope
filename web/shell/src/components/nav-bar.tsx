import { For, type Component } from 'solid-js';
import { useNavigate, useLocation } from '@solidjs/router';
import { services } from '../stores/services';
import { toggleTheme } from '../stores/theme';
import { useTheme } from '@workfort/ui-solid';

const NavBar: Component = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();

  return (
    <nav class="shell-nav">
      <span class="shell-nav__brand">WorkFort</span>
      <wf-list class="shell-nav__tabs">
        <For each={services().filter((s) => s.enabled)}>
          {(svc) => (
            <wf-list-item
              active={location.pathname.startsWith(svc.route)}
              class={!svc.ui ? 'shell-nav__tab--disabled' : ''}
              on:wf-select={() => navigate(svc.route)}
            >
              {svc.label}
            </wf-list-item>
          )}
        </For>
      </wf-list>
      <div class="shell-nav__spacer" />
      <wf-button variant="text" on:wf-click={() => toggleTheme()}>
        {theme() === 'dark' ? 'Light' : 'Dark'}
      </wf-button>
    </nav>
  );
};

export default NavBar;
