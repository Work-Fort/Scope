import { describe, it, expect } from 'vitest';
import { createRoot } from 'solid-js';
import { usePermissions } from '../src/use-permissions.js';

describe('usePermissions (Solid)', () => {
  it('checks initial permissions', () => {
    createRoot((dispose) => {
      const { can, permissions } = usePermissions(['send_message', 'channel_list']);
      expect(can('send_message')).toBe(true);
      expect(can('manage_roles')).toBe(false);
      expect(permissions()).toEqual(['send_message', 'channel_list']);
      dispose();
    });
  });

  it('starts with empty permissions by default', () => {
    createRoot((dispose) => {
      const { can, permissions } = usePermissions();
      expect(can('anything')).toBe(false);
      expect(permissions()).toEqual([]);
      dispose();
    });
  });

  it('updates permissions reactively', () => {
    createRoot((dispose) => {
      const { can, update, permissions } = usePermissions([]);
      expect(can('send_message')).toBe(false);
      update(['send_message', 'manage_roles']);
      expect(can('send_message')).toBe(true);
      expect(can('manage_roles')).toBe(true);
      expect(permissions()).toEqual(['send_message', 'manage_roles']);
      dispose();
    });
  });
});
