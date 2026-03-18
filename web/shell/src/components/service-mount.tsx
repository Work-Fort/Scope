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

  let mounted = false;

  // Mount once when the module loads. Use untrack for props.connected
  // so this effect only re-runs when the module changes, not on every
  // connected state change. The mounted app handles prop updates internally.
  createEffect(() => {
    const m = mod();
    if (m && containerRef && !mounted) {
      const connected = untrack(() => props.connected);
      m.mount(containerRef, { connected });
      mounted = true;
    }
  });

  // Cleanup on unmount
  onCleanup(() => {
    const m = mod();
    if (m && containerRef && mounted) {
      m.unmount(containerRef);
      mounted = false;
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
