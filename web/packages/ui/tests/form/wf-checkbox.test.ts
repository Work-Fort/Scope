import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-checkbox.js';
import '../../src/form/wf-checkbox-group.js';
import type { WfCheckbox } from '../../src/form/wf-checkbox.js';
import type { WfCheckboxGroup } from '../../src/form/wf-checkbox-group.js';

describe('WfCheckbox', () => {
  afterEach(cleanup);

  it('renders with wf-checkbox class', async () => {
    const el = await fixture<WfCheckbox>('wf-checkbox');
    expect(el.classList.contains('wf-field')).toBe(true);
    expect(el.classList.contains('wf-checkbox')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfCheckbox>('wf-checkbox');
    expect(el.shadowRoot).toBeNull();
  });

  it('has native checkbox input', async () => {
    const el = await fixture<WfCheckbox>('wf-checkbox');
    expect((el as any)._inputNode).toBeInstanceOf(HTMLInputElement);
    expect((el as any)._inputNode.type).toBe('checkbox');
  });

  it('checked state toggles', async () => {
    const el = await fixture<WfCheckbox>('wf-checkbox');
    expect((el as any).checked).toBe(false);
    (el as any).checked = true;
    await el.updateComplete;
    expect((el as any).checked).toBe(true);
  });

  it('disabled works', async () => {
    const el = await fixture<WfCheckbox>('wf-checkbox');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-field--disabled')).toBe(true);
  });
});

describe('WfCheckboxGroup', () => {
  afterEach(cleanup);

  it('renders with wf-checkbox-group class', async () => {
    const el = await fixture<WfCheckboxGroup>('wf-checkbox-group');
    expect(el.classList.contains('wf-checkbox-group')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfCheckboxGroup>('wf-checkbox-group');
    expect(el.shadowRoot).toBeNull();
  });
});
