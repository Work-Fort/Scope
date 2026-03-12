import { readable, derived } from 'svelte/store';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

// getAuthClient() is called lazily inside the readable's start function,
// not at module scope. This avoids side-effecting the singleton on import
// and stays consistent with how React/Vue/Solid adapters call it inside
// their hook function bodies.
const user = readable<User | null>(null, (set) => {
  const client = getAuthClient();
  set(client.getUser());
  const onChange = (u: User | null) => set(u);
  const onLogout = () => set(null);
  client.on('change', onChange);
  client.on('logout', onLogout);
  return () => { client.off('change', onChange); client.off('logout', onLogout); };
});

const isAuthenticated = derived(user, ($user) => $user !== null);

export const auth = { user, isAuthenticated };
