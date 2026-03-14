import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-combobox.js';
import type { WfCombobox } from '../../src/form/wf-combobox.js';

const SAMPLE_OPTIONS = [
  { value: 'apple', label: 'Apple' },
  { value: 'banana', label: 'Banana' },
  { value: 'cherry', label: 'Cherry' },
];

describe('WfCombobox', () => {
  afterEach(cleanup);

  it('renders with wf-combobox class', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    expect(el.classList.contains('wf-combobox')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    expect(el.shadowRoot).toBeNull();
  });

  it('has text input with combobox role', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    const input = el.querySelector('input[type="text"]');
    expect(input).not.toBeNull();
    expect(input!.getAttribute('role')).toBe('combobox');
    expect(input!.classList.contains('wf-combobox__input')).toBe(true);
  });

  it('renders label when set', async () => {
    const el = await fixture<WfCombobox>('wf-combobox', { label: 'Fruit' });
    await el.updateComplete;
    const label = el.querySelector('.wf-field__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toBe('Fruit');
  });

  it('disabled state applies class and disables input', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-combobox--disabled')).toBe(true);
    const input = el.querySelector('input') as HTMLInputElement;
    expect(input.disabled).toBe(true);
  });

  it('renders options in listbox', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    el.options = SAMPLE_OPTIONS;
    await el.updateComplete;
    const items = el.querySelectorAll('.wf-combobox__option');
    expect(items.length).toBe(3);
    expect(items[0].textContent!.trim()).toBe('Apple');
  });

  it('opens dropdown via open() method', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    el.options = SAMPLE_OPTIONS;
    await el.updateComplete;
    el.open();
    await el.updateComplete;
    expect(el.classList.contains('wf-combobox--open')).toBe(true);
  });

  it('closes dropdown via close() method', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    el.options = SAMPLE_OPTIONS;
    el._open = true;
    await el.updateComplete;
    el.close();
    await el.updateComplete;
    expect(el.classList.contains('wf-combobox--open')).toBe(false);
  });

  it('fires wf-change event on selectOption()', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    el.options = SAMPLE_OPTIONS;
    el._open = true;
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-change', handler);
    el.selectOption(SAMPLE_OPTIONS[1]);
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({
      value: 'banana',
      label: 'Banana',
    });
    expect(el.value).toBe('banana');
  });

  it('filters options based on _filter', async () => {
    const el = await fixture<WfCombobox>('wf-combobox');
    el.options = SAMPLE_OPTIONS;
    el._filter = 'an';
    await el.updateComplete;
    const filtered = el.filteredOptions;
    expect(filtered.length).toBe(1);
    expect(filtered[0].value).toBe('banana');
  });
});
