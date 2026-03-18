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

export interface ServiceManifest {
  name: string;
  label: string;
  route: string;
  display: 'nav' | 'menu';
}

export interface ServiceModule {
  mount(el: HTMLElement, props: { connected: boolean }): void;
  unmount(el: HTMLElement): void;
  manifest: ServiceManifest;
  mountSidebar?(el: HTMLElement): void;
  unmountSidebar?(el: HTMLElement): void;
}

export async function loadServiceModule(
  serviceName: string,
): Promise<ServiceModule> {
  const mod = await loadRemote<ServiceModule>(`${serviceName}/index`);
  if (!mod || !mod.mount || !mod.unmount || !mod.manifest) {
    throw new Error(
      `Remote "${serviceName}" must export mount, unmount, and manifest`,
    );
  }
  return mod;
}
