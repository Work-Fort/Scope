import { type Component, For } from 'solid-js';
import { useParams } from '@solidjs/router';
import { services } from '../stores/services';

const ServiceStatus: Component = () => {
  const params = useParams<{ fort: string }>();

  return (
    <div class="service-status">
      <h2 class="service-status__title">Service Status</h2>
      <p class="service-status__subtitle">Fort: {params.fort}</p>
      <table class="service-status__table">
        <thead>
          <tr>
            <th>Service</th>
            <th>Status</th>
            <th>UI</th>
            <th>Protocol</th>
            <th>Route</th>
            <th>Display</th>
          </tr>
        </thead>
        <tbody>
          <For each={services()}>
            {(svc) => {
              const isHttps = svc.base_url?.startsWith('https://');
              return (
                <tr>
                  <td>
                    <strong>{svc.label}</strong>
                    <br />
                    <small class="service-status__name">{svc.name}</small>
                  </td>
                  <td>
                    <wf-badge color={svc.connected ? 'green' : 'red'}>
                      {svc.connected ? 'Connected' : 'Disconnected'}
                    </wf-badge>
                  </td>
                  <td>
                    <wf-badge color={svc.ui ? 'blue' : 'yellow'}>
                      {svc.ui ? 'Available' : 'No UI'}
                    </wf-badge>
                  </td>
                  <td>
                    <wf-badge color={isHttps ? 'green' : 'yellow'}>
                      {isHttps ? 'HTTPS' : 'HTTP'}
                    </wf-badge>
                  </td>
                  <td><code>{svc.route}</code></td>
                  <td>{svc.display ?? 'nav'}</td>
                </tr>
              );
            }}
          </For>
        </tbody>
      </table>
    </div>
  );
};

export default ServiceStatus;
