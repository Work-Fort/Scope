import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfErrorFallback extends WfElement {
  @property({ type: String }) title = '';
  @property({ type: String }) message = '';

  private _titleEl: HTMLDivElement | null = null;
  private _messageEl: HTMLDivElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-error-fallback');
    this.setAttribute('role', 'alert');
    this._ensureChildren();
    this._sync();
  }

  updated(): void {
    this._sync();
  }

  private _ensureChildren(): void {
    if (this._titleEl) return;

    const titleEl = document.createElement('div');
    titleEl.className = 'wf-error-fallback__title';
    this.appendChild(titleEl);
    this._titleEl = titleEl;

    const messageEl = document.createElement('div');
    messageEl.className = 'wf-error-fallback__message';
    this.appendChild(messageEl);
    this._messageEl = messageEl;
  }

  private _sync(): void {
    if (this._titleEl) {
      this._titleEl.textContent = this.title;
      this._titleEl.style.display = this.title ? '' : 'none';
    }
    if (this._messageEl) {
      this._messageEl.textContent = this.message;
      this._messageEl.style.display = this.message ? '' : 'none';
    }
  }
}

customElements.define('wf-error-fallback', WfErrorFallback);

declare global {
  interface HTMLElementTagNameMap {
    'wf-error-fallback': WfErrorFallback;
  }
}
