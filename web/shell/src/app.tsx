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
import { services, startPolling, stopPolling, isSetupMode, isAuthRequired, isSessionChecked, clearAuthRequired, clearSetupMode } from './stores/services';
import { fetchForts } from './lib/api';
import './stores/theme';
import ShellLayout from './components/shell-layout';
import ServiceMount from './components/service-mount';
import Unavailable from './components/unavailable';
import FortPicker from './components/fort-picker';
import SetupForm from './components/setup-form';
import SignInForm from './components/sign-in-form';
import type { ServiceModule } from './lib/remotes';
import type { SidebarMount } from './components/shell-layout';

// Context to pass sidebar setter from FortShell to ServicePage.
const FortShellContext = createContext<{
  setSidebarMount: (v: SidebarMount | undefined) => void;
}>({ setSidebarMount: () => {} });

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
  const [sidebarMount, setSidebarMount] = createSignal<SidebarMount | undefined>();

  createEffect(() => {
    const fort = params.fort;
    fetchForts().then((forts) => {
      const info = forts.find((f) => f.name === fort);
      startPolling(fort, info);
    }).catch(() => {
      startPolling(fort);
    });
  });
  onCleanup(() => stopPolling());

  return (
    <FortShellContext.Provider value={{ setSidebarMount }}>
      <Show when={!isSetupMode()} fallback={
        <SetupForm fort={params.fort} onComplete={() => { clearSetupMode(); clearAuthRequired(); }} />
      }>
        <Show when={isSessionChecked()} fallback={
          <div style="display:flex;align-items:center;justify-content:center;height:100vh;color:var(--wf-color-text-secondary)">Loading…</div>
        }>
          <Show when={!isAuthRequired()} fallback={
            <SignInForm fort={params.fort} onComplete={() => { clearAuthRequired(); startPolling(params.fort); }} />
          }>
            <ShellLayout sidebar={sidebarMount()}>{props.children}</ShellLayout>
          </Show>
        </Show>
      </Show>
    </FortShellContext.Provider>
  );
};

const ServicePage: Component = () => {
  const params = useParams<{ fort: string; service: string }>();
  const ctx = useContext(FortShellContext);

  const handleModule = (mod: ServiceModule | null) => {
    ctx.setSidebarMount(
      mod?.mountSidebar && mod?.unmountSidebar
        ? { mount: mod.mountSidebar.bind(mod), unmount: mod.unmountSidebar.bind(mod) }
        : undefined,
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
