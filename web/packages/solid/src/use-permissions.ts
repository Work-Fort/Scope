import { createSignal } from 'solid-js';

export function usePermissions(initial: string[] = []) {
  const [permsSet, setPermsSet] = createSignal<Set<string>>(new Set(initial));

  function can(permission: string): boolean {
    return permsSet().has(permission);
  }

  function update(perms: string[]): void {
    setPermsSet(new Set(perms));
  }

  const permissions = () => [...permsSet()];

  return { can, update, permissions };
}
