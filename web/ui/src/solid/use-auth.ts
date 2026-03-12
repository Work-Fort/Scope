import { createSignal, onCleanup } from 'solid-js';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

export function useAuth() {
  const client = getAuthClient();
  const [user, setUser] = createSignal<User | null>(client.getUser());
  const isAuthenticated = () => user() !== null;

  const onChange = (u: User | null) => setUser(u);
  const onLogout = () => setUser(null);
  client.on('change', onChange);
  client.on('logout', onLogout);
  onCleanup(() => { client.off('change', onChange); client.off('logout', onLogout); });

  return { user, isAuthenticated };
}
