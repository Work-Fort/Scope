import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-dialog.js';
import type { WfDialog } from '../../src/layout/wf-dialog.js';

describe('WfDialog', () => {
  afterEach(cleanup);

  it('renders with wf-dialog class', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    expect(el.classList.contains('wf-dialog')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    expect(el.shadowRoot).toBeNull();
  });

  it('is hidden by default', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    expect(el.open).toBe(false);
    expect(el.classList.contains('wf-dialog--open')).toBe(false);
  });

  it('renders dialog content when opened', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-dialog--open')).toBe(true);
    el.hide();
  });

  it('renders header when set', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.header = 'Confirm';
    el.show();
    await el.updateComplete;
    const header = el.querySelector('.wf-dialog__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toContain('Confirm');
    el.hide();
  });

  it('creates backdrop when opened', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    const backdrop = document.querySelector('.wf-overlay-backdrop');
    expect(backdrop).not.toBeNull();
    el.hide();
    await el.updateComplete;
  });

  it('removes backdrop when closed', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    el.hide();
    await el.updateComplete;
    const backdrop = document.querySelector('.wf-overlay-backdrop');
    expect(backdrop).toBeNull();
  });

  it('fires wf-close event on hide', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('has correct ARIA attributes', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.header = 'Title';
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-dialog__panel');
    expect(panel!.getAttribute('role')).toBe('dialog');
    expect(panel!.getAttribute('aria-modal')).toBe('true');
    el.hide();
  });

  it('close button calls hide()', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.header = 'Title';
    el.show();
    await el.updateComplete;
    const closeBtn = el.querySelector('.wf-dialog__close') as HTMLElement;
    expect(closeBtn).not.toBeNull();
    closeBtn.click();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });
});
