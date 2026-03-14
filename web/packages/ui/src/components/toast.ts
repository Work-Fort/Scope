import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfToast extends WfElement {
  @property({ type: String }) variant: 'error' | 'warning' | 'info' | 'success' = 'info';
  @property({ type: Boolean }) sticky = false;
  @property({ type: Number }) duration = 5000;

  private _timer: ReturnType<typeof setTimeout> | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-toast');
    this._applyVariant();
    this._ensureCloseButton();

    if (!this.sticky) {
      this._timer = setTimeout(() => this._dismiss(), this.duration);
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._timer) {
      clearTimeout(this._timer);
      this._timer = null;
    }
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('variant')) this._applyVariant();
  }

  private _applyVariant(): void {
    this.classList.remove('wf-toast--error', 'wf-toast--warning', 'wf-toast--info', 'wf-toast--success');
    this.classList.add(`wf-toast--${this.variant}`);
  }

  private _ensureCloseButton(): void {
    if (this.querySelector('.wf-toast__close')) return;
    const btn = document.createElement('button');
    btn.className = 'wf-toast__close';
    btn.setAttribute('aria-label', 'Dismiss');
    btn.textContent = '✕';
    btn.addEventListener('click', () => this._dismiss());
    this.appendChild(btn);
  }

  private _dismiss(): void {
    this.dispatchEvent(new CustomEvent('wf-dismiss', { bubbles: true, composed: true }));
    this.remove();
  }
}

customElements.define('wf-toast', WfToast);

declare global {
  interface HTMLElementTagNameMap {
    'wf-toast': WfToast;
  }
}
