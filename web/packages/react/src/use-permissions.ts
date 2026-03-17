import { useState, useCallback } from 'react';
import { PermissionSet } from '@workfort/ui';

export function usePermissions(initial: string[] = []) {
  const [core] = useState(() => new PermissionSet(initial));
  const [, setVersion] = useState(0); // trigger re-render on update

  const can = useCallback((permission: string) => core.can(permission), [core]);

  const update = useCallback((perms: string[]) => {
    core.update(perms);
    setVersion((v) => v + 1);
  }, [core]);

  return { can, update };
}
