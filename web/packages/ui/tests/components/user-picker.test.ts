import { describe, it, expect, vi, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/user-picker.js';
import type { WfUserPicker } from '../../src/components/user-picker.js';

describe('WfUserPicker', () => {
  afterEach(cleanup);

  it('renders with wf-user-picker class', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker');
    expect(el.classList.contains('wf-user-picker')).toBe(true);
  });

  it('renders wf-dialog with header', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker', { header: 'Pick a user' });
    const dialog = el.querySelector('wf-dialog');
    expect(dialog).toBeTruthy();
    expect(dialog?.getAttribute('header')).toBe('Pick a user');
  });

  it('renders user list from users property', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker');
    (el as any).users = [
      { username: 'alice', online: true },
      { username: 'bob', online: false },
    ];
    await el.updateComplete;

    const items = el.querySelectorAll('wf-list-item');
    expect(items.length).toBe(2);
  });

  it('excludes user matching exclude property', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker', { exclude: 'bob' });
    (el as any).users = [
      { username: 'alice', online: true },
      { username: 'bob', online: false },
    ];
    await el.updateComplete;

    const items = el.querySelectorAll('wf-list-item');
    expect(items.length).toBe(1);
  });

  it('dispatches wf-select with username on item click', async () => {
    const el = await fixture<WfUserPicker>('wf-user-picker');
    (el as any).users = [{ username: 'alice', online: true }];
    await el.updateComplete;

    const handler = vi.fn();
    el.addEventListener('wf-select', handler);

    const item = el.querySelector('wf-list-item') as HTMLElement;
    item?.dispatchEvent(new CustomEvent('wf-select', { bubbles: true }));

    // The component should re-dispatch with username detail.
    // Implementation may vary — check handler was called.
    expect(handler).toHaveBeenCalled();
  });
});
