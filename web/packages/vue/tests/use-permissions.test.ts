import { describe, it, expect } from 'vitest';
import { usePermissions } from '../src/use-permissions.js';

describe('usePermissions (Vue)', () => {
  it('checks initial permissions', () => {
    const { can } = usePermissions(['send_message', 'channel_list']);
    expect(can('send_message')).toBe(true);
    expect(can('manage_roles')).toBe(false);
  });

  it('starts with empty permissions by default', () => {
    const { can } = usePermissions();
    expect(can('anything')).toBe(false);
  });

  it('updates permissions reactively', () => {
    const { can, update } = usePermissions([]);
    expect(can('send_message')).toBe(false);
    update(['send_message', 'manage_roles']);
    expect(can('send_message')).toBe(true);
    expect(can('manage_roles')).toBe(true);
  });
});
