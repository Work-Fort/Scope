import { describe, it, expect, afterEach, vi, beforeEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-nav-bar.js';
import type { WfNavBar } from '../../src/layout/wf-nav-bar.js';

// Mock ResizeObserver since happy-dom doesn't support it
class MockResizeObserver {
  callback: ResizeObserverCallback;
  targets: Element[] = [];

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }

  observe(target: Element): void {
    this.targets.push(target);
  }

  unobserve(target: Element): void {
    this.targets = this.targets.filter((t) => t !== target);
  }

  disconnect(): void {
    this.targets = [];
  }
}

// Keep a reference to trigger resize callbacks in tests
let lastResizeObserver: MockResizeObserver | null = null;

describe('WfNavBar', () => {
  beforeEach(() => {
    lastResizeObserver = null;
    (globalThis as any).ResizeObserver = class extends MockResizeObserver {
      constructor(callback: ResizeObserverCallback) {
        super(callback);
        lastResizeObserver = this;
      }
    };
  });

  afterEach(() => {
    cleanup();
    delete (globalThis as any).ResizeObserver;
  });

  it('renders with wf-nav-bar class', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    expect(el.classList.contains('wf-nav-bar')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders brand area', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    const brand = el.querySelector('.wf-nav-bar__brand');
    expect(brand).not.toBeNull();
  });

  it('renders tabs area', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    const tabs = el.querySelector('.wf-nav-bar__tabs');
    expect(tabs).not.toBeNull();
  });

  it('renders actions area', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    const actions = el.querySelector('.wf-nav-bar__actions');
    expect(actions).not.toBeNull();
  });

  it('default breakpoint is 640', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    expect(el.breakpoint).toBe(640);
  });

  it('default hamburger-position is top-right', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    expect(el.hamburgerPosition).toBe('top-right');
  });

  it('accepts custom breakpoint', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar', { breakpoint: 800 });
    expect(el.breakpoint).toBe(800);
  });

  it('accepts custom hamburger-position', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar', { 'hamburger-position': 'top-left' });
    expect(el.hamburgerPosition).toBe('top-left');
  });

  it('renders overflow dropdown container', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    const overflow = el.querySelector('.wf-nav-bar__overflow');
    expect(overflow).not.toBeNull();
    // Overflow should be hidden by default (no overflow)
    expect(overflow!.hasAttribute('hidden')).toBe(true);
  });

  it('sets up ResizeObserver on tabs area', async () => {
    await fixture<WfNavBar>('wf-nav-bar');
    expect(lastResizeObserver).not.toBeNull();
    expect(lastResizeObserver!.targets.length).toBeGreaterThan(0);
  });

  it('renders collapsed class when in collapsed mode', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    // Simulate narrow viewport by setting collapsed state directly
    el.collapsed = true;
    await el.updateComplete;

    expect(el.classList.contains('wf-nav-bar--collapsed')).toBe(true);
  });

  it('shows hamburger when collapsed', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    el.collapsed = true;
    await el.updateComplete;

    const hamburger = el.querySelector('wf-hamburger');
    expect(hamburger).not.toBeNull();
    expect(hamburger!.hasAttribute('hidden')).toBe(false);
  });

  it('hides hamburger when not collapsed', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    const hamburger = el.querySelector('wf-hamburger');
    expect(hamburger).not.toBeNull();
    expect(hamburger!.hasAttribute('hidden')).toBe(true);
  });

  it('hides tabs when collapsed', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    el.collapsed = true;
    await el.updateComplete;

    const tabs = el.querySelector('.wf-nav-bar__tabs');
    expect(tabs!.hasAttribute('hidden')).toBe(true);
  });

  it('forwards hamburger-position to wf-hamburger', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar', { 'hamburger-position': 'bottom-left' });
    el.collapsed = true;
    await el.updateComplete;

    const hamburger = el.querySelector('wf-hamburger');
    expect(hamburger!.getAttribute('position')).toBe('bottom-left');
  });

  it('disconnects ResizeObserver on disconnect', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    expect(lastResizeObserver).not.toBeNull();
    const spy = vi.spyOn(lastResizeObserver!, 'disconnect');

    el.remove();

    expect(spy).toHaveBeenCalledOnce();
  });

  it('shows overflow button when hasOverflow is true', async () => {
    const el = await fixture<WfNavBar>('wf-nav-bar');
    el.hasOverflow = true;
    await el.updateComplete;

    const overflow = el.querySelector('.wf-nav-bar__overflow');
    expect(overflow!.hasAttribute('hidden')).toBe(false);
  });
});
