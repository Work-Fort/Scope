import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfTextInput extends WfElement {
  @property({ type: String }) placeholder = '';
  @property({ type: String }) value = '';
  @property({ type: Boolean, reflect: true }) disabled = false;

  private _input: HTMLInputElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-text-input');
    this._ensureInput();
    this._sync();
  }

  updated(): void {
    this._sync();
  }

  private _ensureInput(): void {
    if (this._input) return;
    const input = document.createElement('input');
    input.className = 'wf-text-input__input';
    input.addEventListener('input', this._onInput);
    input.addEventListener('change', this._onChange);
    this.appendChild(input);
    this._input = input;
  }

  private _sync(): void {
    if (!this._input) return;
    this._input.value = this.value;
    this._input.placeholder = this.placeholder;
    this._input.disabled = this.disabled;
  }

  private _onInput = (e: Event): void => {
    const input = e.target as HTMLInputElement;
    this.value = input.value;
    this.dispatchEvent(new CustomEvent('wf-input', {
      bubbles: true, composed: true, detail: { value: input.value },
    }));
  };

  private _onChange = (e: Event): void => {
    const input = e.target as HTMLInputElement;
    this.dispatchEvent(new CustomEvent('wf-change', {
      bubbles: true, composed: true, detail: { value: input.value },
    }));
  };

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._input) {
      this._input.removeEventListener('input', this._onInput);
      this._input.removeEventListener('change', this._onChange);
    }
  }
}

customElements.define('wf-text-input', WfTextInput);

declare global {
  interface HTMLElementTagNameMap {
    'wf-text-input': WfTextInput;
  }
}
