import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfBanner extends WfElement {
  @property({ type: String }) variant: 'error' | 'warning' | 'info' = 'info';
  @property({ type: Boolean }) dismissible = false;
  @property({ type: String }) headline = '';
  @property({ type: String }) details = '';

  private _expanded = false;
  private _headlineEl: HTMLSpanElement | null = null;
  private _detailsEl: HTMLDivElement | null = null;
  private _toggleBtn: HTMLButtonElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-banner');
    this._buildDOM();
    this._applyVariant();
    this._sync();
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('variant')) this._applyVariant();
    if (changed.has('headline') || changed.has('details') || changed.has('dismissible')) {
      this._sync();
    }
  }

  private _buildDOM(): void {
    const content = document.createElement('div');
    content.className = 'wf-banner__content';

    const icon = document.createElement('span');
    icon.className = 'wf-banner__icon';
    icon.setAttribute('aria-hidden', 'true');
    icon.textContent = '●';
    content.appendChild(icon);

    this._headlineEl = document.createElement('span');
    this._headlineEl.className = 'wf-banner__headline';
    content.appendChild(this._headlineEl);

    const actions = document.createElement('span');
    actions.className = 'wf-banner__actions';

    this._toggleBtn = document.createElement('button');
    this._toggleBtn.className = 'wf-banner__toggle';
    this._toggleBtn.setAttribute('aria-label', 'Toggle details');
    this._toggleBtn.textContent = '▾';
    this._toggleBtn.addEventListener('click', this._toggle);
    actions.appendChild(this._toggleBtn);

    if (this.dismissible) {
      const closeBtn = document.createElement('button');
      closeBtn.className = 'wf-banner__close';
      closeBtn.setAttribute('aria-label', 'Dismiss');
      closeBtn.textContent = '✕';
      closeBtn.addEventListener('click', this._dismiss);
      actions.appendChild(closeBtn);
    }

    content.appendChild(actions);
    this.appendChild(content);

    this._detailsEl = document.createElement('div');
    this._detailsEl.className = 'wf-banner__details';
    this._detailsEl.style.display = 'none';
    this.appendChild(this._detailsEl);
  }

  private _sync(): void {
    if (this._headlineEl) this._headlineEl.textContent = this.headline;
    if (this._detailsEl) this._detailsEl.textContent = this.details;
    if (this._toggleBtn) {
      this._toggleBtn.style.display = this.details ? '' : 'none';
    }
  }

  private _applyVariant(): void {
    this.classList.remove('wf-banner--error', 'wf-banner--warning', 'wf-banner--info');
    this.classList.add(`wf-banner--${this.variant}`);
  }

  private _toggle = (): void => {
    this._expanded = !this._expanded;
    if (this._detailsEl) {
      this._detailsEl.style.display = this._expanded ? '' : 'none';
    }
    if (this._toggleBtn) {
      this._toggleBtn.textContent = this._expanded ? '▴' : '▾';
    }
  };

  private _dismiss = (): void => {
    this.style.display = 'none';
    this.dispatchEvent(new CustomEvent('wf-dismiss', { bubbles: true, composed: true }));
  };

  show(): void {
    this.style.display = '';
  }
}

customElements.define('wf-banner', WfBanner);

declare global {
  interface HTMLElementTagNameMap {
    'wf-banner': WfBanner;
  }
}
