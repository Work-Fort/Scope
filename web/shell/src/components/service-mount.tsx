import { createResource, Suspense, ErrorBoundary, type Component } from 'solid-js';
import { Dynamic } from 'solid-js/web';
import { loadServiceModule, type ServiceModule } from '../lib/remotes';
import Unavailable from './unavailable';

const ServiceMount: Component<{
  name: string;
  label: string;
  onModule?: (mod: ServiceModule | null) => void;
}> = (props) => {
  const [mod] = createResource(
    () => props.name,
    async (name) => {
      const m = await loadServiceModule(name);
      props.onModule?.(m);
      return m;
    },
  );

  return (
    <ErrorBoundary fallback={<Unavailable label={props.label} />}>
      <Suspense fallback={<wf-skeleton width="100%" height="200px" />}>
        {mod() && <Dynamic component={mod()!.default} />}
      </Suspense>
    </ErrorBoundary>
  );
};

export default ServiceMount;
