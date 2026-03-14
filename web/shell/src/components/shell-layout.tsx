import { type Component, type JSX, Show, For } from 'solid-js';
import NavBar from './nav-bar';
import { sortedBanners, dismissBanner } from '../stores/banners';
import { toasts, dismissToast } from '../stores/toasts';

const ShellLayout: Component<{
  sidebar?: () => JSX.Element;
  children: JSX.Element;
}> = (props) => {
  return (
    <div
      class="shell-layout"
      classList={{ 'shell-layout--no-sidebar': !props.sidebar }}
    >
      <div class="shell-banners">
        <For each={sortedBanners()}>
          {(banner) => (
            <wf-banner
              variant={banner.variant}
              headline={banner.headline}
              details={banner.details ?? ''}
              dismissible
              on:wf-dismiss={() => dismissBanner(banner.key)}
            />
          )}
        </For>
      </div>
      <NavBar />
      <Show when={props.sidebar}>
        <aside class="shell-sidebar">{props.sidebar!()}</aside>
      </Show>
      <main class="shell-content">{props.children}</main>
      <wf-toast-container position="top-right">
        <For each={toasts()}>
          {(toast) => (
            <wf-toast
              variant={toast.variant}
              sticky={toast.sticky}
              duration={toast.duration}
              on:wf-dismiss={() => dismissToast(toast.id)}
            >
              {toast.message}
            </wf-toast>
          )}
        </For>
      </wf-toast-container>
    </div>
  );
};

export default ShellLayout;
