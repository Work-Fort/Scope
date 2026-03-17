import { describe, it, expect, vi, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/compose-input.js';
import type { WfComposeInput } from '../../src/components/compose-input.js';

describe('WfComposeInput', () => {
  afterEach(cleanup);

  it('renders with wf-compose-input class', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    expect(el.classList.contains('wf-compose-input')).toBe(true);
  });

  it('renders textarea and send button', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    expect(el.querySelector('textarea')).toBeTruthy();
    expect(el.querySelector('wf-button')).toBeTruthy();
  });

  it('applies placeholder', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input', { placeholder: 'Type here' });
    const textarea = el.querySelector('textarea');
    expect(textarea?.placeholder).toBe('Type here');
  });

  it('dispatches wf-send on Enter', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const handler = vi.fn();
    el.addEventListener('wf-send', handler);

    const textarea = el.querySelector('textarea')!;
    textarea.value = 'hello';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(handler).toHaveBeenCalledOnce();
    expect((handler.mock.calls[0][0] as CustomEvent).detail.body).toBe('hello');
  });

  it('does not dispatch wf-send on Shift+Enter', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const handler = vi.fn();
    el.addEventListener('wf-send', handler);

    const textarea = el.querySelector('textarea')!;
    textarea.value = 'hello';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', shiftKey: true, bubbles: true }));

    expect(handler).not.toHaveBeenCalled();
  });

  it('does not dispatch wf-send when empty', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const handler = vi.fn();
    el.addEventListener('wf-send', handler);

    const textarea = el.querySelector('textarea')!;
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(handler).not.toHaveBeenCalled();
  });

  it('clears textarea after send', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input');
    const textarea = el.querySelector('textarea')!;
    textarea.value = 'hello';
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

    expect(textarea.value).toBe('');
  });

  it('disables when disabled prop is set', async () => {
    const el = await fixture<WfComposeInput>('wf-compose-input', { disabled: true });
    const textarea = el.querySelector('textarea');
    expect(textarea?.disabled).toBe(true);
  });
});
