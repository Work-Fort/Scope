import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-tooltip.js';
import type { WfTooltip } from '../../src/layout/wf-tooltip.js';

describe('WfTooltip', () => {
  afterEach(cleanup);

  it('renders with wf-tooltip class', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    expect(el.classList.contains('wf-tooltip')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    expect(el.shadowRoot).toBeNull();
  });

  it('tooltip content is hidden by default', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help text';
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip).not.toBeNull();
    expect(tip.getAttribute('aria-hidden')).toBe('true');
  });

  it('shows tooltip on mouseenter', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    await el.updateComplete;
    el.dispatchEvent(new MouseEvent('mouseenter'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('false');
  });

  it('hides tooltip on mouseleave', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    await el.updateComplete;
    el.dispatchEvent(new MouseEvent('mouseenter'));
    await el.updateComplete;
    el.dispatchEvent(new MouseEvent('mouseleave'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('true');
  });

  it('shows tooltip on focusin', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    await el.updateComplete;
    el.dispatchEvent(new FocusEvent('focusin'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('false');
  });

  it('hides tooltip on focusout', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    await el.updateComplete;
    el.dispatchEvent(new FocusEvent('focusin'));
    await el.updateComplete;
    el.dispatchEvent(new FocusEvent('focusout'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('true');
  });

  it('applies position class', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    el.position = 'bottom';
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.classList.contains('wf-tooltip__content--bottom')).toBe(true);
  });

  it('default position is top', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.classList.contains('wf-tooltip__content--top')).toBe(true);
  });

  it('has correct ARIA attributes', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    el.content = 'Help';
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content');
    expect(tip!.getAttribute('role')).toBe('tooltip');
  });
});
