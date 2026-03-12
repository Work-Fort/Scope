// tests/components/text-input.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/text-input.js';
import type { WfTextInput } from '../../src/components/text-input.js';

describe('WfTextInput', () => {
  afterEach(cleanup);

  it('renders an input element', async () => {
    const el = await fixture<WfTextInput>('wf-text-input');
    expect(el.querySelector('input')).not.toBeNull();
  });

  it('sets placeholder', async () => {
    const el = await fixture<WfTextInput>('wf-text-input', { placeholder: 'Type here...' });
    expect(el.querySelector('input')!.placeholder).toBe('Type here...');
  });

  it('dispatches wf-input on input', async () => {
    const el = await fixture<WfTextInput>('wf-text-input');
    const handler = vi.fn();
    el.addEventListener('wf-input', handler);
    const input = el.querySelector('input')!;
    input.value = 'hello';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    expect(handler).toHaveBeenCalledOnce();
    expect((handler.mock.calls[0][0] as CustomEvent).detail.value).toBe('hello');
  });

  it('reflects value property', async () => {
    const el = await fixture<WfTextInput>('wf-text-input');
    el.value = 'test';
    await el.updateComplete;
    expect(el.querySelector('input')!.value).toBe('test');
  });
});
