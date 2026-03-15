import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-accordion.js';
import '../../src/layout/wf-accordion-item.js';
import type { WfAccordion } from '../../src/layout/wf-accordion.js';
import type { WfAccordionItem } from '../../src/layout/wf-accordion-item.js';

async function createAccordion(multiple = false): Promise<WfAccordion> {
  const el = await fixture<WfAccordion>('wf-accordion');
  if (multiple) el.setAttribute('multiple', '');
  el.innerHTML = `
    <wf-accordion-item name="one" header="Section One">Content One</wf-accordion-item>
    <wf-accordion-item name="two" header="Section Two">Content Two</wf-accordion-item>
    <wf-accordion-item name="three" header="Section Three">Content Three</wf-accordion-item>
  `;
  await el.updateComplete;
  await Promise.all(
    Array.from(el.querySelectorAll('wf-accordion-item')).map(
      (p) => (p as WfAccordionItem).updateComplete,
    ),
  );
  return el;
}

describe('WfAccordion', () => {
  afterEach(cleanup);

  it('renders with wf-accordion class', async () => {
    const el = await createAccordion();
    expect(el.classList.contains('wf-accordion')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await createAccordion();
    expect(el.shadowRoot).toBeNull();
  });

  it('all items collapsed by default', async () => {
    const el = await createAccordion();
    const items = el.querySelectorAll('wf-accordion-item');
    items.forEach((item) => {
      expect(item.hasAttribute('expanded')).toBe(false);
    });
  });

  it('clicking header expands item', async () => {
    const el = await createAccordion();
    const items = el.querySelectorAll('wf-accordion-item');
    const header = items[0].querySelector('.wf-accordion-item__header') as HTMLElement;
    header.click();
    await (items[0] as WfAccordionItem).updateComplete;
    expect(items[0].hasAttribute('expanded')).toBe(true);
  });

  it('single mode: opening one closes others', async () => {
    const el = await createAccordion(false);
    const items = el.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
    items[0].toggle();
    await el.updateComplete;
    expect(items[0].expanded).toBe(true);
    items[1].toggle();
    await el.updateComplete;
    expect(items[0].expanded).toBe(false);
    expect(items[1].expanded).toBe(true);
  });

  it('multiple mode: opening one does not close others', async () => {
    const el = await createAccordion(true);
    const items = el.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
    items[0].toggle();
    await el.updateComplete;
    items[1].toggle();
    await el.updateComplete;
    expect(items[0].expanded).toBe(true);
    expect(items[1].expanded).toBe(true);
  });

  it('fires wf-accordion-change event', async () => {
    const el = await createAccordion();
    const handler = vi.fn();
    el.addEventListener('wf-accordion-change', handler);
    const items = el.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
    items[0].toggle();
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ name: 'one', expanded: true });
  });

  it('header has correct ARIA attributes', async () => {
    const el = await createAccordion();
    const item = el.querySelector('wf-accordion-item') as WfAccordionItem;
    const header = item.querySelector('.wf-accordion-item__header') as HTMLElement;
    expect(header.getAttribute('role')).toBe('button');
    expect(header.getAttribute('aria-expanded')).toBe('false');
  });
});
