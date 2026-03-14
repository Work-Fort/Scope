// src/form/wf-form.ts
import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-form>` — Form container that wraps child form fields,
 * prevents native submission, and dispatches `wf-submit` with form data.
 *
 * Uses a simpler approach than LionForm to avoid heavy FormGroupMixin
 * dependencies while still providing a clean submission lifecycle.
 *
 * @element wf-form
 * @fires wf-submit - Dispatched on valid submission with `{ detail: { data } }`
 */
export class WfForm extends WfElement {
  private _submitted = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-form');
    this.addEventListener('submit', this._onSubmit as EventListener);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('submit', this._onSubmit as EventListener);
  }

  render() {
    return html`<form @submit=${this._onSubmit}><slot></slot></form>`;
  }

  /** Whether the form has been submitted at least once. */
  get submitted(): boolean {
    return this._submitted;
  }

  /** The native `<form>` element rendered inside. */
  get formNode(): HTMLFormElement | null {
    return this.querySelector('form');
  }

  /**
   * Gather form data from the native `<form>` element.
   * Returns a plain object keyed by field name.
   */
  get serializedValue(): Record<string, FormDataEntryValue> {
    const form = this.formNode;
    if (!form) return {};
    const fd = new FormData(form);
    const data: Record<string, FormDataEntryValue> = {};
    fd.forEach((value, key) => {
      data[key] = value;
    });
    return data;
  }

  /** Programmatically trigger submission. */
  submit(): void {
    const form = this.formNode;
    if (form) {
      form.requestSubmit();
    }
  }

  /** Reset the form and clear submitted state. */
  reset(): void {
    const form = this.formNode;
    if (form) {
      form.reset();
    }
    this._submitted = false;
    this.classList.remove('wf-form--submitted');
  }

  private _onSubmit = (ev: Event): void => {
    ev.preventDefault();
    ev.stopPropagation();

    this._submitted = true;
    this.classList.add('wf-form--submitted');

    this.dispatchEvent(
      new CustomEvent('wf-submit', {
        bubbles: true,
        composed: true,
        detail: { data: this.serializedValue },
      }),
    );
  };
}

customElements.define('wf-form', WfForm);

declare global {
  interface HTMLElementTagNameMap {
    'wf-form': WfForm;
  }
}
