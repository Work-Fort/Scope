import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/scroll-area.js';

describe('WfScrollArea', () => {
  afterEach(cleanup);

  it('renders with wf-scroll-area class', async () => {
    const el = await fixture('wf-scroll-area');
    expect(el.classList.contains('wf-scroll-area')).toBe(true);
  });
});
