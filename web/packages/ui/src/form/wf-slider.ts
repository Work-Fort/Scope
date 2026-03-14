// src/form/wf-slider.ts
// Range slider built with a native <input type="range"> and CSS classes.
// Does not use Lion — purely custom with LitElement base.
import { LitElement, html } from 'lit';

/**
 * `<wf-slider>` — Range slider component with a native range input.
 * Fires `wf-input` continuously as the user drags and `wf-change` on release.
 *
 * @element wf-slider
 * @fires wf-input — Continuously as the user drags the slider.
 * @fires wf-change — When the user releases the slider (value committed).
 */
export class WfSlider extends LitElement {
  static get properties() {
    return {
      min: { type: Number, reflect: true },
      max: { type: Number, reflect: true },
      step: { type: Number, reflect: true },
      value: { type: Number, reflect: true },
      disabled: { type: Boolean, reflect: true },
    };
  }

  min = 0;
  max = 100;
  step = 1;
  value = 0;
  disabled = false;

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-slider');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-slider--disabled', this.disabled);
  }

  private _handleInput(e: Event): void {
    const input = e.target as HTMLInputElement;
    this.value = Number(input.value);
    this.dispatchEvent(
      new CustomEvent('wf-input', {
        detail: { value: this.value },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleChange(e: Event): void {
    e.stopPropagation();
    const input = e.target as HTMLInputElement;
    this.value = Number(input.value);
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
      <input
        type="range"
        class="wf-slider__input"
        .min=${String(this.min)}
        .max=${String(this.max)}
        .step=${String(this.step)}
        .value=${String(this.value)}
        ?disabled=${this.disabled}
        @input=${this._handleInput}
        @change=${this._handleChange}
      />
    `;
  }
}

customElements.define('wf-slider', WfSlider);

declare global {
  interface HTMLElementTagNameMap {
    'wf-slider': WfSlider;
  }
}
