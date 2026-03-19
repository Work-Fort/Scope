export interface FortInfo {
  name: string;
  local: boolean;
  pylon?: string;
}

export interface ServiceInfo {
  name: string;
  label: string;
  route: string;
  enabled: boolean;
  ui: boolean;
  connected: boolean;
  setup_mode?: boolean;
  display?: 'nav' | 'menu';
  base_url?: string;
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

export async function fetchForts(): Promise<FortInfo[]> {
  const res = await fetch('/api/forts');
  if (!res.ok) throw new Error(`/api/forts: ${res.status}`);
  return res.json();
}

export async function fetchServices(fort: string): Promise<ServicesResponse> {
  const res = await fetch(`/forts/${fort}/api/services`);
  if (!res.ok) throw new Error(`/forts/${fort}/api/services: ${res.status}`);
  return res.json();
}

export async function fetchConfig(fort: string): Promise<ConfigResponse> {
  const res = await fetch(`/forts/${fort}/api/config`);
  if (!res.ok) throw new Error(`/forts/${fort}/api/config: ${res.status}`);
  return res.json();
}

export async function checkFortServices(fort: string): Promise<ServiceInfo[]> {
  try {
    const res = await fetchServices(fort);
    return res.services;
  } catch {
    return [];
  }
}

export async function checkSession(fort: string): Promise<boolean> {
  try {
    const res = await fetch(`/forts/${fort}/api/session`);
    if (!res.ok) return false;
    const body = await res.json();
    return body.authenticated === true;
  } catch {
    return false;
  }
}
