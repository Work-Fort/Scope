import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import '../styles/compose-input.css';

export class WfComposeInput extends WfElement {
  @property({ type: String, reflect: true }) placeholder = '';
  @property({ type: Boolean, reflect: true }) disabled = false;

  private _value = '';
  private _textarea: HTMLTextAreaElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-compose-input');
    this._render();
  }

  private _render(): void {
    this.innerHTML = '';

    const box = document.createElement('div');
    box.className = 'wf-compose-input__box';

    const textarea = document.createElement('textarea');
    textarea.className = 'wf-compose-input__field';
    textarea.placeholder = this.placeholder;
    textarea.rows = 1;
    textarea.disabled = this.disabled;
    this._textarea = textarea;

    textarea.addEventListener('input', () => {
      this._value = textarea.value;
    });

    textarea.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        this._send();
      }
    });

    const button = document.createElement('wf-button');
    button.setAttribute('title', 'Send');
    button.style.cssText = 'padding: 4px 10px;';
    button.textContent = '\u2191';
    if (this.disabled) button.setAttribute('disabled', '');

    button.addEventListener('wf-click', () => this._send());

    box.appendChild(textarea);
    box.appendChild(button);
    this.appendChild(box);
  }

  private _send(): void {
    if (this.disabled) return;
    const body = this._value.trim();
    if (!body) return;

    this.dispatchEvent(new CustomEvent('wf-send', {
      bubbles: true,
      composed: true,
      detail: { body },
    }));

    this._value = '';
    if (this._textarea) this._textarea.value = '';
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('placeholder') && this._textarea) {
      this._textarea.placeholder = this.placeholder;
    }
    if (changed.has('disabled') && this._textarea) {
      this._textarea.disabled = this.disabled;
    }
  }
}

customElements.define('wf-compose-input', WfComposeInput);

declare global {
  interface HTMLElementTagNameMap {
    'wf-compose-input': WfComposeInput;
  }
}
