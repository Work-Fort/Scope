import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-card.js';
import type { WfCard } from '../../src/layout/wf-card.js';

describe('WfCard', () => {
  afterEach(cleanup);

  it('renders with wf-card class', async () => {
    const el = await fixture<WfCard>('wf-card');
    expect(el.classList.contains('wf-card')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfCard>('wf-card');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders slot content', async () => {
    const el = await fixture<WfCard>('wf-card');
    el.innerHTML = '<p>Card content</p>';
    await el.updateComplete;
    expect(el.querySelector('p')!.textContent).toBe('Card content');
  });

  it('renders header when set', async () => {
    const el = await fixture<WfCard>('wf-card');
    el.header = 'Title';
    await el.updateComplete;
    const header = el.querySelector('.wf-card__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toBe('Title');
  });

  it('renders footer when set', async () => {
    const el = await fixture<WfCard>('wf-card');
    el.footer = 'Footer text';
    await el.updateComplete;
    const footer = el.querySelector('.wf-card__footer');
    expect(footer).not.toBeNull();
    expect(footer!.textContent).toBe('Footer text');
  });

  it('applies variant class', async () => {
    const el = await fixture<WfCard>('wf-card', { variant: 'outlined' });
    await el.updateComplete;
    expect(el.classList.contains('wf-card--outlined')).toBe(true);
  });

  it('applies padding class when padded attribute set', async () => {
    const el = await fixture<WfCard>('wf-card', { padded: true });
    await el.updateComplete;
    expect(el.classList.contains('wf-card--padded')).toBe(true);
  });
});
