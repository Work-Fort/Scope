import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/breadcrumbs.js';
import type { WfBreadcrumbs } from '../../src/components/breadcrumbs.js';

describe('WfBreadcrumbs', () => {
  afterEach(cleanup);

  it('renders with wf-breadcrumbs class', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    expect(el.classList.contains('wf-breadcrumbs')).toBe(true);
  });

  it('has navigation role and aria-label', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    expect(el.getAttribute('role')).toBe('navigation');
    expect(el.getAttribute('aria-label')).toBe('Breadcrumb');
  });

  it('renders breadcrumb items', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Settings', href: '/settings' },
      { label: 'Profile' },
    ];

    const items = el.querySelectorAll('.wf-breadcrumbs__item');
    expect(items.length).toBe(3);
  });

  it('renders last item as current page', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Profile' },
    ];

    const current = el.querySelector('.wf-breadcrumbs__current');
    expect(current).not.toBeNull();
    expect(current!.textContent).toBe('Profile');
    expect(current!.getAttribute('aria-current')).toBe('page');
  });

  it('renders links for non-last items with href', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Settings', href: '/settings' },
      { label: 'Profile' },
    ];

    const links = el.querySelectorAll('a.wf-breadcrumbs__link');
    expect(links.length).toBe(2);
    expect((links[0] as HTMLAnchorElement).href).toContain('/');
  });

  it('renders separators between items', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Settings', href: '/settings' },
      { label: 'Profile' },
    ];

    const seps = el.querySelectorAll('.wf-breadcrumbs__separator');
    expect(seps.length).toBe(2);
    expect(seps[0].textContent).toBe('/');
  });

  it('uses custom separator', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs', { separator: '>' });
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Profile' },
    ];

    const sep = el.querySelector('.wf-breadcrumbs__separator');
    expect(sep!.textContent).toBe('>');
  });

  it('dispatches wf-navigate on link click', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Profile' },
    ];

    const handler = vi.fn();
    el.addEventListener('wf-navigate', handler);

    const link = el.querySelector('a.wf-breadcrumbs__link') as HTMLAnchorElement;
    link.click();

    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({
      href: '/',
      label: 'Home',
      index: 0,
    });
  });
});
