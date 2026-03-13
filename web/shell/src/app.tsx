import { type Component, createSignal, createMemo } from 'solid-js';
import { Router, Route, Navigate, useParams } from '@solidjs/router';
import { services } from './stores/services';
import './stores/theme';
import ShellLayout from './components/shell-layout';
import ServiceMount from './components/service-mount';
import Unavailable from './components/unavailable';
import type { ServiceModule } from './lib/remotes';

const App: Component = () => {
  const [sidebarComponent, setSidebarComponent] = createSignal<(() => any) | undefined>();

  const handleModule = (mod: ServiceModule | null) => {
    setSidebarComponent(
      // Wrap in arrow to avoid Solid interpreting the function as a setter callback
      mod?.SidebarContent ? () => mod.SidebarContent! : () => undefined,
    );
  };

  const firstRoute = createMemo(() => {
    const enabled = services().find((s) => s.enabled);
    return enabled?.route ?? '/';
  });

  const ServicePage: Component = () => {
    const params = useParams<{ service: string }>();
    const svc = createMemo(() =>
      services().find((s) => s.enabled && s.route === `/${params.service}`),
    );

    return (
      <>
        {svc() ? (
          svc()!.ui ? (
            <ServiceMount name={svc()!.name} label={svc()!.label} onModule={handleModule} />
          ) : (
            <Unavailable label={svc()!.label} />
          )
        ) : (
          <Navigate href="/" />
        )}
      </>
    );
  };

  return (
    <Router root={(props) => <ShellLayout sidebar={sidebarComponent()}>{props.children}</ShellLayout>}>
      <Route path="/:service/*rest" component={ServicePage} />
      <Route path="/" component={() => <Navigate href={firstRoute()} />} />
    </Router>
  );
};

export default App;
