import { ref, readonly, onUnmounted } from 'vue';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

export function useAuth() {
  const client = getAuthClient();
  const user = ref<User | null>(client.getUser());
  const isAuthenticated = ref(client.isAuthenticated);

  const onChange = (u: User | null) => { user.value = u; isAuthenticated.value = u !== null; };
  const onLogout = () => { user.value = null; isAuthenticated.value = false; };

  client.on('change', onChange);
  client.on('logout', onLogout);

  try { onUnmounted(() => { client.off('change', onChange); client.off('logout', onLogout); }); }
  catch { /* not in component setup */ }

  return { user: readonly(user), isAuthenticated: readonly(isAuthenticated) };
}
