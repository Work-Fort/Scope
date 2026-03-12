import { describe, it, expect, afterEach, vi } from 'vitest';
import React from 'react';
import { render, cleanup, act } from '@testing-library/react';
import { useAuth } from '../../src/react/use-auth.js';
import { _resetAuthClient, getAuthClient } from '@workfort/auth';

const MOCK_USER = {
  id: '1', username: 'kazw', name: 'Kaz Walker',
  displayName: 'Kaz', type: 'user' as const,
};
const MOCK_SESSION = {
  id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z',
};

function TestComponent() {
  const { user, isAuthenticated } = useAuth();
  return (
    <div>
      <span data-testid="auth">{isAuthenticated ? 'yes' : 'no'}</span>
      <span data-testid="user">{user?.username ?? 'none'}</span>
    </div>
  );
}

describe('useAuth (React)', () => {
  afterEach(() => { cleanup(); _resetAuthClient(); vi.restoreAllMocks(); });

  it('reflects auth state after init', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }), {
        status: 200, headers: { 'Content-Type': 'application/json' },
      }),
    );
    const { getByTestId } = render(<TestComponent />);
    expect(getByTestId('auth').textContent).toBe('no');
    await act(async () => { await getAuthClient().init(); });
    expect(getByTestId('auth').textContent).toBe('yes');
    expect(getByTestId('user').textContent).toBe('kazw');
  });
});
