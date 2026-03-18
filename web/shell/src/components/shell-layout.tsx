import { type Component, type JSX, Show, For, onCleanup, createEffect } from 'solid-js';
import NavBar from './nav-bar';
import { sortedBanners, dismissBanner } from '../stores/banners';
import { toasts, dismissToast } from '../stores/toasts';

export interface SidebarMount {
  mount(el: HTMLElement): void;
  unmount(el: HTMLElement): void;
}

const ShellLayout: Component<{
  sidebar?: SidebarMount;
  children: JSX.Element;
}> = (props) => {
  let sidebarRef!: HTMLDivElement;

  createEffect(() => {
    const sb = props.sidebar;
    if (sb && sidebarRef) {
      sb.mount(sidebarRef);
    }
  });

  onCleanup(() => {
    const sb = props.sidebar;
    if (sb && sidebarRef) {
      sb.unmount(sidebarRef);
    }
  });

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
        <aside class="shell-sidebar">
          <div ref={sidebarRef} style="display:contents" />
        </aside>
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
