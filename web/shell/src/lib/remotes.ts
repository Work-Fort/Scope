import { init, registerRemotes, loadRemote } from '@module-federation/runtime';
import type { ServiceInfo } from './api';

// Bootstrap MF runtime once.
init({ name: 'shell', remotes: [] });

const registeredNames = new Set<string>();

export function registerNewRemotes(fort: string, services: ServiceInfo[]): void {
  const newRemotes = services
    .filter((s) => s.enabled && s.ui && !registeredNames.has(s.name))
    .map((s) => ({
      name: s.name,
      entry: `/forts/${fort}/api/${s.name}/ui/remoteEntry.js`,
      type: 'module' as const,
    }));

  if (newRemotes.length > 0) {
    registerRemotes(newRemotes);
    newRemotes.forEach((r) => registeredNames.add(r.name));
  }
}

export interface ServiceModule {
  default: (props: { connected: boolean }) => any;
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
