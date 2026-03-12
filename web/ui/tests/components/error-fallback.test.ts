import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/error-fallback.js';
import type { WfErrorFallback } from '../../src/components/error-fallback.js';

describe('WfErrorFallback', () => {
  afterEach(cleanup);

  it('renders title and message', async () => {
    const el = await fixture<WfErrorFallback>('wf-error-fallback', {
      title: 'Oops',
      message: 'Something went wrong',
    });
    expect(el.querySelector('.wf-error-fallback__title')!.textContent).toBe('Oops');
    expect(el.querySelector('.wf-error-fallback__message')!.textContent).toBe('Something went wrong');
  });

  it('has alert role', async () => {
    const el = await fixture('wf-error-fallback');
    expect(el.getAttribute('role')).toBe('alert');
  });
});
