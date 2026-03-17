import { ref } from 'vue';
import { PermissionSet } from '@workfort/ui';

export function usePermissions(initial: string[] = []) {
  const core = new PermissionSet(initial);
  const version = ref(0);

  function can(permission: string): boolean {
    version.value; // track reactivity
    return core.can(permission);
  }

  function update(perms: string[]) {
    core.update(perms);
    version.value++;
  }

  return { can, update };
}
