import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { AuthClient } from '../../src/auth/client.js';
import { AuthInitError } from '../../src/auth/types.js';

const MOCK_USER = {
  id: '1', username: 'kazw', name: 'Kaz Walker',
  displayName: 'Kaz', type: 'user' as const,
};
const MOCK_SESSION = {
  id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z',
};

function mockSessionResponse(status = 200) {
  return new Response(
    status === 200 ? JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }) : '',
    { status, headers: status === 200 ? { 'Content-Type': 'application/json' } : {} },
  );
}

describe('AuthClient', () => {
  let client: AuthClient;

  beforeEach(() => {
    client = new AuthClient();
    vi.restoreAllMocks();
  });

  afterEach(() => { client.destroy(); });

  describe('init()', () => {
    it('fetches session and stores user', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse());
      await client.init();
      expect(client.getUser()).toEqual(MOCK_USER);
      expect(client.getSession()).toEqual(MOCK_SESSION);
      expect(client.isAuthenticated).toBe(true);
    });

    it('sets null on 401 without throwing', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse(401));
      await client.init();
      expect(client.getUser()).toBeNull();
      expect(client.isAuthenticated).toBe(false);
    });

    it('throws AuthInitError on network error', async () => {
      vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new TypeError('Failed to fetch'));
      await expect(client.init()).rejects.toThrow(AuthInitError);
    });

    it('throws AuthInitError on 500', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse(500));
      await expect(client.init()).rejects.toThrow(AuthInitError);
    });

    it('emits change event with user', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse());
      const handler = vi.fn();
      client.on('change', handler);
      await client.init();
      expect(handler).toHaveBeenCalledWith(MOCK_USER);
    });
  });

  describe('refresh()', () => {
    it('re-fetches session', async () => {
      const spy = vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(mockSessionResponse())
        .mockResolvedValueOnce(new Response(
          JSON.stringify({ user: { ...MOCK_USER, displayName: 'K' }, session: MOCK_SESSION }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        ));
      await client.init();
      await client.refresh();
      expect(client.getUser()!.displayName).toBe('K');
      expect(spy).toHaveBeenCalledTimes(2);
    });

    it('emits logout on 401 during refresh', async () => {
      vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(mockSessionResponse())
        .mockResolvedValueOnce(mockSessionResponse(401));
      const logoutHandler = vi.fn();
      client.on('logout', logoutHandler);
      await client.init();
      await client.refresh();
      expect(client.isAuthenticated).toBe(false);
      expect(logoutHandler).toHaveBeenCalledOnce();
    });
  });

  describe('logout()', () => {
    it('clears state and emits logout', async () => {
      vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(mockSessionResponse())
        .mockResolvedValueOnce(new Response('', { status: 200 }));
      await client.init();
      const logoutHandler = vi.fn();
      client.on('logout', logoutHandler);
      await client.logout();
      expect(client.getUser()).toBeNull();
      expect(client.isAuthenticated).toBe(false);
      expect(logoutHandler).toHaveBeenCalledOnce();
    });
  });

  describe('events', () => {
    it('off() removes listener', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      const handler = vi.fn();
      client.on('change', handler);
      client.off('change', handler);
      await client.init();
      expect(handler).not.toHaveBeenCalled();
    });
  });

  describe('visibility change', () => {
    it('calls refresh when visible after >5 min hidden', async () => {
      vi.useFakeTimers();
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      await client.init();
      // Simulate going hidden
      Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      // Advance time past threshold
      vi.advanceTimersByTime(6 * 60 * 1000);
      // Re-spy fetch after init consumed the first mock
      const refreshSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      // Simulate becoming visible
      Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      // refresh() should have been called
      await vi.waitFor(() => expect(refreshSpy).toHaveBeenCalled());
      vi.useRealTimers();
    });

    it('does not call refresh when hidden <5 min', async () => {
      vi.useFakeTimers();
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      await client.init();
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      // Simulate going hidden then immediately visible (no time advance)
      Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      expect(fetchSpy).not.toHaveBeenCalled();
      vi.useRealTimers();
    });
  });
});
