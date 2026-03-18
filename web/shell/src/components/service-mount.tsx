import { createResource, Suspense, ErrorBoundary, Show, onCleanup, createEffect, untrack, type Component } from 'solid-js';
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

  let currentMod: ServiceModule | null = null;

  // Mount/unmount when the module changes (triggered by props.name changing).
  // Unmounts the previous service before mounting the new one.
  createEffect(() => {
    const m = mod();

    // Unmount previous if different
    if (currentMod && currentMod !== m && containerRef) {
      currentMod.unmount(containerRef);
      containerRef.innerHTML = '';
      currentMod = null;
    }

    // Mount new
    if (m && containerRef && currentMod !== m) {
      const connected = untrack(() => props.connected);
      m.mount(containerRef, { connected });
      currentMod = m;
    }
  });

  // Cleanup on component unmount
  onCleanup(() => {
    if (currentMod && containerRef) {
      currentMod.unmount(containerRef);
      currentMod = null;
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
