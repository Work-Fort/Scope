import { describe, it, expect, afterEach, vi } from 'vitest';
import { nextTick } from 'vue';
import { _resetAuthClient, getAuthClient } from '@workfort/auth';
import { useAuth } from '../src/use-auth.js';

const MOCK_USER = { id: '1', username: 'kazw', name: 'Kaz Walker', displayName: 'Kaz', type: 'user' as const };
const MOCK_SESSION = { id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z' };

describe('useAuth (Vue)', () => {
  afterEach(() => { _resetAuthClient(); vi.restoreAllMocks(); });

  it('returns reactive user after init', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    const { user, isAuthenticated } = useAuth();
    expect(user.value).toBeNull();
    await getAuthClient().init();
    await nextTick();
    expect(user.value).toEqual(MOCK_USER);
    expect(isAuthenticated.value).toBe(true);
  });
});
