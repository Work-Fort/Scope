// src/form/wf-date-picker.ts
// Date picker wrapping a native <input type="date"> with WorkFort styling.
import { LitElement, html } from 'lit';

/**
 * `<wf-date-picker>` — Date picker using a native date input.
 * Provides browser-native date picking with WorkFort design tokens.
 *
 * @element wf-date-picker
 * @fires wf-change — When the selected date changes.
 */
export class WfDatePicker extends LitElement {
  static get properties() {
    return {
      label: { type: String, reflect: true },
      value: { type: String, reflect: true },
      min: { type: String, reflect: true },
      max: { type: String, reflect: true },
      disabled: { type: Boolean, reflect: true },
    };
  }

  label = '';
  value = '';
  min = '';
  max = '';
  disabled = false;

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-date-picker', 'wf-field');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
    // Ensure native input reflects disabled state imperatively
    const input = this.querySelector(
      'input[type="date"]',
    ) as HTMLInputElement | null;
    if (input) {
      input.disabled = this.disabled;
    }
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-field--disabled', this.disabled);
  }

  private _handleChange(e: Event): void {
    e.stopPropagation();
    const input = e.target as HTMLInputElement;
    this.value = input.value;
    this.dispatchEvent(
      new CustomEvent('wf-change', {
        detail: { value: this.value },
        bubbles: true,
        composed: true,
      }),
    );
  }

  render() {
    return html`
      <label class="wf-field__label" ?hidden=${!this.label}>${this.label}</label>
      <div class="wf-field__container">
        <input
          type="date"
          class="wf-field__input"
          value=${this.value}
          min=${this.min}
          max=${this.max}
          ?disabled=${this.disabled}
          @change=${this._handleChange}
        />
      </div>
    `;
  }
}

customElements.define('wf-date-picker', WfDatePicker);

declare global {
  interface HTMLElementTagNameMap {
    'wf-date-picker': WfDatePicker;
  }
}
