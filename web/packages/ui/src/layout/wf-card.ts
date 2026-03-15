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

  private _userContent: Node[] = [];
  private _didSetup = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-card');
    this._syncClasses();

    // Capture user-provided children before Lit renders
    if (!this._didSetup) {
      this._userContent = Array.from(this.childNodes);
    }
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();

    // After first render, move captured children into the body div
    if (!this._didSetup) {
      this._didSetup = true;
      const body = this.querySelector('.wf-card__body');
      if (body) {
        for (const node of this._userContent) {
          body.appendChild(node);
        }
      }
    }
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-card--outlined', this.variant === 'outlined');
    this.classList.toggle('wf-card--elevated', this.variant === 'elevated');
    this.classList.toggle('wf-card--padded', this.padded);
  }

  render() {
    return html`
      <div class="wf-card__header" ?hidden=${!this.header}>${this.header}</div>
      <div class="wf-card__body"></div>
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
