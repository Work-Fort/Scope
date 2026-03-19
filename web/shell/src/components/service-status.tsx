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
            {(svc) => (
              <tr>
                <td>
                  <strong>{svc.label}</strong>
                  <br />
                  <small class="service-status__name">{svc.name}</small>
                </td>
                <td>
                  <span class="service-status__indicator">
                    <wf-status-dot status={svc.connected ? 'online' : 'offline'} />
                    {svc.connected ? 'Connected' : 'Disconnected'}
                  </span>
                </td>
                <td>
                  <span class="service-status__indicator">
                    <wf-status-dot status={svc.ui ? 'online' : 'away'} />
                    {svc.ui ? 'Available' : 'No UI'}
                  </span>
                </td>
                <td>
                  <span class="service-status__indicator">
                    <wf-status-dot status={svc.base_url?.startsWith('https://') ? 'online' : 'away'} />
                    {svc.base_url?.startsWith('https://') ? 'HTTPS' : 'HTTP'}
                  </span>
                </td>
                <td><code>{svc.route}</code></td>
                <td>{svc.display ?? 'nav'}</td>
              </tr>
            )}
          </For>
        </tbody>
      </table>
    </div>
  );
};

export default ServiceStatus;
