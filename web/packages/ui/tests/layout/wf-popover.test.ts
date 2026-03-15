import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-popover.js';
import type { WfPopover } from '../../src/layout/wf-popover.js';

describe('WfPopover', () => {
  afterEach(cleanup);

  it('renders with wf-popover class', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    expect(el.classList.contains('wf-popover')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    expect(el.shadowRoot).toBeNull();
  });

  it('content is hidden by default', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    expect(el.open).toBe(false);
  });

  it('toggle() opens the popover', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-popover--open')).toBe(true);
    el.hide();
  });

  it('toggle() closes when open', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    el.toggle();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('closes on outside click', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    // Simulate click outside (deferred handler registered via requestAnimationFrame)
    // Manually invoke the doc click handler since rAF may not fire in happy-dom
    document.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('fires wf-close event when closed', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('applies position class', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.position = 'left';
    el.toggle();
    await el.updateComplete;
    const content = el.querySelector('.wf-popover__content');
    expect(content!.classList.contains('wf-popover__content--left')).toBe(true);
    el.hide();
  });

  it('default position is bottom', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    const content = el.querySelector('.wf-popover__content');
    expect(content!.classList.contains('wf-popover__content--bottom')).toBe(true);
    el.hide();
  });
});
