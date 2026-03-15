import { WfElement } from '../base.js';
import type { WfAccordionItem } from './wf-accordion-item.js';

/**
 * `<wf-accordion>` -- Container for collapsible sections.
 * Add `<wf-accordion-item>` children. Set `multiple` to allow
 * more than one section open at a time.
 *
 * @element wf-accordion
 * @slot - Default slot for wf-accordion-item children.
 */
export class WfAccordion extends WfElement {
  static get properties() {
    return {
      multiple: { type: Boolean, reflect: true },
    };
  }

  multiple = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-accordion');
    this.addEventListener('wf-accordion-change', this._handleItemChange as EventListener);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('wf-accordion-change', this._handleItemChange as EventListener);
  }

  // Skip Lit's template rendering — children are managed externally.
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  protected override update(_changedProperties: Map<string, unknown>): void {
    // Intentionally empty to prevent Lit from rendering into this
    // element's DOM, which conflicts with externally-set innerHTML.
  }

  private _handleItemChange = (e: CustomEvent): void => {
    if (!this.multiple && e.detail.expanded) {
      const items = this.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
      items.forEach((item) => {
        if (item.name !== e.detail.name && item.expanded) {
          item.expanded = false;
          item.requestUpdate();
        }
      });
    }
  };
}

customElements.define('wf-accordion', WfAccordion);

declare global {
  interface HTMLElementTagNameMap {
    'wf-accordion': WfAccordion;
  }
}
