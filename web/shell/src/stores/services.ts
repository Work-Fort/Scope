import { createResource } from 'solid-js';
import { fetchServices, type ServiceInfo, type ServicesResponse } from '../lib/api';
import { initRemotes } from '../lib/remotes';

const [data, { refetch }] = createResource<ServicesResponse>(async () => {
  const res = await fetchServices();
  initRemotes(res.services);
  return res;
});

export const servicesData = data;
export const refetchServices = refetch;

export function services(): ServiceInfo[] {
  return servicesData()?.services ?? [];
}

export function fortName(): string {
  return servicesData()?.fort ?? '';
}
