import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/pagination.js';
import type { WfPagination } from '../../src/components/pagination.js';

describe('WfPagination', () => {
  afterEach(cleanup);

  it('renders with wf-pagination class', async () => {
    const el = await fixture<WfPagination>('wf-pagination');
    expect(el.classList.contains('wf-pagination')).toBe(true);
  });

  it('has navigation role and aria-label', async () => {
    const el = await fixture<WfPagination>('wf-pagination');
    expect(el.getAttribute('role')).toBe('navigation');
    expect(el.getAttribute('aria-label')).toBe('Pagination');
  });

  it('renders page buttons for count', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    const buttons = el.querySelectorAll('.wf-pagination__btn');
    // 5 page buttons + prev + next = 7
    expect(buttons.length).toBe(7);
  });

  it('marks current page with aria-current', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 3 });
    const currentBtn = el.querySelector('[aria-current="true"]');
    expect(currentBtn).not.toBeNull();
    expect(currentBtn!.textContent).toBe('3');
  });

  it('disables previous button on first page', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    const buttons = el.querySelectorAll('.wf-pagination__btn');
    expect((buttons[0] as HTMLButtonElement).disabled).toBe(true);
  });

  it('disables next button on last page', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 5 });
    const buttons = el.querySelectorAll('.wf-pagination__btn');
    const lastBtn = buttons[buttons.length - 1] as HTMLButtonElement;
    expect(lastBtn.disabled).toBe(true);
  });

  it('fires current-changed on page click', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    const handler = vi.fn();
    el.addEventListener('current-changed', handler);

    // Click page 3
    const buttons = el.querySelectorAll('.wf-pagination__btn');
    (buttons[3] as HTMLButtonElement).click(); // pages: prev, 1, 2, 3, ...
    expect(handler).toHaveBeenCalledOnce();
    expect(el.current).toBe(3);
  });

  it('navigates via next() and previous()', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 3 });
    el.next();
    expect(el.current).toBe(4);
    el.previous();
    expect(el.current).toBe(3);
  });

  it('does not go below 1 or above count', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 3, current: 1 });
    el.previous();
    expect(el.current).toBe(1);
    el.current = 3;
    el.next();
    expect(el.current).toBe(3);
  });

  it('shows ellipsis for many pages', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 20, current: 10 });
    const ellipses = el.querySelectorAll('.wf-pagination__ellipsis');
    expect(ellipses.length).toBeGreaterThan(0);
  });

  it('renders nothing when count is 0', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 0 });
    const buttons = el.querySelectorAll('.wf-pagination__btn');
    expect(buttons.length).toBe(0);
  });
});
