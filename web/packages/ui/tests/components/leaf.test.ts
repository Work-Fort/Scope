// tests/components/leaf.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/badge.js';
import '../../src/components/status-dot.js';
import '../../src/components/skeleton.js';
import '../../src/components/divider.js';
import type { WfBadge } from '../../src/components/badge.js';
import type { WfStatusDot } from '../../src/components/status-dot.js';
import type { WfSkeleton } from '../../src/components/skeleton.js';

describe('WfBadge', () => {
  afterEach(cleanup);

  it('renders count', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 5 });
    expect(el.textContent).toContain('5');
  });

  it('shows zero by default', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 0 });
    expect(el.style.display).not.toBe('none');
    expect(el.textContent).toBe('0');
  });

  it('hides when hidden prop is set', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 5, hidden: true });
    expect(el.style.display).toBe('none');
  });

  it('defaults to md size', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 3 });
    expect(el.classList.contains('wf-badge--md')).toBe(true);
  });

  it('applies sm size class', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 3, size: 'sm' });
    expect(el.classList.contains('wf-badge--sm')).toBe(true);
  });

  it('applies lg size class', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 3, size: 'lg' });
    expect(el.classList.contains('wf-badge--lg')).toBe(true);
  });

});

describe('WfStatusDot', () => {
  afterEach(cleanup);

  it('applies status class', async () => {
    const el = await fixture<WfStatusDot>('wf-status-dot', { status: 'online' });
    expect(el.classList.contains('wf-status-dot--online')).toBe(true);
  });

  it('defaults to offline', async () => {
    const el = await fixture<WfStatusDot>('wf-status-dot');
    expect(el.classList.contains('wf-status-dot--offline')).toBe(true);
  });
});

describe('WfSkeleton', () => {
  afterEach(cleanup);

  it('applies dimensions from attributes', async () => {
    const el = await fixture<WfSkeleton>('wf-skeleton', { width: '100px', height: '20px' });
    expect(el.style.width).toBe('100px');
    expect(el.style.height).toBe('20px');
  });
});

describe('WfDivider', () => {
  afterEach(cleanup);

  it('renders with wf-divider class', async () => {
    const el = await fixture('wf-divider');
    expect(el.classList.contains('wf-divider')).toBe(true);
  });

  it('has separator role', async () => {
    const el = await fixture('wf-divider');
    expect(el.getAttribute('role')).toBe('separator');
  });
});
