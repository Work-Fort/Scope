import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/spinner.js';
import type { WfSpinner } from '../../src/components/spinner.js';

describe('WfSpinner', () => {
  afterEach(cleanup);

  it('renders with wf-spinner class', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    expect(el.classList.contains('wf-spinner')).toBe(true);
  });

  it('has status role', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    expect(el.getAttribute('role')).toBe('status');
  });

  it('renders SVG spinner', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    const svg = el.querySelector('.wf-spinner__svg');
    expect(svg).not.toBeNull();
  });

  it('renders screen reader label', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    const label = el.querySelector('.wf-spinner__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toBe('Loading');
  });

  it('applies default md size class', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    expect(el.classList.contains('wf-spinner--md')).toBe(true);
  });

  it('applies sm size class', async () => {
    const el = await fixture<WfSpinner>('wf-spinner', { size: 'sm' });
    expect(el.classList.contains('wf-spinner--sm')).toBe(true);
  });

  it('applies lg size class', async () => {
    const el = await fixture<WfSpinner>('wf-spinner', { size: 'lg' });
    expect(el.classList.contains('wf-spinner--lg')).toBe(true);
  });

  it('uses custom label', async () => {
    const el = await fixture<WfSpinner>('wf-spinner', { label: 'Saving' });
    const label = el.querySelector('.wf-spinner__label');
    expect(label!.textContent).toBe('Saving');
  });
});
