import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-tabs.js';
import '../../src/layout/wf-tab-panel.js';
import type { WfTabs } from '../../src/layout/wf-tabs.js';
import type { WfTabPanel } from '../../src/layout/wf-tab-panel.js';

async function createTabs(): Promise<WfTabs> {
  const el = await fixture<WfTabs>('wf-tabs');
  el.innerHTML = `
    <wf-tab-panel name="one" label="Tab One">Content One</wf-tab-panel>
    <wf-tab-panel name="two" label="Tab Two">Content Two</wf-tab-panel>
    <wf-tab-panel name="three" label="Tab Three">Content Three</wf-tab-panel>
  `;
  await el.updateComplete;
  // Allow child panels to upgrade
  await Promise.all(
    Array.from(el.querySelectorAll('wf-tab-panel')).map(
      (p) => (p as WfTabPanel).updateComplete,
    ),
  );
  el.selectTab('one');
  await el.updateComplete;
  return el;
}

describe('WfTabs', () => {
  afterEach(cleanup);

  it('renders with wf-tabs class', async () => {
    const el = await createTabs();
    expect(el.classList.contains('wf-tabs')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await createTabs();
    expect(el.shadowRoot).toBeNull();
  });

  it('renders tab buttons for each panel', async () => {
    const el = await createTabs();
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    expect(buttons.length).toBe(3);
    expect(buttons[0].textContent!.trim()).toBe('Tab One');
  });

  it('shows first panel by default', async () => {
    const el = await createTabs();
    const panels = el.querySelectorAll('wf-tab-panel');
    expect(panels[0].hasAttribute('active')).toBe(true);
    expect(panels[1].hasAttribute('active')).toBe(false);
  });

  it('switches panels on tab click', async () => {
    const el = await createTabs();
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    (buttons[1] as HTMLElement).click();
    await el.updateComplete;
    const panels = el.querySelectorAll('wf-tab-panel');
    expect(panels[0].hasAttribute('active')).toBe(false);
    expect(panels[1].hasAttribute('active')).toBe(true);
  });

  it('fires wf-tab-change event', async () => {
    const el = await createTabs();
    const handler = vi.fn();
    el.addEventListener('wf-tab-change', handler);
    el.selectTab('two');
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ name: 'two' });
  });

  it('supports keyboard navigation (ArrowRight)', async () => {
    const el = await createTabs();
    const tabList = el.querySelector('.wf-tabs__list') as HTMLElement;
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    (buttons[0] as HTMLElement).focus();
    tabList.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }));
    await el.updateComplete;
    expect(el.activeTab).toBe('two');
  });

  it('tab buttons have correct ARIA attributes', async () => {
    const el = await createTabs();
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    expect(buttons[0].getAttribute('role')).toBe('tab');
    expect(buttons[0].getAttribute('aria-selected')).toBe('true');
    expect(buttons[1].getAttribute('aria-selected')).toBe('false');
  });
});
