import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
import { Required } from '@lion/ui/form-core.js';
import '../../src/form/wf-input.js';
import type { WfInput } from '../../src/form/wf-input.js';

describe('WfInput', () => {
  afterEach(cleanup);

  it('renders with wf-field and wf-input classes', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el.classList.contains('wf-field')).toBe(true);
    expect(el.classList.contains('wf-input')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders native input with wf-field__input class', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el._inputNode).toBeInstanceOf(HTMLInputElement);
    expect(el._inputNode.classList.contains('wf-field__input')).toBe(true);
  });

  it('renders label when set', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.label = 'Email';
    await el.updateComplete;
    const label = el.querySelector('[slot="label"]');
    expect(label).not.toBeNull();
    expect(label!.textContent).toContain('Email');
    expect(label!.classList.contains('wf-field__label')).toBe(true);
  });

  it('syncs modelValue to native input', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.modelValue = 'test@example.com';
    await el.updateComplete;
    expect(el._inputNode.value).toBe('test@example.com');
  });

  it('validates with Required', async () => {
    const el = await fixture<WfInput>('wf-input');
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
    const el = await fixture<WfInput>('wf-input');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-field--disabled')).toBe(true);
    expect(el._inputNode.disabled).toBe(true);
  });

  it('sets input type', async () => {
    const el = await fixture<WfInput>('wf-input', { type: 'email' });
    await el.updateComplete;
    expect(el._inputNode.type).toBe('email');
  });
});
