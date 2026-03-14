// tests/spike/lion-light-dom.test.ts
//
// Self-contained spike: defines WfInputSpike inline, registers it,
// and verifies that Lion's form system works in light DOM.
// This test determines GO/NO-GO for the entire Phase 2.

import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import { LionInput } from '@lion/ui/input.js';
import { Required } from '@lion/ui/form-core.js';

// --- Spike class (inline — not a real component) ---

class WfInputSpike extends LionInput {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-input-spike');
  }
}

if (!customElements.get('wf-input-spike')) {
  customElements.define('wf-input-spike', WfInputSpike);
}

// --- GO/NO-GO Tests ---

describe('Lion + Light DOM Spike (GO/NO-GO)', () => {
  afterEach(cleanup);

  // GO criterion 1: Renders without shadow DOM
  it('renders without shadow DOM', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect(el.shadowRoot).toBeNull();
    expect(document.querySelector('wf-input-spike')).toBe(el);
  });

  // GO criterion 2: _inputNode resolves
  it('resolves _inputNode to a native <input>', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect(el._inputNode).toBeInstanceOf(HTMLInputElement);
  });

  // GO criterion 3: modelValue syncs
  it('syncs modelValue to native input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.modelValue = 'hello';
    await el.updateComplete;
    expect(el._inputNode.value).toBe('hello');
  });

  // GO criterion 4: Validation works
  it('Required validator sets hasFeedbackFor to include error', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.validators = [new Required()];
    el.modelValue = '';
    // Force interaction states so feedback shows
    (el as any).touched = true;
    (el as any).dirty = true;
    (el as any).submitted = true;
    await el.updateComplete;
    // Lion validation may be async — wait for feedbackComplete if available
    if ('feedbackComplete' in el) {
      await (el as any).feedbackComplete;
    }
    expect((el as any).hasFeedbackFor).toContain('error');
  });

  // GO criterion 5: Dirty state tracking
  it('tracks dirty state after user input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect((el as any).dirty).toBe(false);
    el._inputNode.value = 'typed';
    el._inputNode.dispatchEvent(new Event('input', { bubbles: true }));
    await el.updateComplete;
    expect((el as any).dirty).toBe(true);
  });

  // GO criterion 6: Touched state tracking
  // Lion's InteractionStateMixin listens for 'blur' on the host element.
  it('tracks touched state after blur', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect((el as any).touched).toBe(false);
    el.dispatchEvent(new Event('blur'));
    await el.updateComplete;
    expect((el as any).touched).toBe(true);
  });

  // GO criterion 7: Disabled propagates
  it('propagates disabled to native input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.disabled = true;
    await el.updateComplete;
    expect(el._inputNode.disabled).toBe(true);
  });
});
