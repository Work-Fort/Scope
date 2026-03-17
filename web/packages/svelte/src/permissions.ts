import { writable } from 'svelte/store';
import { PermissionSet } from '@workfort/ui';

export function createPermissions(initial: string[] = []) {
  const core = new PermissionSet(initial);
  const store = writable(initial);

  function can(permission: string): boolean {
    return core.can(permission);
  }

  function update(perms: string[]) {
    core.update(perms);
    store.set(perms);
  }

  return { can, update, subscribe: store.subscribe };
}
