# Phase 2: Forms — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add 11 Lion-backed form components to `@workfort/ui` with validation, interaction state tracking, and framework adapter updates.

**Architecture:** Each form component extends Lion's class hierarchy (LionField/LionInput/etc.), overrides `createRenderRoot()` for light DOM, and provides a custom `render()` method with WorkFort CSS classes and token-based styling. Lion provides validation lifecycle (dirty/touched/submitted), ARIA attributes, keyboard navigation, and focus management. A spike validates this integration pattern before building all components.

**Tech Stack:** @lion/ui 0.16.x, Lit 3, TypeScript, Vite (library mode), Vitest (happy-dom)

**Spec:** `docs/ui-component-library-design.md` — Phase 2: Forms section

---

## Lion Integration Model

Lion's class hierarchy for form inputs:

```
LitElement
  └─ SlotMixin
       └─ ValidateMixin
            └─ FormatMixin
                 └─ FocusMixin
                      └─ InteractionStateMixin
                           └─ FormControlMixin
                                └─ LionField
                                     └─ NativeTextFieldMixin(LionField) = LionInput
```

**Key behaviors we inherit:**
- `FormControlMixin.render()` produces `<div class="form-field__group-one">` / `<div class="form-field__group-two">` with `<slot name="label">`, `<slot name="input">`, `<slot name="feedback">`, etc.
- `SlotMixin` creates child elements with `slot="X"` attributes (e.g., `LionInput` creates a native `<input slot="input">`). `_inputNode` queries `[slot=input]`.
- `ScopedElementsMixin` (used by ValidateMixin for `lion-validation-feedback`) falls back to global registry when no shadow root exists.
- `InteractionStateMixin` provides `dirty`, `touched`, `submitted`, `prefilled` states.
- `ValidateMixin` provides `validators` array, `hasFeedbackFor`, and `showsFeedbackFor`.

**Light DOM challenge:** `<slot>` elements are INERT in light DOM — they don't project content. SlotMixin's children with `slot="X"` attributes exist as siblings of the rendered template, not projected into `<slot>` elements. This is the primary risk the spike addresses.

**WorkFort pattern (to be validated by spike):**

```typescript
class WfInput extends LionInput {
  createRenderRoot() { return this; }

  connectedCallback() {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-input');
  }

  // Override render to replace Lion's slot-based template with direct DOM
  render() {
    return html`
      <label class="wf-field__label" @click=${() => this._inputNode?.focus()}>
        ${this.label}
      </label>
      <div class="wf-field__container">
        <slot name="prefix"></slot>
        <slot name="input"></slot>
        <slot name="suffix"></slot>
      </div>
      ${this.helpText ? html`<div class="wf-field__help-text">${this.helpText}</div>` : nothing}
      ${this._feedbackTemplate()}
    `;
  }
}
```

**Note:** The `<slot name="input">` in the override may or may not work in light DOM. SlotMixin appends an `<input slot="input">` as a child of the host. In light DOM, both the `<slot>` element and the `<input>` are siblings — no projection occurs. The spike will determine whether we need to manually position the input node, use a different approach, or whether Lion handles this gracefully.

---

## Component Matrix

| Component | Tag | Lion Base | Import Path | Overlay? |
|-----------|-----|-----------|-------------|----------|
| Input | `wf-input` | `LionInput` | `@lion/ui/input.js` | No |
| Textarea | `wf-textarea` | `LionTextarea` | `@lion/ui/textarea.js` | No |
| Checkbox | `wf-checkbox` | `LionCheckbox` | `@lion/ui/checkbox-group.js` | No |
| CheckboxGroup | `wf-checkbox-group` | `LionCheckboxGroup` | `@lion/ui/checkbox-group.js` | No |
| Radio | `wf-radio` | `LionRadio` | `@lion/ui/radio-group.js` | No |
| RadioGroup | `wf-radio-group` | `LionRadioGroup` | `@lion/ui/radio-group.js` | No |
| Toggle | `wf-toggle` | Custom (`ChoiceInputMixin`) | `@lion/ui/form-core.js` | No |
| Select | `wf-select` | `LionSelectRich` | `@lion/ui/select-rich.js` | Yes |
| Combobox | `wf-combobox` | `LionCombobox` | `@lion/ui/combobox.js` | Yes |
| Slider | `wf-slider` | `LionInputRange` | `@lion/ui/input-range.js` | No |
| DatePicker | `wf-date-picker` | `LionInputDatepicker` | `@lion/ui/input-datepicker.js` | Yes |
| FileUpload | `wf-file-upload` | `LionInputFile` | `@lion/ui/input-file.js` | No |
| Form | `wf-form` | `LionForm` | `@lion/ui/form.js` | No |

---

## GO/NO-GO Gate Definition

Chunk 0 is a spike. **ALL** of the following must pass for GO:

### GO criteria (ALL must pass)

1. `WfInputSpike` renders without runtime errors
2. No shadow root is created (`el.shadowRoot === null`)
3. `el._inputNode` resolves to an `HTMLInputElement`
4. Setting `el.modelValue = 'hello'` updates `el._inputNode.value`
5. `Required` validator sets `hasFeedbackFor` to include `'error'` when value is empty
6. `dirty` becomes `true` after dispatching an `input` event on the native input
7. `disabled` attribute propagates to the native input

### NO-GO criteria (ANY of these = STOP)

- SlotMixin throws because it expects `shadowRoot` to be non-null
- `_inputNode` returns `null` or `undefined` after `updateComplete`
- Validation never fires (validators array is set but `hasFeedbackFor` stays empty)
- `ScopedElementsMixin` throws about scoped custom element registries
- Happy-dom incompatibilities that prevent any test from running (if this happens, try jsdom before declaring NO-GO)
- Lit render conflicts with SlotMixin's child management (clobbered children on re-render)

### If NO-GO

**STOP. Do not proceed with any tasks after Chunk 0.** Report the specific failures and wait for guidance. Possible fallback approaches:

- **Approach B:** Override both `createRenderRoot()` and `render()`, manually position `_inputNode` in `firstUpdated()`.
- **Approach C:** Don't extend Lion. Use Lion's validators and interaction state mixins as standalone utilities applied to `WfElement`.
- **Approach D:** Abandon Lion entirely. Build form validation from scratch.

---

## Chunk 0: Lion Integration Spike (GO/NO-GO GATE)

### Task 1: Install @lion/ui

**Files:**
- Modify: `web/packages/ui/package.json`

- [ ] **Step 1: Add @lion/ui dependency**

Add `@lion/ui` to the `dependencies` section of `web/packages/ui/package.json`:

```json
"dependencies": {
  "lit": "^3.2.0",
  "@lion/ui": "^0.16.0"
}
```

- [ ] **Step 2: Install dependencies**

```bash
cd web/packages/ui && npm install
```

- [ ] **Step 3: Verify installation**

```bash
ls node_modules/@lion/ui/
```

Confirm that `form-core.js`, `input.js`, and other expected entry points exist.

---

### Task 2: Write Lion + Light DOM spike test

**Files:**
- Create: `web/packages/ui/tests/spike/lion-light-dom.test.ts`

- [ ] **Step 1: Create test directory**

```bash
mkdir -p web/packages/ui/tests/spike
```

- [ ] **Step 2: Write the spike test**

```typescript
// tests/spike/lion-light-dom.test.ts
//
// Self-contained spike: defines WfInputSpike inline, registers it,
// and verifies that Lion's form system works in light DOM.

import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import { html, nothing } from 'lit';
import { LionInput, Required } from '@lion/ui/form-core.js';

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

// --- Tests ---

describe('Lion + Light DOM Spike', () => {
  afterEach(cleanup);

  it('renders without shadow DOM', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect(el.shadowRoot).toBeNull();
    // The element should exist in the document
    expect(document.querySelector('wf-input-spike')).toBe(el);
  });

  it('resolves _inputNode to a native <input>', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect(el._inputNode).toBeInstanceOf(HTMLInputElement);
  });

  it('syncs modelValue to native input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.modelValue = 'hello';
    await el.updateComplete;
    expect(el._inputNode.value).toBe('hello');
  });

  it('Required validator sets hasFeedbackFor to include error', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.validators = [new Required()];
    el.modelValue = '';
    // Force interaction states so feedback shows
    el.touched = true;
    el.dirty = true;
    el.submitted = true;
    await el.updateComplete;
    // Lion's validation is async — wait for validate-complete event or poll
    await el.feedbackComplete;
    expect(el.hasFeedbackFor).toContain('error');
  });

  it('tracks dirty state after user input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect(el.dirty).toBe(false);
    // Simulate user typing
    el._inputNode.value = 'typed';
    el._inputNode.dispatchEvent(new Event('input', { bubbles: true }));
    await el.updateComplete;
    expect(el.dirty).toBe(true);
  });

  it('tracks touched state after blur', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    expect(el.touched).toBe(false);
    el._inputNode.dispatchEvent(new Event('focusin', { bubbles: true }));
    el._inputNode.dispatchEvent(new Event('focusout', { bubbles: true }));
    await el.updateComplete;
    expect(el.touched).toBe(true);
  });

  it('propagates disabled to native input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.disabled = true;
    await el.updateComplete;
    expect(el._inputNode.disabled).toBe(true);
  });

  it('propagates readonly to native input', async () => {
    const el = await fixture<WfInputSpike>('wf-input-spike');
    el.readOnly = true;
    await el.updateComplete;
    expect(el._inputNode.readOnly).toBe(true);
  });
});
```

- [ ] **Step 3: Run the spike test**

```bash
cd web/packages/ui && npx vitest run tests/spike/lion-light-dom.test.ts
```

- [ ] **Step 4: Evaluate results**

If **all tests pass**: **GO** — proceed to Chunk 1.

If **any test fails**: Record the exact failure. Consult the NO-GO criteria above. If the failure matches a NO-GO criterion, **STOP** and report. Do not proceed to Chunk 1.

If tests fail due to happy-dom limitations (not Lion issues), try switching the test to jsdom before declaring NO-GO:

```typescript
// Add at top of test file:
// @vitest-environment jsdom
```

---

## Chunk 1: Form Field Foundation

> **Prerequisite:** Chunk 0 passed GO.

### Task 3: Create shared form field CSS

**Files:**
- Create: `web/packages/ui/src/styles/form-field.css`
- Modify: `web/packages/ui/src/styles/components.css` (add `@import './form-field.css';`)

- [ ] **Step 1: Add form-field component tokens to components.css**

Add to the `:root` block in `components.css`:

```css
  --wf-field-label-size: var(--wf-text-sm);
  --wf-field-label-color: var(--wf-color-text-secondary);
  --wf-field-label-weight: var(--wf-weight-medium);
  --wf-field-label-gap: var(--wf-space-xs);
  --wf-field-help-size: var(--wf-text-xs);
  --wf-field-help-color: var(--wf-color-text-muted);
  --wf-field-feedback-size: var(--wf-text-xs);
  --wf-field-feedback-color: var(--wf-color-error);
  --wf-field-disabled-opacity: 0.5;
```

- [ ] **Step 2: Write form-field.css**

```css
/* form-field.css — Shared styles for all Lion-backed form components */

/* Base field layout */
.wf-field {
  display: block;
  font-family: var(--wf-font-sans);
  color: var(--wf-color-text);
}

/* Label */
.wf-field__label {
  display: block;
  font-size: var(--wf-field-label-size);
  font-weight: var(--wf-field-label-weight);
  color: var(--wf-field-label-color);
  margin-bottom: var(--wf-field-label-gap);
  cursor: pointer;
  user-select: none;
}

/* Input container — wraps the native input + prefix/suffix */
.wf-field__container {
  display: flex;
  align-items: center;
  width: 100%;
  min-height: var(--wf-input-height);
  padding: var(--wf-input-padding);
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-input-radius);
  box-sizing: border-box;
  transition: border-color 150ms ease;
}

/* Focus state — applied via Lion's focus tracking */
.wf-field--focused .wf-field__container,
.wf-field__container:focus-within {
  border-color: var(--wf-color-border-focus);
}

/* Error state */
.wf-field--error .wf-field__container {
  border-color: var(--wf-color-error);
}

/* Native input within container */
.wf-field__container input,
.wf-field__container textarea {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-family: inherit;
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
  padding: 0;
  margin: 0;
  width: 100%;
}

.wf-field__container input::placeholder,
.wf-field__container textarea::placeholder {
  color: var(--wf-color-text-muted);
}

/* Help text */
.wf-field__help-text {
  font-size: var(--wf-field-help-size);
  color: var(--wf-field-help-color);
  margin-top: var(--wf-space-xs);
}

/* Validation feedback */
.wf-field__feedback {
  font-size: var(--wf-field-feedback-size);
  color: var(--wf-field-feedback-color);
  margin-top: var(--wf-space-xs);
}

/* Disabled state */
.wf-field[disabled],
.wf-field--disabled {
  opacity: var(--wf-field-disabled-opacity);
  cursor: not-allowed;
  pointer-events: none;
}

/* Required indicator */
.wf-field__label--required::after {
  content: ' *';
  color: var(--wf-color-error);
}

/* Prefix/suffix slots */
.wf-field__prefix,
.wf-field__suffix {
  display: flex;
  align-items: center;
  color: var(--wf-color-text-muted);
  flex-shrink: 0;
}
.wf-field__prefix { margin-right: var(--wf-space-xs); }
.wf-field__suffix { margin-left: var(--wf-space-xs); }
```

- [ ] **Step 3: Import form-field.css in components.css**

Add `@import './form-field.css';` to `components.css`:

```css
@import './banner.css';
@import './toast.css';
@import './form-field.css';
```

---

### Task 4: Create WfField base class

**Files:**
- Create: `web/packages/ui/src/form/wf-field.ts`

This is the base class for all form components. It extends `LionField` (not `LionInput`) and provides the light DOM + CSS class foundation.

- [ ] **Step 1: Create form directory**

```bash
mkdir -p web/packages/ui/src/form
```

- [ ] **Step 2: Write wf-field.ts**

```typescript
// src/form/wf-field.ts
import { LionField } from '@lion/ui/form-core.js';
import { html, nothing, type TemplateResult } from 'lit';

/**
 * Base class for WorkFort form fields.
 * Extends LionField with light DOM rendering and WorkFort CSS classes.
 *
 * Subclasses override:
 * - `_inputTemplate()` for the input area
 * - `_wfFieldClass` for the component-specific CSS class (e.g., 'wf-input')
 */
export class WfField extends LionField {
  /**
   * Render into light DOM instead of shadow DOM.
   */
  createRenderRoot(): this {
    return this;
  }

  /**
   * CSS class added to the host element (override in subclasses).
   */
  protected get _wfFieldClass(): string {
    return '';
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field');
    if (this._wfFieldClass) {
      this.classList.add(this._wfFieldClass);
    }
  }

  /**
   * Override Lion's render to use WorkFort CSS classes instead of <slot> elements.
   */
  render(): TemplateResult {
    return html`
      ${this._labelTemplate()}
      ${this._inputContainerTemplate()}
      ${this._helpTextTemplate()}
      ${this._feedbackTemplate()}
    `;
  }

  protected _labelTemplate(): TemplateResult | typeof nothing {
    if (!this.label) return nothing;
    const requiredClass = this.validators?.some(
      (v: { constructor: { validatorName?: string } }) => v.constructor.validatorName === 'Required'
    )
      ? ' wf-field__label--required'
      : '';
    return html`
      <label
        class="wf-field__label${requiredClass}"
        @click=${() => this._inputNode?.focus()}
      >${this.label}</label>
    `;
  }

  protected _inputContainerTemplate(): TemplateResult {
    return html`
      <div class="wf-field__container">
        ${this._prefixTemplate()}
        ${this._inputTemplate()}
        ${this._suffixTemplate()}
      </div>
    `;
  }

  /**
   * Override in subclasses to provide the input element.
   * Default: render a slot for the input node created by SlotMixin.
   */
  protected _inputTemplate(): TemplateResult {
    return html`<slot name="input"></slot>`;
  }

  protected _prefixTemplate(): TemplateResult | typeof nothing {
    return nothing;
  }

  protected _suffixTemplate(): TemplateResult | typeof nothing {
    return nothing;
  }

  protected _helpTextTemplate(): TemplateResult | typeof nothing {
    if (!(this as unknown as { helpText?: string }).helpText) return nothing;
    return html`
      <div class="wf-field__help-text">
        ${(this as unknown as { helpText: string }).helpText}
      </div>
    `;
  }

  protected _feedbackTemplate(): TemplateResult | typeof nothing {
    if (!this.showsFeedbackFor?.length) return nothing;
    return html`
      <div class="wf-field__feedback">
        <slot name="feedback"></slot>
      </div>
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    // Toggle error class based on feedback state
    this.classList.toggle(
      'wf-field--error',
      this.showsFeedbackFor?.includes('error') ?? false,
    );
    // Toggle focused class
    this.classList.toggle('wf-field--focused', this.focused ?? false);
  }
}
```

**Note:** This class is intentionally NOT registered as a custom element — it's an abstract base. Subclasses register themselves.

---

### Task 5: Create WfInput component (replaces wf-text-input)

**Files:**
- Create: `web/packages/ui/src/form/wf-input.ts`
- Create: `web/packages/ui/tests/form/wf-input.test.ts`

- [ ] **Step 1: Write wf-input.ts**

```typescript
// src/form/wf-input.ts
import { LionInput } from '@lion/ui/input.js';
import { html, nothing, type TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';

/**
 * Text input component backed by Lion's form system.
 *
 * @element wf-input
 *
 * @prop {string} label - Input label text
 * @prop {string} helpText - Help text displayed below the input
 * @prop {string} type - Input type (text, email, password, number, tel, url, search)
 * @prop {string} placeholder - Placeholder text
 * @prop {boolean} disabled - Disabled state
 * @prop {boolean} readOnly - Read-only state
 * @prop {Validator[]} validators - Array of Lion validators
 *
 * @fires wf-input - On every keystroke (detail: { value: string })
 * @fires wf-change - On blur after value change (detail: { value: string })
 */
export class WfInput extends LionInput {
  @property({ type: String, attribute: 'help-text' })
  helpText = '';

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-input');
    // Forward Lion events as wf-* custom events
    this.addEventListener('model-value-changed', this._onModelValueChanged);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('model-value-changed', this._onModelValueChanged);
  }

  render(): TemplateResult {
    return html`
      ${this._labelTemplate()}
      ${this._inputContainerTemplate()}
      ${this._helpTextTemplate()}
      ${this._feedbackTemplate()}
    `;
  }

  protected _labelTemplate(): TemplateResult | typeof nothing {
    if (!this.label) return nothing;
    const isRequired = this.validators?.some(
      (v: { constructor: { validatorName?: string } }) =>
        (v.constructor as { validatorName?: string }).validatorName === 'Required',
    );
    return html`
      <label
        class="wf-field__label${isRequired ? ' wf-field__label--required' : ''}"
        @click=${() => this._inputNode?.focus()}
      >${this.label}</label>
    `;
  }

  protected _inputContainerTemplate(): TemplateResult {
    return html`
      <div class="wf-field__container">
        <slot name="prefix"></slot>
        <slot name="input"></slot>
        <slot name="suffix"></slot>
      </div>
    `;
  }

  protected _helpTextTemplate(): TemplateResult | typeof nothing {
    if (!this.helpText) return nothing;
    return html`<div class="wf-field__help-text">${this.helpText}</div>`;
  }

  protected _feedbackTemplate(): TemplateResult | typeof nothing {
    if (!this.showsFeedbackFor?.length) return nothing;
    return html`
      <div class="wf-field__feedback">
        <slot name="feedback"></slot>
      </div>
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle(
      'wf-field--error',
      this.showsFeedbackFor?.includes('error') ?? false,
    );
  }

  private _onModelValueChanged = (): void => {
    this.dispatchEvent(
      new CustomEvent('wf-change', {
        bubbles: true,
        composed: true,
        detail: { value: this.modelValue },
      }),
    );
  };
}

customElements.define('wf-input', WfInput);

declare global {
  interface HTMLElementTagNameMap {
    'wf-input': WfInput;
  }
}
```

- [ ] **Step 2: Write wf-input tests**

```bash
mkdir -p web/packages/ui/tests/form
```

```typescript
// tests/form/wf-input.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../../src/form/wf-input.js';
import type { WfInput } from '../../../src/form/wf-input.js';
import { Required, MinLength } from '@lion/ui/form-core.js';

describe('wf-input', () => {
  afterEach(cleanup);

  it('renders in light DOM with correct CSS classes', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el.shadowRoot).toBeNull();
    expect(el.classList.contains('wf-field')).toBe(true);
    expect(el.classList.contains('wf-input')).toBe(true);
  });

  it('renders label when provided', async () => {
    const el = await fixture<WfInput>('wf-input', { label: 'Email' });
    await el.updateComplete;
    const label = el.querySelector('.wf-field__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toContain('Email');
  });

  it('resolves _inputNode', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el._inputNode).toBeInstanceOf(HTMLInputElement);
  });

  it('syncs modelValue to native input', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.modelValue = 'test@example.com';
    await el.updateComplete;
    expect(el._inputNode.value).toBe('test@example.com');
  });

  it('validates with Required', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.validators = [new Required()];
    el.modelValue = '';
    el.touched = true;
    el.dirty = true;
    el.submitted = true;
    await el.updateComplete;
    await el.feedbackComplete;
    expect(el.hasFeedbackFor).toContain('error');
    expect(el.classList.contains('wf-field--error')).toBe(true);
  });

  it('validates with MinLength', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.validators = [new MinLength(3)];
    el.modelValue = 'ab';
    el.submitted = true;
    await el.updateComplete;
    await el.feedbackComplete;
    expect(el.hasFeedbackFor).toContain('error');
  });

  it('clears error when value becomes valid', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.validators = [new Required()];
    el.modelValue = '';
    el.submitted = true;
    await el.updateComplete;
    await el.feedbackComplete;
    expect(el.hasFeedbackFor).toContain('error');

    el.modelValue = 'valid';
    await el.updateComplete;
    await el.feedbackComplete;
    expect(el.hasFeedbackFor).not.toContain('error');
  });

  it('tracks dirty state', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el.dirty).toBe(false);
    el._inputNode.value = 'typed';
    el._inputNode.dispatchEvent(new Event('input', { bubbles: true }));
    await el.updateComplete;
    expect(el.dirty).toBe(true);
  });

  it('tracks touched state', async () => {
    const el = await fixture<WfInput>('wf-input');
    expect(el.touched).toBe(false);
    el._inputNode.dispatchEvent(new Event('focusin', { bubbles: true }));
    el._inputNode.dispatchEvent(new Event('focusout', { bubbles: true }));
    await el.updateComplete;
    expect(el.touched).toBe(true);
  });

  it('fires wf-change on model value change', async () => {
    const el = await fixture<WfInput>('wf-input');
    let detail: { value: string } | null = null;
    el.addEventListener('wf-change', ((e: CustomEvent) => {
      detail = e.detail;
    }) as EventListener);
    el.modelValue = 'changed';
    await el.updateComplete;
    expect(detail).toEqual({ value: 'changed' });
  });

  it('propagates disabled', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.disabled = true;
    await el.updateComplete;
    expect(el._inputNode.disabled).toBe(true);
  });

  it('propagates readOnly', async () => {
    const el = await fixture<WfInput>('wf-input');
    el.readOnly = true;
    await el.updateComplete;
    expect(el._inputNode.readOnly).toBe(true);
  });

  it('renders help text', async () => {
    const el = await fixture<WfInput>('wf-input', { 'help-text': 'Enter your email' });
    await el.updateComplete;
    const helpText = el.querySelector('.wf-field__help-text');
    expect(helpText).not.toBeNull();
    expect(helpText!.textContent).toContain('Enter your email');
  });

  it('shows required indicator in label', async () => {
    const el = await fixture<WfInput>('wf-input', { label: 'Name' });
    el.validators = [new Required()];
    await el.updateComplete;
    const label = el.querySelector('.wf-field__label');
    expect(label?.classList.contains('wf-field__label--required')).toBe(true);
  });
});
```

- [ ] **Step 3: Run tests**

```bash
cd web/packages/ui && npx vitest run tests/form/wf-input.test.ts
```

---

### Task 6: Register wf-input in index.ts and deprecate wf-text-input

**Files:**
- Modify: `web/packages/ui/src/index.ts`
- Modify: `web/packages/ui/src/components/text-input.ts` (add deprecation notice)

- [ ] **Step 1: Add wf-input export to index.ts**

Add to `web/packages/ui/src/index.ts`:

```typescript
import './form/wf-input.js';
// ...
export { WfInput } from './form/wf-input.js';
```

- [ ] **Step 2: Add deprecation comment to text-input.ts**

Add a JSDoc `@deprecated` tag to `WfTextInput`:

```typescript
/**
 * @deprecated Use `wf-input` (WfInput) instead. This component will be removed in v1.0.
 */
export class WfTextInput extends WfElement {
```

- [ ] **Step 3: Verify build**

```bash
cd web/packages/ui && npm run build
```

---

## Chunk 2: Text-Like Form Components

### Task 7: Create wf-textarea

**Files:**
- Create: `web/packages/ui/src/form/wf-textarea.ts`
- Create: `web/packages/ui/src/styles/textarea.css`
- Create: `web/packages/ui/tests/form/wf-textarea.test.ts`
- Modify: `web/packages/ui/src/styles/components.css` (add `@import './textarea.css';`)
- Modify: `web/packages/ui/src/index.ts` (add export)

- [ ] **Step 1: Write textarea.css**

```css
/* textarea.css */
.wf-textarea .wf-field__container {
  min-height: calc(var(--wf-input-height) * 3);
  align-items: flex-start;
}

.wf-textarea textarea {
  resize: vertical;
  min-height: calc(var(--wf-input-height) * 2);
  line-height: var(--wf-leading-normal);
}
```

- [ ] **Step 2: Write wf-textarea.ts**

Follow the same pattern as `wf-input.ts`, but extend `LionTextarea` from `@lion/ui/textarea.js`. Key differences:
- Extends `LionTextarea` instead of `LionInput`
- `_wfFieldClass` returns `'wf-textarea'`
- `LionTextarea` auto-resizes — verify this works in light DOM

```typescript
// src/form/wf-textarea.ts
import { LionTextarea } from '@lion/ui/textarea.js';
import { html, nothing, type TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';

export class WfTextarea extends LionTextarea {
  @property({ type: String, attribute: 'help-text' })
  helpText = '';

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-textarea');
    this.addEventListener('model-value-changed', this._onModelValueChanged);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('model-value-changed', this._onModelValueChanged);
  }

  render(): TemplateResult {
    return html`
      ${this._wfLabelTemplate()}
      <div class="wf-field__container">
        <slot name="input"></slot>
      </div>
      ${this._wfHelpTextTemplate()}
      ${this._wfFeedbackTemplate()}
    `;
  }

  protected _wfLabelTemplate(): TemplateResult | typeof nothing {
    if (!this.label) return nothing;
    return html`<label class="wf-field__label" @click=${() => this._inputNode?.focus()}>${this.label}</label>`;
  }

  protected _wfHelpTextTemplate(): TemplateResult | typeof nothing {
    if (!this.helpText) return nothing;
    return html`<div class="wf-field__help-text">${this.helpText}</div>`;
  }

  protected _wfFeedbackTemplate(): TemplateResult | typeof nothing {
    if (!this.showsFeedbackFor?.length) return nothing;
    return html`<div class="wf-field__feedback"><slot name="feedback"></slot></div>`;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-field--error', this.showsFeedbackFor?.includes('error') ?? false);
  }

  private _onModelValueChanged = (): void => {
    this.dispatchEvent(new CustomEvent('wf-change', {
      bubbles: true, composed: true, detail: { value: this.modelValue },
    }));
  };
}

customElements.define('wf-textarea', WfTextarea);

declare global {
  interface HTMLElementTagNameMap {
    'wf-textarea': WfTextarea;
  }
}
```

- [ ] **Step 3: Write tests**

Tests should cover: light DOM rendering, `_inputNode` resolves to `<textarea>`, modelValue sync, Required validation, dirty/touched states, auto-resize behavior (if supported in happy-dom), disabled/readonly.

- [ ] **Step 4: Import in components.css and index.ts**

---

### Task 8: Create wf-select

**Files:**
- Create: `web/packages/ui/src/form/wf-select.ts`
- Create: `web/packages/ui/src/styles/select.css`
- Create: `web/packages/ui/tests/form/wf-select.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Determine which Lion select to use**

Lion offers two:
- `LionSelect` — wraps native `<select>`. Simpler, less customizable.
- `LionSelectRich` — custom dropdown with `OverlayMixin`. Fully customizable, but uses overlays.

Use `LionSelectRich` for the full-featured version. This requires overlay setup.

- [ ] **Step 2: Write select.css**

```css
/* select.css */
.wf-select .wf-field__container {
  cursor: pointer;
  position: relative;
}

.wf-select__invoker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-height: var(--wf-input-height);
  padding: var(--wf-input-padding);
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-input-radius);
  cursor: pointer;
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
}

.wf-select__invoker:focus {
  border-color: var(--wf-color-border-focus);
  outline: none;
}

.wf-select__listbox {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  max-height: 200px;
  overflow-y: auto;
  background: var(--wf-color-bg-elevated);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-input-radius);
  margin-top: var(--wf-space-xs);
  z-index: var(--wf-z-dropdown, 100);
  scrollbar-width: thin;
}

.wf-select__option {
  padding: var(--wf-input-padding) var(--wf-button-padding-x);
  cursor: pointer;
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
}

.wf-select__option:hover,
.wf-select__option--focused {
  background: var(--wf-color-bg-secondary);
}

.wf-select__option--selected {
  font-weight: var(--wf-weight-medium);
}

.wf-select__option--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
```

- [ ] **Step 3: Write wf-select.ts**

Extend `LionSelectRich`, override `createRenderRoot()`. The select needs:
- An invoker button showing the selected value
- A listbox overlay with options
- Keyboard navigation (handled by Lion)
- ARIA attributes (handled by Lion)

This component is more complex than input/textarea because of the overlay. If `LionSelectRich` doesn't work in light DOM, fall back to wrapping a native `<select>` with `LionSelect`.

```typescript
// src/form/wf-select.ts
import { LionSelectRich } from '@lion/ui/select-rich.js';
import { html, nothing, type TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';

export class WfSelect extends LionSelectRich {
  @property({ type: String, attribute: 'help-text' })
  helpText = '';

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-select');
  }

  // Note: LionSelectRich has a complex render() with overlay management.
  // We may need to keep Lion's render() and only override CSS.
  // The spike results will inform whether a full render override is feasible.
  // If not, we apply CSS classes via updated() and connectedCallback() instead.

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-field--error', this.showsFeedbackFor?.includes('error') ?? false);
  }
}

customElements.define('wf-select', WfSelect);

declare global {
  interface HTMLElementTagNameMap {
    'wf-select': WfSelect;
  }
}
```

**Implementation note:** `LionSelectRich` uses `OverlayMixin` and `OverlayController` for the dropdown. This is the first overlay component. The pattern established here will be reused by `wf-combobox` and `wf-date-picker`. If the overlay system doesn't work in light DOM, we need to evaluate whether to:
1. Keep shadow DOM just for overlay components (hybrid approach)
2. Use a simpler CSS-based dropdown (no Lion overlay)
3. Use a different positioning library (e.g., Floating UI)

- [ ] **Step 4: Write tests**

Tests: light DOM rendering, option selection, keyboard navigation (arrow keys, enter, escape), modelValue sync, validation, disabled state, opening/closing the dropdown.

- [ ] **Step 5: Register in index.ts and components.css**

---

## Chunk 3: Choice Components

### Task 9: Create wf-checkbox and wf-checkbox-group

**Files:**
- Create: `web/packages/ui/src/form/wf-checkbox.ts`
- Create: `web/packages/ui/src/form/wf-checkbox-group.ts`
- Create: `web/packages/ui/src/styles/checkbox.css`
- Create: `web/packages/ui/tests/form/wf-checkbox.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write checkbox.css**

```css
/* checkbox.css */
.wf-checkbox {
  display: flex;
  align-items: flex-start;
  gap: var(--wf-space-sm);
  cursor: pointer;
}

.wf-checkbox__control {
  position: relative;
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  margin-top: 0.125rem; /* Align with text baseline */
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-xs);
  background: var(--wf-color-bg);
  transition: background 150ms ease, border-color 150ms ease;
}

.wf-checkbox--checked .wf-checkbox__control {
  background: var(--wf-color-accent);
  border-color: var(--wf-color-accent);
}

.wf-checkbox--indeterminate .wf-checkbox__control {
  background: var(--wf-color-accent);
  border-color: var(--wf-color-accent);
}

/* Checkmark icon via CSS */
.wf-checkbox--checked .wf-checkbox__control::after {
  content: '';
  position: absolute;
  left: 3px;
  top: 1px;
  width: 5px;
  height: 8px;
  border: solid var(--wf-color-text-on-accent);
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

/* Indeterminate dash */
.wf-checkbox--indeterminate .wf-checkbox__control::after {
  content: '';
  position: absolute;
  left: 3px;
  top: 6px;
  width: 8px;
  height: 2px;
  background: var(--wf-color-text-on-accent);
}

.wf-checkbox__label {
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
  user-select: none;
}

.wf-checkbox[disabled] {
  opacity: var(--wf-field-disabled-opacity);
  cursor: not-allowed;
}

/* Checkbox group */
.wf-checkbox-group {
  display: flex;
  flex-direction: column;
  gap: var(--wf-space-sm);
}
```

- [ ] **Step 2: Write wf-checkbox.ts**

Extend `LionCheckboxIndeterminate` (or `LionCheckbox`) from `@lion/ui/checkbox-group.js`:

```typescript
// src/form/wf-checkbox.ts
import { LionCheckboxIndeterminate } from '@lion/ui/checkbox-group.js';
import { html, type TemplateResult } from 'lit';

export class WfCheckbox extends LionCheckboxIndeterminate {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-checkbox');
  }

  render(): TemplateResult {
    return html`
      <div class="wf-checkbox__control">
        <slot name="input"></slot>
      </div>
      <label class="wf-checkbox__label" @click=${() => this._inputNode?.click()}>
        ${this.label}
      </label>
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-checkbox--checked', this.checked ?? false);
    this.classList.toggle('wf-checkbox--indeterminate', this.indeterminate ?? false);
  }
}

customElements.define('wf-checkbox', WfCheckbox);

declare global {
  interface HTMLElementTagNameMap {
    'wf-checkbox': WfCheckbox;
  }
}
```

- [ ] **Step 3: Write wf-checkbox-group.ts**

Extend `LionCheckboxGroup` from `@lion/ui/checkbox-group.js`:

```typescript
// src/form/wf-checkbox-group.ts
import { LionCheckboxGroup } from '@lion/ui/checkbox-group.js';
import { html, nothing, type TemplateResult } from 'lit';

export class WfCheckboxGroup extends LionCheckboxGroup {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-checkbox-group');
  }

  render(): TemplateResult {
    return html`
      ${this.label ? html`<div class="wf-field__label">${this.label}</div>` : nothing}
      <div class="wf-checkbox-group__items" role="group">
        <slot></slot>
      </div>
      ${this.showsFeedbackFor?.length
        ? html`<div class="wf-field__feedback"><slot name="feedback"></slot></div>`
        : nothing}
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-field--error', this.showsFeedbackFor?.includes('error') ?? false);
  }
}

customElements.define('wf-checkbox-group', WfCheckboxGroup);

declare global {
  interface HTMLElementTagNameMap {
    'wf-checkbox-group': WfCheckboxGroup;
  }
}
```

- [ ] **Step 4: Write tests**

Tests: checked/unchecked toggle, indeterminate state, group validation (e.g., Required on group = at least one checked), disabled state, keyboard toggle (space key), wf-change events, label click toggles.

- [ ] **Step 5: Register in index.ts and components.css**

---

### Task 10: Create wf-radio and wf-radio-group

**Files:**
- Create: `web/packages/ui/src/form/wf-radio.ts`
- Create: `web/packages/ui/src/form/wf-radio-group.ts`
- Create: `web/packages/ui/src/styles/radio.css`
- Create: `web/packages/ui/tests/form/wf-radio.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write radio.css**

```css
/* radio.css */
.wf-radio {
  display: flex;
  align-items: flex-start;
  gap: var(--wf-space-sm);
  cursor: pointer;
}

.wf-radio__control {
  position: relative;
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  margin-top: 0.125rem;
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-full);
  background: var(--wf-color-bg);
  transition: border-color 150ms ease;
}

.wf-radio--checked .wf-radio__control {
  border-color: var(--wf-color-accent);
}

.wf-radio--checked .wf-radio__control::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--wf-color-accent);
}

.wf-radio__label {
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
  user-select: none;
}

.wf-radio[disabled] {
  opacity: var(--wf-field-disabled-opacity);
  cursor: not-allowed;
}

.wf-radio-group {
  display: flex;
  flex-direction: column;
  gap: var(--wf-space-sm);
}
```

- [ ] **Step 2: Write wf-radio.ts**

Same pattern as checkbox — extend `LionRadio` from `@lion/ui/radio-group.js`:

```typescript
// src/form/wf-radio.ts
import { LionRadio } from '@lion/ui/radio-group.js';
import { html, type TemplateResult } from 'lit';

export class WfRadio extends LionRadio {
  createRenderRoot(): this { return this; }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-radio');
  }

  render(): TemplateResult {
    return html`
      <div class="wf-radio__control">
        <slot name="input"></slot>
      </div>
      <label class="wf-radio__label" @click=${() => this._inputNode?.click()}>
        ${this.label}
      </label>
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-radio--checked', this.checked ?? false);
  }
}

customElements.define('wf-radio', WfRadio);

declare global {
  interface HTMLElementTagNameMap { 'wf-radio': WfRadio; }
}
```

- [ ] **Step 3: Write wf-radio-group.ts**

Same pattern as `WfCheckboxGroup` — extend `LionRadioGroup`.

- [ ] **Step 4: Write tests**

Tests: radio selection, group mutual exclusivity, Required validation on group, keyboard navigation (arrow keys cycle options), disabled state.

- [ ] **Step 5: Register in index.ts and components.css**

---

### Task 11: Create wf-toggle

**Files:**
- Create: `web/packages/ui/src/form/wf-toggle.ts`
- Create: `web/packages/ui/src/styles/toggle.css`
- Create: `web/packages/ui/tests/form/wf-toggle.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

Lion doesn't have a dedicated toggle/switch component. Build it using `ChoiceInputMixin` from `@lion/ui/form-core.js` applied to `WfElement`, or extend `LionCheckbox` and restyle.

- [ ] **Step 1: Write toggle.css**

```css
/* toggle.css */
.wf-toggle {
  display: flex;
  align-items: center;
  gap: var(--wf-space-sm);
  cursor: pointer;
}

.wf-toggle__track {
  position: relative;
  width: 2.25rem;
  height: 1.25rem;
  flex-shrink: 0;
  border-radius: var(--wf-radius-full);
  background: var(--wf-color-border-strong);
  transition: background 150ms ease;
}

.wf-toggle--checked .wf-toggle__track {
  background: var(--wf-color-accent);
}

.wf-toggle__thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 1rem;
  height: 1rem;
  border-radius: 50%;
  background: var(--wf-color-bg);
  transition: transform 150ms ease;
}

.wf-toggle--checked .wf-toggle__thumb {
  transform: translateX(1rem);
}

.wf-toggle__label {
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
  user-select: none;
}

.wf-toggle[disabled] {
  opacity: var(--wf-field-disabled-opacity);
  cursor: not-allowed;
}

/* Hide the native checkbox */
.wf-toggle input[type="checkbox"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
  pointer-events: none;
}
```

- [ ] **Step 2: Write wf-toggle.ts**

Extend `LionCheckbox` and override the visual template:

```typescript
// src/form/wf-toggle.ts
import { LionCheckbox } from '@lion/ui/checkbox-group.js';
import { html, type TemplateResult } from 'lit';

export class WfToggle extends LionCheckbox {
  createRenderRoot(): this { return this; }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-toggle');
    this.setAttribute('role', 'switch');
  }

  render(): TemplateResult {
    return html`
      <slot name="input"></slot>
      <div class="wf-toggle__track" @click=${() => this._inputNode?.click()}>
        <div class="wf-toggle__thumb"></div>
      </div>
      <label class="wf-toggle__label" @click=${() => this._inputNode?.click()}>
        ${this.label}
      </label>
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-toggle--checked', this.checked ?? false);
    this.setAttribute('aria-checked', String(this.checked ?? false));
  }
}

customElements.define('wf-toggle', WfToggle);

declare global {
  interface HTMLElementTagNameMap { 'wf-toggle': WfToggle; }
}
```

- [ ] **Step 3: Write tests**

Tests: toggle on/off, `role="switch"`, `aria-checked` updates, keyboard toggle (space/enter), disabled state, wf-change event.

- [ ] **Step 4: Register in index.ts and components.css**

---

## Chunk 4: Advanced Components

### Task 12: Create wf-slider

**Files:**
- Create: `web/packages/ui/src/form/wf-slider.ts`
- Create: `web/packages/ui/src/styles/slider.css`
- Create: `web/packages/ui/tests/form/wf-slider.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write slider.css**

```css
/* slider.css */
.wf-slider .wf-field__container {
  flex-direction: column;
  border: none;
  padding: 0;
}

.wf-slider__track-container {
  position: relative;
  width: 100%;
  height: 1.5rem;
  display: flex;
  align-items: center;
}

.wf-slider__track {
  width: 100%;
  height: 4px;
  border-radius: var(--wf-radius-full);
  background: var(--wf-color-border);
}

.wf-slider__fill {
  height: 100%;
  border-radius: var(--wf-radius-full);
  background: var(--wf-color-accent);
}

.wf-slider__value {
  font-family: var(--wf-font-mono);
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text-secondary);
  margin-left: var(--wf-space-sm);
  min-width: 3ch;
  text-align: right;
}

/* Style the native range input */
.wf-slider input[type="range"] {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
  margin: 0;
}
```

- [ ] **Step 2: Write wf-slider.ts**

Extend `LionInputRange` from `@lion/ui/input-range.js`:

```typescript
// src/form/wf-slider.ts
import { LionInputRange } from '@lion/ui/input-range.js';
import { html, nothing, type TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';

export class WfSlider extends LionInputRange {
  @property({ type: String, attribute: 'help-text' })
  helpText = '';

  @property({ type: Boolean, attribute: 'show-value' })
  showValue = false;

  createRenderRoot(): this { return this; }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-slider');
  }

  render(): TemplateResult {
    const pct = this._inputNode
      ? ((Number(this.modelValue) - Number(this.min)) / (Number(this.max) - Number(this.min))) * 100
      : 0;

    return html`
      ${this.label ? html`<label class="wf-field__label">${this.label}</label>` : nothing}
      <div class="wf-field__container">
        <div class="wf-slider__track-container">
          <div class="wf-slider__track">
            <div class="wf-slider__fill" style="width: ${pct}%"></div>
          </div>
          <slot name="input"></slot>
        </div>
        ${this.showValue
          ? html`<span class="wf-slider__value">${this.modelValue}</span>`
          : nothing}
      </div>
      ${this.helpText ? html`<div class="wf-field__help-text">${this.helpText}</div>` : nothing}
    `;
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-field--error', this.showsFeedbackFor?.includes('error') ?? false);
  }
}

customElements.define('wf-slider', WfSlider);

declare global {
  interface HTMLElementTagNameMap { 'wf-slider': WfSlider; }
}
```

- [ ] **Step 3: Write tests**

Tests: value setting via modelValue, min/max boundaries, step increments, keyboard (arrow keys), disabled state.

- [ ] **Step 4: Register in index.ts and components.css**

---

### Task 13: Create wf-combobox

**Files:**
- Create: `web/packages/ui/src/form/wf-combobox.ts`
- Create: `web/packages/ui/src/styles/combobox.css`
- Create: `web/packages/ui/tests/form/wf-combobox.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write combobox.css**

Reuse select dropdown styles for the listbox. Add typeahead-specific styles:

```css
/* combobox.css */
.wf-combobox .wf-field__container {
  position: relative;
}

.wf-combobox__listbox {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  max-height: 200px;
  overflow-y: auto;
  background: var(--wf-color-bg-elevated);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-input-radius);
  margin-top: var(--wf-space-xs);
  z-index: var(--wf-z-dropdown, 100);
  display: none;
}

.wf-combobox--open .wf-combobox__listbox {
  display: block;
}

.wf-combobox__option {
  padding: var(--wf-input-padding) var(--wf-button-padding-x);
  cursor: pointer;
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
}

.wf-combobox__option:hover,
.wf-combobox__option--focused {
  background: var(--wf-color-bg-secondary);
}

.wf-combobox__option--selected {
  font-weight: var(--wf-weight-medium);
}

.wf-combobox__no-results {
  padding: var(--wf-input-padding) var(--wf-button-padding-x);
  color: var(--wf-color-text-muted);
  font-size: var(--wf-text-sm);
}
```

- [ ] **Step 2: Write wf-combobox.ts**

Extend `LionCombobox` from `@lion/ui/combobox.js`. This uses `OverlayMixin` for the dropdown. Same overlay concerns as `wf-select` — if Lion's overlay system doesn't work in light DOM, may need a CSS-only dropdown.

- [ ] **Step 3: Write tests**

Tests: typing filters options, selecting an option sets modelValue, keyboard navigation (arrow down opens, arrow keys navigate, enter selects, escape closes), no-results message, autocomplete attribute.

- [ ] **Step 4: Register in index.ts and components.css**

---

### Task 14: Create wf-date-picker

**Files:**
- Create: `web/packages/ui/src/form/wf-date-picker.ts`
- Create: `web/packages/ui/src/styles/date-picker.css`
- Create: `web/packages/ui/tests/form/wf-date-picker.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write date-picker.css**

Calendar grid styles, month/year navigation, day cell states (today, selected, disabled, range):

```css
/* date-picker.css */
.wf-date-picker .wf-field__container {
  position: relative;
}

.wf-date-picker__calendar {
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: var(--wf-space-xs);
  background: var(--wf-color-bg-elevated);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-input-radius);
  padding: var(--wf-space-md);
  z-index: var(--wf-z-dropdown, 100);
  display: none;
}

.wf-date-picker--open .wf-date-picker__calendar {
  display: block;
}

.wf-date-picker__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--wf-space-sm);
}

.wf-date-picker__nav-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--wf-color-text);
  padding: var(--wf-space-xs);
  border-radius: var(--wf-radius-sm);
}

.wf-date-picker__nav-btn:hover {
  background: var(--wf-color-bg-secondary);
}

.wf-date-picker__grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 1px;
  text-align: center;
}

.wf-date-picker__day-header {
  font-size: var(--wf-text-xs);
  color: var(--wf-color-text-muted);
  padding: var(--wf-space-xs);
  font-weight: var(--wf-weight-medium);
}

.wf-date-picker__day {
  padding: var(--wf-space-xs);
  border-radius: var(--wf-radius-sm);
  cursor: pointer;
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text);
}

.wf-date-picker__day:hover {
  background: var(--wf-color-bg-secondary);
}

.wf-date-picker__day--today {
  font-weight: var(--wf-weight-semibold);
  border: 1px solid var(--wf-color-border-strong);
}

.wf-date-picker__day--selected {
  background: var(--wf-color-accent);
  color: var(--wf-color-text-on-accent);
}

.wf-date-picker__day--disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.wf-date-picker__day--outside {
  color: var(--wf-color-text-muted);
}
```

- [ ] **Step 2: Write wf-date-picker.ts**

Extend `LionInputDatepicker` from `@lion/ui/input-datepicker.js`. Lion's date picker uses `OverlayMixin` for the calendar popup and provides:
- Date parsing/formatting via `FormatMixin`
- Calendar navigation
- Min/max date constraints
- Localization

This is the most complex component. If Lion's date picker doesn't work in light DOM, consider:
1. Using only `LionInput` with Lion's date validators and building a custom calendar overlay
2. Using a third-party date picker and wrapping it

- [ ] **Step 3: Write tests**

Tests: date selection from calendar, manual text entry with parsing, min/max date validation, modelValue is a Date object, format display, keyboard navigation within calendar, opening/closing calendar overlay.

- [ ] **Step 4: Register in index.ts and components.css**

---

### Task 15: Create wf-file-upload

**Files:**
- Create: `web/packages/ui/src/form/wf-file-upload.ts`
- Create: `web/packages/ui/src/styles/file-upload.css`
- Create: `web/packages/ui/tests/form/wf-file-upload.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write file-upload.css**

```css
/* file-upload.css */
.wf-file-upload .wf-field__container {
  flex-direction: column;
  align-items: stretch;
  padding: var(--wf-space-md);
  border-style: dashed;
  cursor: pointer;
  text-align: center;
  min-height: 5rem;
  justify-content: center;
}

.wf-file-upload--dragover .wf-field__container {
  border-color: var(--wf-color-accent);
  background: var(--wf-color-bg-secondary);
}

.wf-file-upload__prompt {
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text-secondary);
}

.wf-file-upload__prompt strong {
  color: var(--wf-color-accent);
}

.wf-file-upload__file-list {
  margin-top: var(--wf-space-sm);
  display: flex;
  flex-direction: column;
  gap: var(--wf-space-xs);
}

.wf-file-upload__file {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--wf-space-xs) var(--wf-space-sm);
  background: var(--wf-color-bg-secondary);
  border-radius: var(--wf-radius-sm);
  font-size: var(--wf-text-sm);
}

.wf-file-upload__file-name {
  color: var(--wf-color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wf-file-upload__file-size {
  color: var(--wf-color-text-muted);
  flex-shrink: 0;
  margin-left: var(--wf-space-sm);
}

.wf-file-upload__remove {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--wf-color-text-muted);
  padding: var(--wf-space-xs);
  flex-shrink: 0;
}

.wf-file-upload__remove:hover {
  color: var(--wf-color-error);
}

/* Hide native file input */
.wf-file-upload input[type="file"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
  pointer-events: none;
}
```

- [ ] **Step 2: Write wf-file-upload.ts**

Extend `LionInputFile` from `@lion/ui/input-file.js`:

```typescript
// src/form/wf-file-upload.ts
import { LionInputFile } from '@lion/ui/input-file.js';
import { html, nothing, type TemplateResult } from 'lit';
import { property } from 'lit/decorators.js';

export class WfFileUpload extends LionInputFile {
  @property({ type: String, attribute: 'help-text' })
  helpText = '';

  @property({ type: String })
  accept = '';

  @property({ type: Boolean })
  multiple = false;

  createRenderRoot(): this { return this; }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-field', 'wf-file-upload');
  }

  render(): TemplateResult {
    return html`
      ${this.label ? html`<label class="wf-field__label">${this.label}</label>` : nothing}
      <div
        class="wf-field__container"
        @dragover=${this._onDragOver}
        @dragleave=${this._onDragLeave}
        @drop=${this._onDrop}
        @click=${() => this._inputNode?.click()}
      >
        <slot name="input"></slot>
        <div class="wf-file-upload__prompt">
          <strong>Click to upload</strong> or drag and drop
        </div>
      </div>
      ${this.helpText ? html`<div class="wf-field__help-text">${this.helpText}</div>` : nothing}
      ${this.showsFeedbackFor?.length
        ? html`<div class="wf-field__feedback"><slot name="feedback"></slot></div>`
        : nothing}
    `;
  }

  private _onDragOver = (e: DragEvent): void => {
    e.preventDefault();
    this.classList.add('wf-file-upload--dragover');
  };

  private _onDragLeave = (): void => {
    this.classList.remove('wf-file-upload--dragover');
  };

  private _onDrop = (e: DragEvent): void => {
    e.preventDefault();
    this.classList.remove('wf-file-upload--dragover');
    // Lion's LionInputFile should handle the file processing
  };

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this.classList.toggle('wf-field--error', this.showsFeedbackFor?.includes('error') ?? false);
  }
}

customElements.define('wf-file-upload', WfFileUpload);

declare global {
  interface HTMLElementTagNameMap { 'wf-file-upload': WfFileUpload; }
}
```

- [ ] **Step 3: Write tests**

Tests: file selection via input, drag-and-drop visual state, multiple files, accept filter, file list display, remove file, validation (required, max file size).

- [ ] **Step 4: Register in index.ts and components.css**

---

## Chunk 5: Form Container and Validation

### Task 16: Create wf-form

**Files:**
- Create: `web/packages/ui/src/form/wf-form.ts`
- Create: `web/packages/ui/src/styles/form.css`
- Create: `web/packages/ui/tests/form/wf-form.test.ts`
- Modify: `web/packages/ui/src/styles/components.css`
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Write form.css**

```css
/* form.css */
.wf-form {
  display: block;
}

.wf-form__fields {
  display: flex;
  flex-direction: column;
  gap: var(--wf-space-md);
}

.wf-form__actions {
  display: flex;
  gap: var(--wf-space-sm);
  margin-top: var(--wf-space-lg);
}
```

- [ ] **Step 2: Write wf-form.ts**

Extend `LionForm` from `@lion/ui/form.js`. Lion's `LionForm` uses `FormGroupMixin` to:
- Aggregate field values into a single `modelValue` object
- Track group-level `dirty`, `touched`, `submitted` states
- Validate all child fields on submit
- Prevent native form submission

```typescript
// src/form/wf-form.ts
import { LionForm } from '@lion/ui/form.js';
import { html, type TemplateResult } from 'lit';

export class WfForm extends LionForm {
  createRenderRoot(): this { return this; }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-form');
  }

  render(): TemplateResult {
    return html`
      <form @submit=${this._onSubmit}>
        <div class="wf-form__fields">
          <slot></slot>
        </div>
      </form>
    `;
  }

  private _onSubmit = (e: Event): void => {
    e.preventDefault();
    this.submitted = true;
    if (this.hasFeedbackFor?.includes('error')) {
      // Don't dispatch submit event if there are errors
      return;
    }
    this.dispatchEvent(new CustomEvent('wf-submit', {
      bubbles: true,
      composed: true,
      detail: { value: this.modelValue },
    }));
  };
}

customElements.define('wf-form', WfForm);

declare global {
  interface HTMLElementTagNameMap { 'wf-form': WfForm; }
}
```

- [ ] **Step 3: Write tests**

Tests: aggregated modelValue from child fields, submit triggers validation on all children, wf-submit fires only when valid, submitted state propagates to children (so they all show errors), reset clears all fields and states.

- [ ] **Step 4: Register in index.ts and components.css**

---

### Task 17: Export validators and re-export Lion utilities

**Files:**
- Create: `web/packages/ui/src/form/validators.ts`
- Modify: `web/packages/ui/src/index.ts`

Re-export Lion's validators under the `@workfort/ui` namespace so consumers don't need to import from `@lion/ui` directly:

```typescript
// src/form/validators.ts
export {
  Required,
  MinLength,
  MaxLength,
  MinMaxLength,
  IsEmail,
  Pattern,
  IsNumber,
  MinNumber,
  MaxNumber,
  MinMaxNumber,
  IsDate,
  MinDate,
  MaxDate,
  MinMaxDate,
  Validator,
} from '@lion/ui/form-core.js';
```

Add to `index.ts`:

```typescript
export {
  Required, MinLength, MaxLength, MinMaxLength,
  IsEmail, Pattern,
  IsNumber, MinNumber, MaxNumber, MinMaxNumber,
  IsDate, MinDate, MaxDate, MinMaxDate,
  Validator,
} from './form/validators.js';
```

---

## Chunk 6: Framework Adapters

### Task 18: Update React wrappers

**Files:**
- Modify: `web/packages/react/src/components.tsx`

- [ ] **Step 1: Add type imports for new form components**

```typescript
import type {
  WfPanel, WfButton, WfBadge, WfStatusDot, WfSkeleton,
  WfTextInput, WfList, WfListItem, WfScrollArea, WfErrorFallback,
  WfInput, WfTextarea, WfSelect, WfCheckbox, WfCheckboxGroup,
  WfRadio, WfRadioGroup, WfToggle, WfSlider, WfDatePicker,
  WfFileUpload, WfForm,
} from '@workfort/ui';
```

- [ ] **Step 2: Add form component wrappers**

```typescript
// Form components
export const Input = wrapWc<WfInput, {
  label?: string;
  value?: string;
  'help-text'?: string;
  placeholder?: string;
  type?: string;
  disabled?: boolean;
  readonly?: boolean;
}>('wf-input', 'Input');

export const Textarea = wrapWc<WfTextarea, {
  label?: string;
  value?: string;
  'help-text'?: string;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  readonly?: boolean;
}>('wf-textarea', 'Textarea');

export const Select = wrapWc<WfSelect, {
  label?: string;
  'help-text'?: string;
  disabled?: boolean;
}>('wf-select', 'Select');

export const Checkbox = wrapWc<WfCheckbox, {
  label?: string;
  checked?: boolean;
  disabled?: boolean;
  indeterminate?: boolean;
}>('wf-checkbox', 'Checkbox');

export const CheckboxGroup = wrapWc<WfCheckboxGroup, {
  label?: string;
}>('wf-checkbox-group', 'CheckboxGroup');

export const Radio = wrapWc<WfRadio, {
  label?: string;
  checked?: boolean;
  disabled?: boolean;
}>('wf-radio', 'Radio');

export const RadioGroup = wrapWc<WfRadioGroup, {
  label?: string;
}>('wf-radio-group', 'RadioGroup');

export const Toggle = wrapWc<WfToggle, {
  label?: string;
  checked?: boolean;
  disabled?: boolean;
}>('wf-toggle', 'Toggle');

export const Slider = wrapWc<WfSlider, {
  label?: string;
  min?: number;
  max?: number;
  step?: number;
  'show-value'?: boolean;
  disabled?: boolean;
}>('wf-slider', 'Slider');

export const DatePicker = wrapWc<WfDatePicker, {
  label?: string;
  'help-text'?: string;
  disabled?: boolean;
}>('wf-date-picker', 'DatePicker');

export const FileUpload = wrapWc<WfFileUpload, {
  label?: string;
  'help-text'?: string;
  accept?: string;
  multiple?: boolean;
  disabled?: boolean;
}>('wf-file-upload', 'FileUpload');

export const Form = wrapWc<WfForm, {}>('wf-form', 'Form');
```

- [ ] **Step 3: Add React useFormValidation hook**

**Files:**
- Create: `web/packages/react/src/use-form-validation.ts`
- Modify: `web/packages/react/src/index.tsx` (add export)

```typescript
// src/use-form-validation.ts
import { useRef, useState, useEffect, useCallback } from 'react';

interface FormValidationState {
  dirty: boolean;
  touched: boolean;
  submitted: boolean;
  hasFeedbackFor: string[];
  modelValue: unknown;
}

/**
 * React hook for interacting with Lion-backed form elements.
 * Syncs Lion's interaction states to React state for conditional rendering.
 */
export function useFormValidation<E extends HTMLElement>(
  ref: React.RefObject<E | null>,
): FormValidationState & { submit: () => void; reset: () => void } {
  const [state, setState] = useState<FormValidationState>({
    dirty: false,
    touched: false,
    submitted: false,
    hasFeedbackFor: [],
    modelValue: undefined,
  });

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    const handler = () => {
      setState({
        dirty: (el as any).dirty ?? false,
        touched: (el as any).touched ?? false,
        submitted: (el as any).submitted ?? false,
        hasFeedbackFor: (el as any).hasFeedbackFor ?? [],
        modelValue: (el as any).modelValue,
      });
    };

    el.addEventListener('model-value-changed', handler);
    el.addEventListener('focusout', handler);
    return () => {
      el.removeEventListener('model-value-changed', handler);
      el.removeEventListener('focusout', handler);
    };
  }, [ref]);

  const submit = useCallback(() => {
    const el = ref.current;
    if (el && 'submitted' in el) {
      (el as any).submitted = true;
    }
  }, [ref]);

  const reset = useCallback(() => {
    const el = ref.current;
    if (el && 'resetGroup' in el) {
      (el as any).resetGroup();
    }
  }, [ref]);

  return { ...state, submit, reset };
}
```

---

### Task 19: Add Vue form validation composable

**Files:**
- Create: `web/packages/vue/src/use-form-validation.ts`

```typescript
// src/use-form-validation.ts
import { ref, onMounted, onUnmounted, type Ref } from 'vue';

export function useFormValidation(elRef: Ref<HTMLElement | null>) {
  const dirty = ref(false);
  const touched = ref(false);
  const submitted = ref(false);
  const hasFeedbackFor = ref<string[]>([]);
  const modelValue = ref<unknown>(undefined);

  let cleanup: (() => void) | null = null;

  onMounted(() => {
    const el = elRef.value;
    if (!el) return;

    const handler = () => {
      dirty.value = (el as any).dirty ?? false;
      touched.value = (el as any).touched ?? false;
      submitted.value = (el as any).submitted ?? false;
      hasFeedbackFor.value = (el as any).hasFeedbackFor ?? [];
      modelValue.value = (el as any).modelValue;
    };

    el.addEventListener('model-value-changed', handler);
    el.addEventListener('focusout', handler);
    cleanup = () => {
      el.removeEventListener('model-value-changed', handler);
      el.removeEventListener('focusout', handler);
    };
  });

  onUnmounted(() => cleanup?.());

  return { dirty, touched, submitted, hasFeedbackFor, modelValue };
}
```

---

### Task 20: Add Svelte form validation store

**Files:**
- Create: `web/packages/svelte/src/use-form-validation.ts`

```typescript
// src/use-form-validation.ts
import { writable, type Readable } from 'svelte/store';

interface FormValidationState {
  dirty: boolean;
  touched: boolean;
  submitted: boolean;
  hasFeedbackFor: string[];
  modelValue: unknown;
}

export function createFormValidation(node: HTMLElement): Readable<FormValidationState> {
  const { subscribe, set } = writable<FormValidationState>({
    dirty: false,
    touched: false,
    submitted: false,
    hasFeedbackFor: [],
    modelValue: undefined,
  });

  const handler = () => {
    set({
      dirty: (node as any).dirty ?? false,
      touched: (node as any).touched ?? false,
      submitted: (node as any).submitted ?? false,
      hasFeedbackFor: (node as any).hasFeedbackFor ?? [],
      modelValue: (node as any).modelValue,
    });
  };

  node.addEventListener('model-value-changed', handler);
  node.addEventListener('focusout', handler);

  return {
    subscribe,
    // Caller should call this in onDestroy
    destroy() {
      node.removeEventListener('model-value-changed', handler);
      node.removeEventListener('focusout', handler);
    },
  } as Readable<FormValidationState> & { destroy: () => void };
}
```

---

### Task 21: Add Solid form validation primitive

**Files:**
- Create: `web/packages/solid/src/use-form-validation.ts`

```typescript
// src/use-form-validation.ts
import { createSignal, onCleanup, type Accessor } from 'solid-js';

interface FormValidationState {
  dirty: boolean;
  touched: boolean;
  submitted: boolean;
  hasFeedbackFor: string[];
  modelValue: unknown;
}

export function useFormValidation(el: Accessor<HTMLElement | null>): Accessor<FormValidationState> {
  const [state, setState] = createSignal<FormValidationState>({
    dirty: false,
    touched: false,
    submitted: false,
    hasFeedbackFor: [],
    modelValue: undefined,
  });

  const handler = () => {
    const node = el();
    if (!node) return;
    setState({
      dirty: (node as any).dirty ?? false,
      touched: (node as any).touched ?? false,
      submitted: (node as any).submitted ?? false,
      hasFeedbackFor: (node as any).hasFeedbackFor ?? [],
      modelValue: (node as any).modelValue,
    });
  };

  // Setup when element is available
  const node = el();
  if (node) {
    node.addEventListener('model-value-changed', handler);
    node.addEventListener('focusout', handler);
    onCleanup(() => {
      node.removeEventListener('model-value-changed', handler);
      node.removeEventListener('focusout', handler);
    });
  }

  return state;
}
```

---

## Chunk 7: Migration and Cleanup

### Task 22: Write migration guide

**Files:**
- Modify: `web/packages/ui/src/components/text-input.ts` (add console.warn on use)

- [ ] **Step 1: Add deprecation warning to wf-text-input**

```typescript
connectedCallback(): void {
  super.connectedCallback();
  console.warn(
    '[WorkFort UI] <wf-text-input> is deprecated. Use <wf-input> instead. ' +
    'See migration guide: wf-text-input -> wf-input'
  );
  this.classList.add('wf-text-input');
  this._ensureInput();
  this._sync();
}
```

- [ ] **Step 2: Document migration mapping**

Add a comment block at the top of `wf-input.ts` documenting the migration:

```typescript
/**
 * Migration from wf-text-input to wf-input:
 *
 * | wf-text-input        | wf-input                          |
 * |----------------------|-----------------------------------|
 * | value="..."          | .modelValue = "..."               |
 * | @wf-input            | @model-value-changed              |
 * | @wf-change           | @wf-change                        |
 * | disabled             | disabled                          |
 * | placeholder          | placeholder                       |
 * | (no label)           | label="..."                       |
 * | (no validation)      | .validators=[new Required()]      |
 * | (no dirty/touched)   | .dirty / .touched / .submitted    |
 */
```

---

### Task 23: Update package.json exports and run full test suite

**Files:**
- Modify: `web/packages/ui/package.json` (verify exports, add form subpath if needed)

- [ ] **Step 1: Verify package exports**

The current export map exposes `"."` which imports `index.ts`. Since we added form component imports to `index.ts`, they're included automatically. No additional export entries needed unless we want tree-shakeable subpath exports:

```json
"exports": {
  ".": { ... },
  "./form": {
    "types": "./dist/form/index.d.ts",
    "development": "./src/form/index.ts",
    "default": "./dist/form/index.js"
  },
  "./validators": {
    "types": "./dist/form/validators.d.ts",
    "development": "./src/form/validators.ts",
    "default": "./dist/form/validators.js"
  },
  "./style.css": { ... }
}
```

Consider adding these subpath exports for consumers who want only form components without pulling in the entire library.

- [ ] **Step 2: Run full test suite**

```bash
cd web/packages/ui && npm test
```

All existing tests must still pass. New form tests must pass.

- [ ] **Step 3: Run build**

```bash
cd web/packages/ui && npm run build
```

Verify the build output includes form components and that bundle size is reasonable. Check that Lion is tree-shakeable — consumers importing only `WfButton` should not get Lion's form system.

- [ ] **Step 4: Verify tree-shaking**

Create a minimal consumer test:

```bash
# In a temp directory, create a minimal app that imports only WfButton
# Build it and check bundle size — should NOT include @lion/ui
```

---

## Summary

| Chunk | Tasks | Components | Key Risk |
|-------|-------|------------|----------|
| 0 — Spike | 1–2 | (spike only) | Lion + light DOM compatibility |
| 1 — Foundation | 3–6 | WfInput (replaces WfTextInput) | Render override pattern |
| 2 — Text-Like | 7–8 | WfTextarea, WfSelect | Overlay system in light DOM |
| 3 — Choice | 9–11 | WfCheckbox/Group, WfRadio/Group, WfToggle | ChoiceInputMixin in light DOM |
| 4 — Advanced | 12–15 | WfSlider, WfCombobox, WfDatePicker, WfFileUpload | Complex overlay components |
| 5 — Form Container | 16–17 | WfForm, validator re-exports | FormGroupMixin aggregation |
| 6 — Adapters | 18–21 | React wrappers, Vue/Svelte/Solid composables | Event forwarding |
| 7 — Migration | 22–23 | (cleanup) | Backward compatibility |

**Total new files:** ~30 (components, CSS, tests, adapters)
**Total modified files:** ~5 (index.ts, components.css, package.json, react components.tsx, text-input.ts)

**Critical path:** Chunk 0 (spike) must pass before any other work begins. If the spike fails, the entire plan is blocked pending architectural review.
