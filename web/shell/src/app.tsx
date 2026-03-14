import {
  type Component,
  Show,
  createContext,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  useContext,
} from 'solid-js';
import { Navigate, Route, Router, useParams } from '@solidjs/router';
import { services, startPolling, stopPolling } from './stores/services';
import './stores/theme';
import ShellLayout from './components/shell-layout';
import ServiceMount from './components/service-mount';
import Unavailable from './components/unavailable';
import FortPicker from './components/fort-picker';
import type { ServiceModule } from './lib/remotes';

// Context to pass sidebar setter from FortShell to ServicePage.
const FortShellContext = createContext<{
  setSidebarComponent: (v: (() => any) | undefined) => void;
}>({ setSidebarComponent: () => {} });

const App: Component = () => {
  return (
    <Router>
      <Route path="/" component={FortPicker} />
      <Route path="/forts/:fort" component={FortShell}>
        <Route path="/:service/*rest" component={ServicePage} />
        <Route path="/" component={FortIndex} />
      </Route>
    </Router>
  );
};

const FortShell: Component = (props: { children?: any }) => {
  const params = useParams<{ fort: string }>();
  const [sidebarComponent, setSidebarComponent] = createSignal<(() => any) | undefined>();

  createEffect(() => {
    const fort = params.fort;
    startPolling(fort);
  });
  onCleanup(() => stopPolling());

  return (
    <FortShellContext.Provider value={{ setSidebarComponent }}>
      <ShellLayout sidebar={sidebarComponent()}>{props.children}</ShellLayout>
    </FortShellContext.Provider>
  );
};

const ServicePage: Component = () => {
  const params = useParams<{ fort: string; service: string }>();
  const ctx = useContext(FortShellContext);

  const handleModule = (mod: ServiceModule | null) => {
    ctx.setSidebarComponent(
      mod?.SidebarContent ? () => mod.SidebarContent! : undefined,
    );
  };

  const svc = createMemo(() =>
    services().find((s) => s.enabled && s.route === `/${params.service}`),
  );

  return (
    <>
      {svc() ? (
        svc()!.ui ? (
          <ServiceMount name={svc()!.name} label={svc()!.label} connected={svc()!.connected} onModule={handleModule} />
        ) : (
          <Unavailable label={svc()!.label} />
        )
      ) : (
        <Navigate href={`/forts/${params.fort}`} />
      )}
    </>
  );
};

const FortIndex: Component = () => {
  const params = useParams<{ fort: string }>();
  const firstRoute = createMemo(() => {
    const enabled = services().find((s) => s.enabled);
    return enabled ? `/forts/${params.fort}${enabled.route}` : null;
  });

  return (
    <Show when={firstRoute()} fallback={<div class="shell-unavailable">No services available.</div>}>
      <Navigate href={firstRoute()!} />
    </Show>
  );
};

export default App;
