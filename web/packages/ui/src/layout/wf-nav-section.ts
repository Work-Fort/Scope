import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

/**
 * `<wf-nav-section>` -- A collapsible section for use inside wf-nav-sidebar.
 *
 * @element wf-nav-section
 * @slot - Default slot for list items (placed in the body).
 * @slot section-actions - Buttons shown in the header next to the heading.
 */
export class WfNavSection extends WfElement {
  @property({ type: String }) heading = '';
  @property({ type: Boolean, reflect: true }) collapsed = false;

  private _userContent: Node[] = [];
  private _didSetup = false;
  private _childObserver: MutationObserver | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-nav-section');

    if (!this._didSetup) {
      this._userContent = Array.from(this.childNodes);
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._childObserver) {
      this._childObserver.disconnect();
      this._childObserver = null;
    }
  }

  protected override updated(changed: Map<string, unknown>): void {
    super.updated(changed);

    if (!this._didSetup) {
      this._didSetup = true;
      this._distributeSlots();
      this._observeChildren();
    }
  }

  private _distributeSlots(): void {
    const actionsSlot = this.querySelector('.wf-nav-section__actions');
    const bodySlot = this.querySelector('.wf-nav-section__body');

    for (const node of this._userContent) {
      if (node instanceof Element) {
        const slot = node.getAttribute('slot');
        if (slot === 'section-actions' && actionsSlot) {
          actionsSlot.appendChild(node);
        } else if (bodySlot) {
          bodySlot.appendChild(node);
        }
      } else if (node.textContent?.trim() && bodySlot) {
        bodySlot.appendChild(node);
      }
    }
  }

  private _observeChildren(): void {
    const actionsSlot = this.querySelector('.wf-nav-section__actions');
    const bodySlot = this.querySelector('.wf-nav-section__body');

    this._childObserver = new MutationObserver((mutations) => {
      for (const m of mutations) {
        for (const node of m.addedNodes) {
          if (node instanceof Element && (
            node.classList.contains('wf-nav-section__header') ||
            node.classList.contains('wf-nav-section__body')
          )) continue;

          if (node instanceof Element) {
            const slot = node.getAttribute('slot');
            if (slot === 'section-actions' && actionsSlot) {
              actionsSlot.appendChild(node);
            } else if (!slot && bodySlot) {
              bodySlot.appendChild(node);
            }
          }
        }
      }
    });

    this._childObserver.observe(this, { childList: true });
  }

  private _toggle(): void {
    this.collapsed = !this.collapsed;
  }

  render() {
    return html`
      <div class="wf-nav-section__header" @click=${() => this._toggle()}>
        <span class="wf-nav-section__heading">${this.heading}</span>
        <span class="wf-nav-section__actions"></span>
      </div>
      <div class="wf-nav-section__body" ?hidden=${this.collapsed}></div>
    `;
  }
}

customElements.define('wf-nav-section', WfNavSection);

declare global {
  interface HTMLElementTagNameMap {
    'wf-nav-section': WfNavSection;
  }
}
