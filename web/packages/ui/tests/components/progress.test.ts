import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/progress.js';
import type { WfProgress } from '../../src/components/progress.js';

describe('WfProgress', () => {
  afterEach(cleanup);

  it('renders with wf-progress class', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    expect(el.classList.contains('wf-progress')).toBe(true);
  });

  it('has progressbar role', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    expect(el.getAttribute('role')).toBe('progressbar');
  });

  it('renders track and fill elements', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    expect(el.querySelector('.wf-progress__track')).not.toBeNull();
    expect(el.querySelector('.wf-progress__fill')).not.toBeNull();
  });

  it('sets aria-valuenow for determinate progress', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 50 });
    expect(el.getAttribute('aria-valuenow')).toBe('50');
    expect(el.getAttribute('aria-valuemin')).toBe('0');
    expect(el.getAttribute('aria-valuemax')).toBe('100');
  });

  it('calculates percentage correctly', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 50, min: 0, max: 100 });
    expect(el.percentage).toBe(50);
  });

  it('calculates percentage with custom min/max', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 75, min: 50, max: 100 });
    expect(el.percentage).toBe(50);
  });

  it('sets fill width based on value', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 60 });
    const fill = el.querySelector('.wf-progress__fill') as HTMLElement;
    expect(fill.style.width).toBe('60%');
  });

  it('removes aria-valuenow for indeterminate mode', async () => {
    const el = await fixture<WfProgress>('wf-progress', { indeterminate: true });
    expect(el.hasAttribute('aria-valuenow')).toBe(false);
    expect(el.classList.contains('wf-progress--indeterminate')).toBe(true);
  });

  it('applies size classes', async () => {
    const el = await fixture<WfProgress>('wf-progress', { size: 'sm' });
    expect(el.classList.contains('wf-progress--sm')).toBe(true);
  });

  it('shows label text', async () => {
    const el = await fixture<WfProgress>('wf-progress', { label: 'Uploading' });
    const labelEl = el.querySelector('.wf-progress__label-text');
    expect(labelEl!.textContent).toBe('Uploading');
  });

  it('shows percentage text', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 42 });
    const valueEl = el.querySelector('.wf-progress__value-text');
    expect(valueEl!.textContent).toBe('42%');
  });

  it('clamps value to min/max', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 150 });
    expect(el.percentage).toBe(100);
  });
});
