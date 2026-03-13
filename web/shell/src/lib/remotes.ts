import { init, registerRemotes, loadRemote } from '@module-federation/runtime';
import type { ServiceInfo } from './api';

let initialized = false;

// Bootstrap the MF runtime once (required before registerRemotes).
init({ name: 'shell', remotes: [] });

export function initRemotes(services: ServiceInfo[]): void {
  if (initialized) return;

  const remotes = services
    .filter((s) => s.enabled && s.ui)
    .map((s) => ({
      name: s.name,
      entry: `/api/${s.name}/ui/remoteEntry.js`,
    }));

  if (remotes.length > 0) {
    registerRemotes(remotes);
  }

  initialized = true;
}

export interface ServiceModule {
  default: () => any;
  manifest: { name: string; label: string; route: string; minWidth?: number };
  SidebarContent?: () => any;
  HeaderActions?: () => any;
}

export async function loadServiceModule(
  serviceName: string,
): Promise<ServiceModule> {
  const mod = await loadRemote<ServiceModule>(`${serviceName}/index`);
  if (!mod || !mod.default || !mod.manifest) {
    throw new Error(
      `Remote "${serviceName}" did not export required fields (default, manifest)`,
    );
  }
  return mod;
}
