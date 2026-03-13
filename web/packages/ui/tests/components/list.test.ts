// tests/components/list.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/list.js';
import '../../src/components/list-item.js';
import type { WfListItem } from '../../src/components/list-item.js';

describe('WfList', () => {
  afterEach(cleanup);

  it('renders with wf-list class and list role', async () => {
    const el = await fixture('wf-list');
    expect(el.classList.contains('wf-list')).toBe(true);
    expect(el.getAttribute('role')).toBe('list');
  });
});

describe('WfListItem', () => {
  afterEach(cleanup);

  it('renders with wf-list-item class and listitem role', async () => {
    const el = await fixture<WfListItem>('wf-list-item');
    expect(el.classList.contains('wf-list-item')).toBe(true);
    expect(el.getAttribute('role')).toBe('listitem');
  });

  it('applies active class', async () => {
    const el = await fixture<WfListItem>('wf-list-item', { active: true });
    expect(el.classList.contains('wf-list-item--active')).toBe(true);
  });

  it('wraps trailing content in trailing container', async () => {
    const el = await fixture<WfListItem>('wf-list-item');
    const trailing = document.createElement('span');
    trailing.setAttribute('data-wf', 'trailing');
    trailing.textContent = '3';
    el.appendChild(trailing);
    el.requestUpdate();
    await el.updateComplete;
    const container = el.querySelector('.wf-list-item__trailing');
    expect(container).not.toBeNull();
    expect(container!.contains(trailing)).toBe(true);
  });

  it('dispatches wf-select on click', async () => {
    const el = await fixture<WfListItem>('wf-list-item');
    const handler = vi.fn();
    el.addEventListener('wf-select', handler);
    el.click();
    expect(handler).toHaveBeenCalledOnce();
  });
});
