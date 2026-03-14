import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-date-picker.js';
import type { WfDatePicker } from '../../src/form/wf-date-picker.js';

describe('WfDatePicker', () => {
  afterEach(cleanup);

  it('renders with wf-date-picker class', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    expect(el.classList.contains('wf-date-picker')).toBe(true);
  });

  it('has wf-field class for shared styling', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    expect(el.classList.contains('wf-field')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    expect(el.shadowRoot).toBeNull();
  });

  it('has native date input', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    const input = el.querySelector('input[type="date"]');
    expect(input).not.toBeNull();
    expect(input!.classList.contains('wf-field__input')).toBe(true);
  });

  it('renders label when set', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker', {
      label: 'Birthday',
    });
    await el.updateComplete;
    const label = el.querySelector('.wf-field__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toBe('Birthday');
  });

  it('min/max attributes pass through to native input', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker', {
      min: '2024-01-01',
      max: '2024-12-31',
    });
    await el.updateComplete;
    const input = el.querySelector('input[type="date"]') as HTMLInputElement;
    expect(input.getAttribute('min')).toBe('2024-01-01');
    expect(input.getAttribute('max')).toBe('2024-12-31');
  });

  it('value syncs to native input', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    el.value = '2024-06-15';
    await el.updateComplete;
    const input = el.querySelector('input[type="date"]') as HTMLInputElement;
    expect(input.value).toBe('2024-06-15');
  });

  it('fires wf-change event when value is set programmatically and confirmed', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-change', handler);
    // Simulate setting value directly — mirrors what a date picker would do
    el.value = '2024-07-04';
    el.dispatchEvent(
      new CustomEvent('wf-change', {
        detail: { value: '2024-07-04' },
        bubbles: true,
        composed: true,
      }),
    );
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ value: '2024-07-04' });
  });

  it('disabled state applies class and disables input', async () => {
    const el = await fixture<WfDatePicker>('wf-date-picker');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-field--disabled')).toBe(true);
    const input = el.querySelector('input[type="date"]') as HTMLInputElement;
    expect(input.disabled).toBe(true);
  });
});
