import { createSignal } from 'solid-js';
import { PermissionSet } from '@workfort/ui';

export function usePermissions(initial: string[] = []) {
  const core = new PermissionSet(initial);
  const [permissions, setPermissions] = createSignal<string[]>(initial);

  function can(permission: string): boolean {
    return core.can(permission);
  }

  function update(perms: string[]): void {
    core.update(perms);
    setPermissions(perms);
  }

  return { can, update, permissions };
}
