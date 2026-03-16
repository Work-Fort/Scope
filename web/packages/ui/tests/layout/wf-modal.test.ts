import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-modal.js';
import type { WfModal } from '../../src/layout/wf-modal.js';

describe('WfModal', () => {
  afterEach(cleanup);

  it('renders with wf-modal class', async () => {
    const el = await fixture<WfModal>('wf-modal');
    expect(el.classList.contains('wf-modal')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfModal>('wf-modal');
    expect(el.shadowRoot).toBeNull();
  });

  it('is hidden by default', async () => {
    const el = await fixture<WfModal>('wf-modal');
    expect(el.open).toBe(false);
    expect(el.classList.contains('wf-modal--open')).toBe(false);
  });

  it('renders dialog content when opened', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.show();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-modal--open')).toBe(true);
    el.hide();
  });

  it('renders header when set', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.header = 'Confirm';
    el.show();
    await el.updateComplete;
    const header = el.querySelector('.wf-modal__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toContain('Confirm');
    el.hide();
  });

  it('creates backdrop when opened', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.show();
    await el.updateComplete;
    const backdrop = document.querySelector('.wf-overlay-backdrop');
    expect(backdrop).not.toBeNull();
    el.hide();
    await el.updateComplete;
  });

  it('removes backdrop when closed', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.show();
    await el.updateComplete;
    el.hide();
    await el.updateComplete;
    const backdrop = document.querySelector('.wf-overlay-backdrop');
    expect(backdrop).toBeNull();
  });

  it('fires wf-close event on hide', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.show();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.show();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('has correct ARIA attributes', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.header = 'Title';
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-modal__panel');
    expect(panel!.getAttribute('role')).toBe('dialog');
    expect(panel!.getAttribute('aria-modal')).toBe('true');
    el.hide();
  });

  it('close button calls hide()', async () => {
    const el = await fixture<WfModal>('wf-modal');
    el.header = 'Title';
    el.show();
    await el.updateComplete;
    const closeBtn = el.querySelector('.wf-modal__close') as HTMLElement;
    expect(closeBtn).not.toBeNull();
    closeBtn.click();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });
});
