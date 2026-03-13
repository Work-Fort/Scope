import { type Component, type JSX, Show } from 'solid-js';
import NavBar from './nav-bar';

const ShellLayout: Component<{
  sidebar?: () => JSX.Element;
  children: JSX.Element;
}> = (props) => {
  return (
    <div
      class="shell-layout"
      classList={{ 'shell-layout--no-sidebar': !props.sidebar }}
    >
      <NavBar />
      <Show when={props.sidebar}>
        <aside class="shell-sidebar">{props.sidebar!()}</aside>
      </Show>
      <main class="shell-content">{props.children}</main>
    </div>
  );
};

export default ShellLayout;
