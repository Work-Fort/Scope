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

  it('hides when count is 0', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 0 });
    expect(el.style.display).toBe('none');
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
