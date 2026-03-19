import { createResource, createSignal, createEffect, Show, For, type Component } from 'solid-js';
import { Navigate, useNavigate } from '@solidjs/router';
import { fetchForts, checkFortServices } from '../lib/api';

const FortPicker: Component = () => {
  const [forts] = createResource(fetchForts);
  const [httpWarnings, setHttpWarnings] = createSignal<Record<string, boolean>>({});
  const [checking, setChecking] = createSignal(false);
  const navigate = useNavigate();

  createEffect(() => {
    const fortList = forts();
    if (!fortList) return;

    const pylonForts = fortList.filter((f) => f.pylon);
    if (pylonForts.length === 0) return;

    setChecking(true);
    Promise.all(
      pylonForts.map(async (f) => {
        const services = await checkFortServices(f.name);
        const hasHttp = services.some((s) => s.base_url?.startsWith('http://'));
        return [f.name, hasHttp] as const;
      }),
    ).then((results) => {
      const warnings: Record<string, boolean> = {};
      for (const [name, hasHttp] of results) {
        if (hasHttp) warnings[name] = true;
      }
      setHttpWarnings(warnings);
      setChecking(false);
    });
  });

  return (
    <Show when={!forts.loading && !checking()} fallback={<wf-skeleton width="100%" height="200px" />}>
      <Show
        when={forts() && forts()!.length !== 1}
        fallback={
          forts() && forts()!.length === 1
            ? <Navigate href={`/forts/${forts()![0].name}`} />
            : <div class="shell-unavailable">No forts configured.</div>
        }
      >
        <div class="fort-picker">
          <h2 class="fort-picker__title">Select a Fort</h2>
          <wf-list>
            <For each={forts()}>
              {(fort) => (
                <wf-list-item on:wf-select={() => navigate(`/forts/${fort.name}`)}>
                  {fort.name}
                  <Show when={httpWarnings()[fort.name]}>
                    <wf-badge color="yellow">HTTP</wf-badge>
                  </Show>
                </wf-list-item>
              )}
            </For>
          </wf-list>
        </div>
      </Show>
    </Show>
  );
};

export default FortPicker;
