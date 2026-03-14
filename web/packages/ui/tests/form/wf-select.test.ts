import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
// @ts-expect-error — @lion/ui has no bundled type declarations
import { Required } from '@lion/ui/form-core.js';
import '../../src/form/wf-select.js';
import type { WfSelect } from '../../src/form/wf-select.js';

describe('WfSelect', () => {
  afterEach(cleanup);

  it('renders with wf-field and wf-select classes', async () => {
    const el = await fixture<WfSelect>('wf-select');
    expect(el.classList.contains('wf-field')).toBe(true);
    expect(el.classList.contains('wf-select')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfSelect>('wf-select');
    expect(el.shadowRoot).toBeNull();
  });

  it('has a native select element', async () => {
    const el = await fixture<WfSelect>('wf-select');
    expect(el._inputNode).toBeInstanceOf(HTMLSelectElement);
    expect(el._inputNode.classList.contains('wf-field__input')).toBe(true);
  });

  it('validates with Required', async () => {
    const el = await fixture<WfSelect>('wf-select');
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
    const el = await fixture<WfSelect>('wf-select');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-field--disabled')).toBe(true);
    expect(el._inputNode.disabled).toBe(true);
  });
});
