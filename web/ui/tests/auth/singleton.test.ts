import { describe, it, expect, beforeEach } from 'vitest';
import { getAuthClient, _resetAuthClient } from '../../src/auth/index.js';

describe('getAuthClient', () => {
  beforeEach(() => { _resetAuthClient(); });

  it('returns the same instance on repeated calls', () => {
    const a = getAuthClient();
    const b = getAuthClient();
    expect(a).toBe(b);
  });

  it('returns an AuthClient instance', () => {
    const client = getAuthClient();
    expect(typeof client.init).toBe('function');
    expect(typeof client.getUser).toBe('function');
  });
});
