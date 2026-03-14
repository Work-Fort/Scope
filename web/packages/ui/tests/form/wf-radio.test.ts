import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-radio.js';
import '../../src/form/wf-radio-group.js';
import type { WfRadio } from '../../src/form/wf-radio.js';
import type { WfRadioGroup } from '../../src/form/wf-radio-group.js';

describe('WfRadio', () => {
  afterEach(cleanup);

  it('renders with wf-radio class', async () => {
    const el = await fixture<WfRadio>('wf-radio');
    expect(el.classList.contains('wf-field')).toBe(true);
    expect(el.classList.contains('wf-radio')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfRadio>('wf-radio');
    expect(el.shadowRoot).toBeNull();
  });

  it('has native radio input', async () => {
    const el = await fixture<WfRadio>('wf-radio');
    expect((el as any)._inputNode).toBeInstanceOf(HTMLInputElement);
    expect((el as any)._inputNode.type).toBe('radio');
  });

  it('checked state works', async () => {
    const el = await fixture<WfRadio>('wf-radio');
    expect((el as any).checked).toBe(false);
    (el as any).checked = true;
    await el.updateComplete;
    expect((el as any).checked).toBe(true);
  });

  it('disabled works', async () => {
    const el = await fixture<WfRadio>('wf-radio');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-field--disabled')).toBe(true);
  });
});

describe('WfRadioGroup', () => {
  afterEach(cleanup);

  it('renders with wf-radio-group class', async () => {
    const el = await fixture<WfRadioGroup>('wf-radio-group');
    expect(el.classList.contains('wf-radio-group')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfRadioGroup>('wf-radio-group');
    expect(el.shadowRoot).toBeNull();
  });
});
