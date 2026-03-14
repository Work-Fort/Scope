import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
import { Required } from '@lion/ui/form-core.js';
import '../../src/form/wf-textarea.js';
import type { WfTextarea } from '../../src/form/wf-textarea.js';

describe('WfTextarea', () => {
  afterEach(cleanup);

  it('renders with wf-field and wf-textarea classes', async () => {
    const el = await fixture<WfTextarea>('wf-textarea');
    expect(el.classList.contains('wf-field')).toBe(true);
    expect(el.classList.contains('wf-textarea')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfTextarea>('wf-textarea');
    expect(el.shadowRoot).toBeNull();
  });

  it('has a native textarea element', async () => {
    const el = await fixture<WfTextarea>('wf-textarea');
    expect(el._inputNode).toBeInstanceOf(HTMLTextAreaElement);
    expect(el._inputNode.classList.contains('wf-field__input')).toBe(true);
  });

  it('syncs modelValue to native textarea', async () => {
    const el = await fixture<WfTextarea>('wf-textarea');
    el.modelValue = 'Hello\nWorld';
    await el.updateComplete;
    expect(el._inputNode.value).toBe('Hello\nWorld');
  });

  it('validates with Required', async () => {
    const el = await fixture<WfTextarea>('wf-textarea');
    el.validators = [new Required()];
    el.modelValue = '';
    (el as any).touched = true;
    (el as any).dirty = true;
    (el as any).submitted = true;
    await el.updateComplete;
    if ('feedbackComplete' in el) await (el as any).feedbackComplete;
    expect((el as any).hasFeedbackFor).toContain('error');
    expect(el.classList.contains('wf-field--error')).toBe(true);
  });

  it('applies disabled class and state', async () => {
    const el = await fixture<WfTextarea>('wf-textarea');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-field--disabled')).toBe(true);
    expect(el._inputNode.disabled).toBe(true);
  });
});
