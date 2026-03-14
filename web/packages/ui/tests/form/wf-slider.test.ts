import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-slider.js';
import type { WfSlider } from '../../src/form/wf-slider.js';

describe('WfSlider', () => {
  afterEach(cleanup);

  it('renders with wf-slider class', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    expect(el.classList.contains('wf-slider')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    expect(el.shadowRoot).toBeNull();
  });

  it('has native range input', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    const input = el.querySelector('input[type="range"]');
    expect(input).not.toBeNull();
    expect(input!.classList.contains('wf-slider__input')).toBe(true);
  });

  it('has default min/max/step', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    expect(el.min).toBe(0);
    expect(el.max).toBe(100);
    expect(el.step).toBe(1);
  });

  it('value syncs to native input', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    el.value = 42;
    await el.updateComplete;
    const input = el.querySelector('input[type="range"]') as HTMLInputElement;
    expect(input.value).toBe('42');
  });

  it('fires wf-input event on input', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    const handler = vi.fn();
    el.addEventListener('wf-input', handler);
    const input = el.querySelector('input[type="range"]') as HTMLInputElement;
    input.value = '50';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ value: 50 });
  });

  it('fires wf-change event on change', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    const handler = vi.fn();
    el.addEventListener('wf-change', handler);
    const input = el.querySelector('input[type="range"]') as HTMLInputElement;
    input.value = '75';
    input.dispatchEvent(new Event('change', { bubbles: true }));
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ value: 75 });
  });

  it('disabled state applies class and disables input', async () => {
    const el = await fixture<WfSlider>('wf-slider');
    el.disabled = true;
    await el.updateComplete;
    expect(el.classList.contains('wf-slider--disabled')).toBe(true);
    const input = el.querySelector('input[type="range"]') as HTMLInputElement;
    expect(input.disabled).toBe(true);
  });
});
