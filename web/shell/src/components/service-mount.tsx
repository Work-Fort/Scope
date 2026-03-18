import { createResource, Suspense, ErrorBoundary, Show, onCleanup, createEffect, type Component } from 'solid-js';
import { loadServiceModule, type ServiceModule } from '../lib/remotes';
import Unavailable from './unavailable';

const ServiceMount: Component<{
  name: string;
  label: string;
  connected: boolean;
  onModule?: (mod: ServiceModule | null) => void;
}> = (props) => {
  let containerRef!: HTMLDivElement;

  const [mod] = createResource(
    () => props.name,
    async (name) => {
      const m = await loadServiceModule(name);
      props.onModule?.(m);
      return m;
    },
  );

  // Mount/unmount the service when the module loads
  createEffect(() => {
    const m = mod();
    if (m && containerRef) {
      m.mount(containerRef, { connected: props.connected });
    }
  });

  // Cleanup on unmount
  onCleanup(() => {
    const m = mod();
    if (m && containerRef) {
      m.unmount(containerRef);
    }
  });

  return (
    <ErrorBoundary fallback={<Unavailable label={props.label} />}>
      <Suspense fallback={<wf-skeleton width="100%" height="200px" />}>
        <Show
          when={mod() || props.connected}
          fallback={
            <wf-banner
              variant="warning"
              headline={`${props.label} is starting up or temporarily unavailable. This page will update automatically when it's ready.`}
            />
          }
        >
          <div ref={containerRef} style="display:contents" />
        </Show>
      </Suspense>
    </ErrorBoundary>
  );
};

export default ServiceMount;
