import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-toggle.js';
import type { WfToggle } from '../../src/form/wf-toggle.js';

describe('WfToggle', () => {
  afterEach(cleanup);

  it('renders with wf-toggle class', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    expect(el.classList.contains('wf-toggle')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    expect(el.shadowRoot).toBeNull();
  });

  it('has hidden checkbox input', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    const input = el.querySelector('input[type="checkbox"]');
    expect(input).not.toBeNull();
  });

  it('clicking toggles checked state', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    expect(el.checked).toBe(false);
    const track = el.querySelector('.wf-toggle__track') as HTMLElement;
    track.click();
    await el.updateComplete;
    expect(el.checked).toBe(true);
    expect(el.classList.contains('wf-toggle--checked')).toBe(true);
  });

  it('fires wf-change event', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    const handler = vi.fn();
    el.addEventListener('wf-change', handler);
    const track = el.querySelector('.wf-toggle__track') as HTMLElement;
    track.click();
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ checked: true });
  });

  it('disabled prevents toggle', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-toggle--disabled')).toBe(true);
    const track = el.querySelector('.wf-toggle__track') as HTMLElement;
    track.click();
    await el.updateComplete;
    expect(el.checked).toBe(false);
  });

  it('has track and thumb elements', async () => {
    const el = await fixture<WfToggle>('wf-toggle');
    expect(el.querySelector('.wf-toggle__track')).not.toBeNull();
    expect(el.querySelector('.wf-toggle__thumb')).not.toBeNull();
  });
});
