import { createResource, Show, For, type Component } from 'solid-js';
import { Navigate, useNavigate } from '@solidjs/router';
import { fetchForts } from '../lib/api';

const FortPicker: Component = () => {
  const [forts] = createResource(fetchForts);
  const navigate = useNavigate();

  return (
    <Show when={!forts.loading} fallback={<wf-skeleton width="100%" height="200px" />}>
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
