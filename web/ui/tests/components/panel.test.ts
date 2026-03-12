// tests/components/panel.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/panel.js';
import type { WfPanel } from '../../src/components/panel.js';

describe('WfPanel', () => {
  afterEach(cleanup);

  it('is registered as a custom element', () => {
    expect(customElements.get('wf-panel')).toBeDefined();
  });

  it('renders with wf-panel class', async () => {
    const el = await fixture<WfPanel>('wf-panel');
    expect(el.classList.contains('wf-panel')).toBe(true);
  });

  it('renders label when provided', async () => {
    const el = await fixture<WfPanel>('wf-panel', { label: 'Channels' });
    const label = el.querySelector('.wf-panel__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toBe('Channels');
  });

  it('does not render label when empty', async () => {
    const el = await fixture<WfPanel>('wf-panel');
    const label = el.querySelector('.wf-panel__label');
    expect(label).toBeNull();
  });

  it('preserves consumer-provided children', async () => {
    const el = await fixture<WfPanel>('wf-panel');
    const child = document.createElement('div');
    child.className = 'user-content';
    child.textContent = 'Hello';
    el.appendChild(child);
    await el.updateComplete;
    const found = el.querySelector('.user-content');
    expect(found).not.toBeNull();
    expect(found!.textContent).toBe('Hello');
  });
});
