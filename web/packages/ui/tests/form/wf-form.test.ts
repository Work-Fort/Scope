import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/form/wf-form.js';
import type { WfForm } from '../../src/form/wf-form.js';

describe('WfForm', () => {
  afterEach(cleanup);

  it('renders with wf-form class', async () => {
    const el = await fixture<WfForm>('wf-form');
    expect(el.classList.contains('wf-form')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfForm>('wf-form');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders a native <form> element', async () => {
    const el = await fixture<WfForm>('wf-form');
    const form = el.querySelector('form');
    expect(form).toBeInstanceOf(HTMLFormElement);
  });

  it('has serializedValue getter that returns empty object without fields', async () => {
    const el = await fixture<WfForm>('wf-form');
    expect(el.serializedValue).toEqual({});
  });

  it('starts with submitted = false', async () => {
    const el = await fixture<WfForm>('wf-form');
    expect(el.submitted).toBe(false);
  });

  it('dispatches wf-submit on form submission', async () => {
    const el = await fixture<WfForm>('wf-form');
    await el.updateComplete;

    const handler = vi.fn();
    el.addEventListener('wf-submit', handler);

    // Simulate a submit event on the inner form
    const form = el.querySelector('form')!;
    form.dispatchEvent(new Event('submit', { cancelable: true, bubbles: true }));

    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler.mock.calls[0][0].detail).toHaveProperty('data');
  });

  it('sets submitted state and class after submission', async () => {
    const el = await fixture<WfForm>('wf-form');
    await el.updateComplete;

    const form = el.querySelector('form')!;
    form.dispatchEvent(new Event('submit', { cancelable: true, bubbles: true }));

    expect(el.submitted).toBe(true);
    expect(el.classList.contains('wf-form--submitted')).toBe(true);
  });

  it('serializes child input values', async () => {
    const el = await fixture<WfForm>('wf-form');
    await el.updateComplete;

    const form = el.querySelector('form')!;
    const input = document.createElement('input');
    input.name = 'username';
    input.value = 'alice';
    form.appendChild(input);

    expect(el.serializedValue).toEqual({ username: 'alice' });
  });

  it('resets form and clears submitted state', async () => {
    const el = await fixture<WfForm>('wf-form');
    await el.updateComplete;

    // Submit first
    const form = el.querySelector('form')!;
    form.dispatchEvent(new Event('submit', { cancelable: true, bubbles: true }));
    expect(el.submitted).toBe(true);

    // Reset
    el.reset();
    expect(el.submitted).toBe(false);
    expect(el.classList.contains('wf-form--submitted')).toBe(false);
  });
});
