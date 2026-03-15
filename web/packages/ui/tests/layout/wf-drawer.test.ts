import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-drawer.js';
import type { WfDrawer } from '../../src/layout/wf-drawer.js';

describe('WfDrawer', () => {
  afterEach(cleanup);

  it('renders with wf-drawer class', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    expect(el.classList.contains('wf-drawer')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    expect(el.shadowRoot).toBeNull();
  });

  it('is hidden by default', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    expect(el.open).toBe(false);
  });

  it('shows panel when opened', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-drawer--open')).toBe(true);
    const panel = el.querySelector('.wf-drawer__panel');
    expect(panel).not.toBeNull();
    el.hide();
  });

  it('applies position class', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.position = 'left';
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-drawer__panel');
    expect(panel!.classList.contains('wf-drawer__panel--left')).toBe(true);
    el.hide();
  });

  it('default position is right', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-drawer__panel');
    expect(panel!.classList.contains('wf-drawer__panel--right')).toBe(true);
    el.hide();
  });

  it('creates backdrop when opened', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    expect(document.querySelector('.wf-overlay-backdrop')).not.toBeNull();
    el.hide();
    await el.updateComplete;
  });

  it('fires wf-close event on hide', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('renders header when set', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.header = 'Settings';
    el.show();
    await el.updateComplete;
    const header = el.querySelector('.wf-drawer__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toContain('Settings');
    el.hide();
  });
});
