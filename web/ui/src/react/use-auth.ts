import { useSyncExternalStore, useCallback } from 'react';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

export function useAuth(): { user: User | null; isAuthenticated: boolean } {
  const client = getAuthClient();
  const subscribe = useCallback((cb: () => void) => {
    client.on('change', cb);
    client.on('logout', cb);
    return () => { client.off('change', cb); client.off('logout', cb); };
  }, [client]);
  const getSnapshot = useCallback(() => client.getUser(), [client]);
  const user = useSyncExternalStore(subscribe, getSnapshot);
  return { user, isAuthenticated: user !== null };
}
