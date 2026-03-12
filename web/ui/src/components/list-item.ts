import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfListItem extends WfElement {
  @property({ type: Boolean, reflect: true }) active = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-list-item');
    this.setAttribute('role', 'listitem');
    this.addEventListener('click', this._handleClick);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('click', this._handleClick);
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('active')) {
      this.classList.toggle('wf-list-item--active', this.active);
    }
    this._wrapTrailingContent();
  }

  /** Finds [data-wf="trailing"] children and wraps in .wf-list-item__trailing. */
  private _wrapTrailingContent(): void {
    const trailing = this.querySelectorAll('[data-wf="trailing"]');
    if (trailing.length === 0) return;
    let container = this.querySelector('.wf-list-item__trailing');
    if (!container) {
      container = document.createElement('span');
      container.className = 'wf-list-item__trailing';
      this.appendChild(container);
    }
    trailing.forEach((el) => {
      if (el.parentElement !== container) container!.appendChild(el);
    });
  }

  private _handleClick = (): void => {
    this.dispatchEvent(
      new CustomEvent('wf-select', { bubbles: true, composed: true }),
    );
  };
}

customElements.define('wf-list-item', WfListItem);

declare global {
  interface HTMLElementTagNameMap {
    'wf-list-item': WfListItem;
  }
}
