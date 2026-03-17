import { describe, it, expect } from 'vitest';
import { PermissionSet } from '../../src/core/permission-set.js';

describe('PermissionSet', () => {
  it('reports permissions correctly', () => {
    const perms = new PermissionSet(['send_message', 'channel_list']);
    expect(perms.can('send_message')).toBe(true);
    expect(perms.can('channel_list')).toBe(true);
    expect(perms.can('manage_roles')).toBe(false);
  });

  it('starts empty', () => {
    const perms = new PermissionSet([]);
    expect(perms.can('anything')).toBe(false);
  });

  it('updates permissions', () => {
    const perms = new PermissionSet([]);
    perms.update(['send_message']);
    expect(perms.can('send_message')).toBe(true);
  });

  it('returns all permissions', () => {
    const perms = new PermissionSet(['a', 'b']);
    expect(perms.all()).toEqual(['a', 'b']);
  });
});
