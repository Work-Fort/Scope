import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-card>` -- Content container with optional header and footer.
 *
 * @element wf-card
 * @slot - Default slot for card body content.
 */
export class WfCard extends WfElement {
  @property({ type: String, reflect: true }) header = '';
  @property({ type: String, reflect: true }) footer = '';
  @property({ type: String, reflect: true }) variant: 'default' | 'outlined' | 'elevated' = 'default';
  @property({ type: Boolean, reflect: true }) padded = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-card');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-card--outlined', this.variant === 'outlined');
    this.classList.toggle('wf-card--elevated', this.variant === 'elevated');
    this.classList.toggle('wf-card--padded', this.padded);
  }

  render() {
    return html`
      <div class="wf-card__header" ?hidden=${!this.header}>${this.header}</div>
      <div class="wf-card__body"><slot></slot></div>
      <div class="wf-card__footer" ?hidden=${!this.footer}>${this.footer}</div>
    `;
  }
}

customElements.define('wf-card', WfCard);

declare global {
  interface HTMLElementTagNameMap {
    'wf-card': WfCard;
  }
}
