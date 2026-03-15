import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-accordion-item>` -- Collapsible section within a `<wf-accordion>`.
 *
 * @element wf-accordion-item
 * @slot - Default slot for collapsed content.
 * @fires wf-accordion-change -- Bubbles to parent accordion on toggle.
 */
export class WfAccordionItem extends WfElement {
  static get properties() {
    return {
      name: { type: String, reflect: true },
      header: { type: String, reflect: true },
      expanded: { type: Boolean, reflect: true },
    };
  }

  name = '';
  header = '';
  expanded = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-accordion-item');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-accordion-item--expanded', this.expanded);
  }

  toggle(): void {
    this.expanded = !this.expanded;
    this.requestUpdate();
    this.dispatchEvent(
      new CustomEvent('wf-accordion-change', {
        detail: { name: this.name, expanded: this.expanded },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleHeaderClick(): void {
    this.toggle();
  }

  private _handleKeydown(e: KeyboardEvent): void {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      this.toggle();
    }
  }

  render() {
    return html`
      <div
        class="wf-accordion-item__header"
        role="button"
        tabindex="0"
        aria-expanded=${this.expanded}
        @click=${this._handleHeaderClick}
        @keydown=${this._handleKeydown}
      >
        <span class="wf-accordion-item__title">${this.header}</span>
        <span class="wf-accordion-item__icon">${this.expanded ? '\u2212' : '+'}</span>
      </div>
      <div class="wf-accordion-item__body" ?hidden=${!this.expanded}>
        <slot></slot>
      </div>
    `;
  }
}

customElements.define('wf-accordion-item', WfAccordionItem);

declare global {
  interface HTMLElementTagNameMap {
    'wf-accordion-item': WfAccordionItem;
  }
}
