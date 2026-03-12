import { describe, it, expect, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { _resetAuthClient, getAuthClient } from '../../src/auth/index.js';
import { auth } from '../../src/svelte/auth.js';

const MOCK_USER = { id: '1', username: 'kazw', name: 'Kaz Walker', displayName: 'Kaz', type: 'user' as const };
const MOCK_SESSION = { id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z' };

describe('auth store (Svelte)', () => {
  afterEach(() => { _resetAuthClient(); vi.restoreAllMocks(); });

  it('provides reactive user via store', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    expect(get(auth.user)).toBeNull();
    await getAuthClient().init();
    expect(get(auth.user)).toEqual(MOCK_USER);
    expect(get(auth.isAuthenticated)).toBe(true);
  });
});
