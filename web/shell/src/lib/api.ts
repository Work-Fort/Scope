export interface ServiceInfo {
  name: string;
  label: string;
  route: string;
  enabled: boolean;
  ui: boolean;
  connected: boolean;
}

export interface Conflict {
  url: string;
  name: string;
  reason: string;
}

export interface ServicesResponse {
  fort: string;
  services: ServiceInfo[];
  conflicts: Conflict[];
}

export interface ConfigResponse {
  fort: string;
}

export async function fetchServices(): Promise<ServicesResponse> {
  const res = await fetch('/api/services');
  if (!res.ok) throw new Error(`/api/services: ${res.status}`);
  return res.json();
}

export async function fetchConfig(): Promise<ConfigResponse> {
  const res = await fetch('/api/config');
  if (!res.ok) throw new Error(`/api/config: ${res.status}`);
  return res.json();
}
