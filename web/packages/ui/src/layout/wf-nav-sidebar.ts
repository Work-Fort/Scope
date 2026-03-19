import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import './wf-nav-section.js';

/**
 * `<wf-nav-sidebar>` -- Structural shell for a service navigation sidebar.
 *
 * @element wf-nav-sidebar
 * @slot actions - Buttons in the header next to the title.
 * @slot - Default slot for wf-nav-section children.
 * @fires wf-search - When the search input changes (detail: { term: string }).
 */
export class WfNavSidebar extends WfElement {
  @property({ type: String }) heading = '';
  @property({ attribute: 'search-placeholder', type: String }) searchPlaceholder = 'Search\u2026';

  private _userContent: Node[] = [];
  private _didSetup = false;
  private _childObserver: MutationObserver | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-nav-sidebar');

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
    const actionsSlot = this.querySelector('.wf-nav-sidebar__actions');
    const bodySlot = this.querySelector('.wf-nav-sidebar__body');

    for (const node of this._userContent) {
      if (node instanceof Element) {
        const slot = node.getAttribute('slot');
        if (slot === 'actions' && actionsSlot) {
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
    const actionsSlot = this.querySelector('.wf-nav-sidebar__actions');
    const bodySlot = this.querySelector('.wf-nav-sidebar__body');

    this._childObserver = new MutationObserver((mutations) => {
      for (const m of mutations) {
        for (const node of m.addedNodes) {
          if (node instanceof Element && (
            node.classList.contains('wf-nav-sidebar__header') ||
            node.classList.contains('wf-nav-sidebar__search') ||
            node.classList.contains('wf-nav-sidebar__body')
          )) continue;

          if (node instanceof Element) {
            const slot = node.getAttribute('slot');
            if (slot === 'actions' && actionsSlot) {
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

  private _onSearch(e: InputEvent): void {
    const input = e.target as HTMLInputElement;
    this.dispatchEvent(
      new CustomEvent('wf-search', {
        bubbles: true,
        composed: true,
        detail: { term: input.value },
      }),
    );
  }

  render() {
    return html`
      <div class="wf-nav-sidebar__header">
        <span class="wf-nav-sidebar__title">${this.heading}</span>
        <span class="wf-nav-sidebar__actions"></span>
      </div>
      <div class="wf-nav-sidebar__search">
        <input class="wf-nav-sidebar__search-input"
          type="text"
          placeholder=${this.searchPlaceholder}
          @input=${(e: InputEvent) => this._onSearch(e)} />
      </div>
      <div class="wf-nav-sidebar__body"></div>
    `;
  }
}

customElements.define('wf-nav-sidebar', WfNavSidebar);

declare global {
  interface HTMLElementTagNameMap {
    'wf-nav-sidebar': WfNavSidebar;
  }
}
