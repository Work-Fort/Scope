import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-table.js';
import type { WfTable, WfTableColumn } from '../../src/layout/wf-table.js';

const COLUMNS: WfTableColumn[] = [
  { key: 'name', header: 'Name' },
  { key: 'email', header: 'Email' },
  { key: 'role', header: 'Role' },
];

const DATA = [
  { name: 'Alice', email: 'alice@example.com', role: 'Admin' },
  { name: 'Bob', email: 'bob@example.com', role: 'User' },
  { name: 'Charlie', email: 'charlie@example.com', role: 'User' },
];

describe('WfTable', () => {
  afterEach(cleanup);

  it('renders with wf-table class', async () => {
    const el = await fixture<WfTable>('wf-table');
    expect(el.classList.contains('wf-table')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfTable>('wf-table');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders table headers from columns', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    const ths = el.querySelectorAll('th');
    expect(ths.length).toBe(3);
    expect(ths[0].textContent).toContain('Name');
    expect(ths[1].textContent).toContain('Email');
  });

  it('renders data rows', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(3);
    const cells = rows[0].querySelectorAll('td');
    expect(cells[0].textContent!.trim()).toBe('Alice');
    expect(cells[1].textContent!.trim()).toBe('alice@example.com');
  });

  it('renders empty state when no data', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = [];
    await el.updateComplete;
    const empty = el.querySelector('.wf-table__empty');
    expect(empty).not.toBeNull();
  });

  it('applies striped variant', async () => {
    const el = await fixture<WfTable>('wf-table', { striped: true });
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    expect(el.classList.contains('wf-table--striped')).toBe(true);
  });

  it('fires wf-row-click event on row click', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-row-click', handler);
    const row = el.querySelector('tbody tr') as HTMLElement;
    row.click();
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ row: DATA[0], index: 0 });
  });

  // Sorting tests
  it('sorts by column when sortable', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [
      { key: 'name', header: 'Name', sortable: true },
      { key: 'email', header: 'Email' },
    ];
    el.data = [
      { name: 'Charlie', email: 'c@x.com' },
      { name: 'Alice', email: 'a@x.com' },
      { name: 'Bob', email: 'b@x.com' },
    ];
    await el.updateComplete;

    el.sort('name');
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Alice');
    expect(rows[2].querySelector('td')!.textContent!.trim()).toBe('Charlie');
  });

  it('toggles sort direction', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name', sortable: true }];
    el.data = [
      { name: 'Alice' },
      { name: 'Bob' },
      { name: 'Charlie' },
    ];
    await el.updateComplete;

    el.sort('name');
    expect(el.sortDirection).toBe('asc');
    el.sort('name');
    expect(el.sortDirection).toBe('desc');
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Charlie');
  });

  it('fires wf-sort event', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name', sortable: true }];
    el.data = [{ name: 'Alice' }];
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-sort', handler);
    el.sort('name');
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ key: 'name', direction: 'asc' });
  });

  // Pagination tests
  it('paginates data', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;

    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(10);
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Row 1');
  });

  it('navigates pages', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;

    el.goToPage(2);
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(10);
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Row 11');
  });

  it('last page shows remaining rows', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;

    el.goToPage(3);
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(5);
  });

  it('fires wf-page-change event', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-page-change', handler);
    el.goToPage(2);
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ page: 2 });
  });

  it('supports custom column render function', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [
      {
        key: 'name',
        header: 'Name',
        render: (val: unknown) => `**${val}**`,
      },
    ];
    el.data = [{ name: 'Alice' }];
    await el.updateComplete;
    const td = el.querySelector('tbody td');
    expect(td!.textContent!.trim()).toBe('**Alice**');
  });
});
