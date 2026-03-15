// tests/components/button.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/button.js';
import type { WfButton } from '../../src/components/button.js';

describe('WfButton', () => {
  afterEach(cleanup);

  it('renders with wf-button class', async () => {
    const el = await fixture<WfButton>('wf-button');
    expect(el.classList.contains('wf-button')).toBe(true);
  });

  it('applies filled variant class', async () => {
    const el = await fixture<WfButton>('wf-button', { variant: 'filled' });
    expect(el.classList.contains('wf-button--filled')).toBe(true);
  });

  it('dispatches wf-click event on click', async () => {
    const el = await fixture<WfButton>('wf-button');
    const handler = vi.fn();
    el.addEventListener('wf-click', handler);
    el.click();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('does not dispatch wf-click when disabled', async () => {
    const el = await fixture<WfButton>('wf-button', { disabled: true });
    const handler = vi.fn();
    el.addEventListener('wf-click', handler);
    el.click();
    expect(handler).not.toHaveBeenCalled();
  });

  it('has button role and tabindex', async () => {
    const el = await fixture<WfButton>('wf-button');
    expect(el.getAttribute('role')).toBe('button');
    expect(el.getAttribute('tabindex')).toBe('0');
  });

  it('applies color class', async () => {
    const el = await fixture<WfButton>('wf-button', { color: 'red' });
    expect(el.classList.contains('wf-button--red')).toBe(true);
  });

  it('defaults to outline variant', async () => {
    const el = await fixture<WfButton>('wf-button');
    expect(el.classList.contains('wf-button--filled')).toBe(false);
  });
});
