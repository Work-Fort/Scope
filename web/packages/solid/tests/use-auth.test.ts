import { describe, it, expect, afterEach, vi } from 'vitest';
import { createRoot } from 'solid-js';
import { _resetAuthClient, getAuthClient } from '@workfort/auth';
import { useAuth } from '../src/use-auth.js';

const MOCK_USER = { id: '1', username: 'kazw', name: 'Kaz Walker', displayName: 'Kaz', type: 'user' as const };
const MOCK_SESSION = { id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z' };

describe('useAuth (Solid)', () => {
  afterEach(() => { _resetAuthClient(); vi.restoreAllMocks(); });

  it('provides reactive signals', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    await createRoot(async (dispose) => {
      const { user, isAuthenticated } = useAuth();
      expect(user()).toBeNull();
      await getAuthClient().init();
      expect(user()).toEqual(MOCK_USER);
      expect(isAuthenticated()).toBe(true);
      dispose();
    });
  });
});
