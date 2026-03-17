import { describe, it, expect } from 'vitest';
import { get } from 'svelte/store';
import { createPermissions } from '../src/permissions.js';

describe('createPermissions (Svelte)', () => {
  it('checks initial permissions', () => {
    const perms = createPermissions(['send_message', 'channel_list']);
    expect(perms.can('send_message')).toBe(true);
    expect(perms.can('manage_roles')).toBe(false);
    expect(get(perms)).toEqual(['send_message', 'channel_list']);
  });

  it('starts with empty permissions by default', () => {
    const perms = createPermissions();
    expect(perms.can('anything')).toBe(false);
    expect(get(perms)).toEqual([]);
  });

  it('updates permissions reactively', () => {
    const perms = createPermissions([]);
    expect(perms.can('send_message')).toBe(false);
    perms.update(['send_message', 'manage_roles']);
    expect(perms.can('send_message')).toBe(true);
    expect(perms.can('manage_roles')).toBe(true);
    expect(get(perms)).toEqual(['send_message', 'manage_roles']);
  });
});
