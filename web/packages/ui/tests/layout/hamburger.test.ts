import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-hamburger.js';
import type { WfHamburger } from '../../src/layout/wf-hamburger.js';

describe('WfHamburger', () => {
  afterEach(cleanup);

  it('renders with wf-hamburger class', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    expect(el.classList.contains('wf-hamburger')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders hamburger icon button', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    const btn = el.querySelector('.wf-hamburger__button');
    expect(btn).not.toBeNull();
    expect(btn!.textContent!.trim()).toBe('\u2630');
  });

  it('dispatches wf-toggle on click with open=true', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    const handler = vi.fn();
    el.addEventListener('wf-toggle', handler);

    const btn = el.querySelector('.wf-hamburger__button') as HTMLButtonElement;
    btn.click();

    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ open: true });
  });

  it('dispatches wf-toggle on click with open=false when already open', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    el.open = true;
    await el.updateComplete;

    const handler = vi.fn();
    el.addEventListener('wf-toggle', handler);

    const btn = el.querySelector('.wf-hamburger__button') as HTMLButtonElement;
    btn.click();

    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ open: false });
  });

  it('default position is top-right', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    expect(el.position).toBe('top-right');
  });

  it('applies position class to button', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger', { position: 'bottom-left' });
    const btn = el.querySelector('.wf-hamburger__button');
    expect(btn!.classList.contains('wf-hamburger__button--bottom-left')).toBe(true);
  });

  it('shows overlay panel when open', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    el.open = true;
    await el.updateComplete;

    const panel = el.querySelector('.wf-hamburger__panel');
    expect(panel).not.toBeNull();
    expect(panel!.getAttribute('hidden')).toBeNull();
  });

  it('hides overlay panel when closed', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    expect(el.open).toBe(false);
    const panel = el.querySelector('.wf-hamburger__panel');
    expect(panel).not.toBeNull();
    expect(panel!.hasAttribute('hidden')).toBe(true);
  });

  it('creates backdrop when opened', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    el.open = true;
    await el.updateComplete;

    expect(document.querySelector('.wf-overlay-backdrop')).not.toBeNull();
    el.open = false;
    await el.updateComplete;
  });

  it('closes on backdrop click', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    el.open = true;
    await el.updateComplete;

    const backdrop = document.querySelector('.wf-overlay-backdrop') as HTMLElement;
    expect(backdrop).not.toBeNull();
    backdrop.click();
    await el.updateComplete;

    expect(el.open).toBe(false);
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    el.open = true;
    await el.updateComplete;

    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;

    expect(el.open).toBe(false);
  });

  it('sets aria-expanded on button', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    const btn = el.querySelector('.wf-hamburger__button') as HTMLButtonElement;
    expect(btn.getAttribute('aria-expanded')).toBe('false');

    el.open = true;
    await el.updateComplete;

    const btn2 = el.querySelector('.wf-hamburger__button') as HTMLButtonElement;
    expect(btn2.getAttribute('aria-expanded')).toBe('true');
  });

  it('sets aria-label on button', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger');
    const btn = el.querySelector('.wf-hamburger__button') as HTMLButtonElement;
    expect(btn.getAttribute('aria-label')).toBe('Menu');
  });

  it('applies panel slide direction based on position', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger', { position: 'top-left' });
    el.open = true;
    await el.updateComplete;

    const panel = el.querySelector('.wf-hamburger__panel');
    expect(panel!.classList.contains('wf-hamburger__panel--left')).toBe(true);
  });

  it('slides panel from right for top-right position', async () => {
    const el = await fixture<WfHamburger>('wf-hamburger', { position: 'top-right' });
    el.open = true;
    await el.updateComplete;

    const panel = el.querySelector('.wf-hamburger__panel');
    expect(panel!.classList.contains('wf-hamburger__panel--right')).toBe(true);
  });
});
