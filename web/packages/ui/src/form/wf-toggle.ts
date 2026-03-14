// src/form/wf-toggle.ts
// Visual toggle/switch built with a native hidden checkbox and CSS transitions.
// Does not use Lion — purely custom with LitElement base.
import { LitElement, html } from 'lit';

/**
 * `<wf-toggle>` — Toggle switch component with a hidden checkbox input.
 * Fires `wf-change` event when toggled.
 *
 * @element wf-toggle
 * @fires wf-change — When the checked state changes via user interaction.
 */
export class WfToggle extends LitElement {
  static get properties() {
    return {
      checked: { type: Boolean, reflect: true },
      disabled: { type: Boolean, reflect: true },
    };
  }

  checked = false;
  disabled = false;

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-toggle');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-toggle--checked', this.checked);
    this.classList.toggle('wf-toggle--disabled', this.disabled);
  }

  private _handleClick(): void {
    if (this.disabled) return;
    this.checked = !this.checked;
    this._syncClasses();
    this.dispatchEvent(
      new CustomEvent('wf-change', {
        detail: { checked: this.checked },
        bubbles: true,
        composed: true,
      }),
    );
  }

  render() {
    return html`
      <input
        type="checkbox"
        .checked=${this.checked}
        ?disabled=${this.disabled}
        style="position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0;"
        @change=${(e: Event) => e.stopPropagation()}
        aria-hidden="true"
        tabindex="-1"
      />
      <div class="wf-toggle__track" @click=${this._handleClick}>
        <div class="wf-toggle__thumb"></div>
      </div>
      <slot @click=${this._handleClick}></slot>
    `;
  }
}

customElements.define('wf-toggle', WfToggle);

declare global {
  interface HTMLElementTagNameMap {
    'wf-toggle': WfToggle;
  }
}
