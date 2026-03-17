import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';
import type { HamburgerPosition } from './wf-hamburger.js';
import './wf-hamburger.js';

/**
 * `<wf-nav-bar>` -- Responsive navigation bar with overflow detection.
 *
 * Uses named slots: `brand` (left), default (tabs), `actions` (right).
 * Detects tab overflow via ResizeObserver and moves excess into a "more" dropdown.
 * Below the breakpoint, collapses entirely into a wf-hamburger.
 *
 * @element wf-nav-bar
 * @slot brand - Left-aligned brand content.
 * @slot - Default slot for tab items.
 * @slot actions - Right-aligned action buttons.
 */
export class WfNavBar extends WfElement {
  @property({ type: Number, reflect: true }) breakpoint = 640;
  @property({ attribute: 'hamburger-position', type: String, reflect: true })
  hamburgerPosition: HamburgerPosition = 'top-right';

  /** Whether the nav bar is in collapsed (hamburger) mode. */
  @property({ type: Boolean, reflect: true }) collapsed = false;

  /** Whether there are overflowing tabs. */
  @property({ type: Boolean }) hasOverflow = false;

  private _resizeObserver: ResizeObserver | null = null;
  private _userContent: Node[] = [];
  private _didSetup = false;
  private _hamburgerOpen = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-nav-bar');

    if (!this._didSetup) {
      this._userContent = Array.from(this.childNodes);
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this._resizeObserver) {
      this._resizeObserver.disconnect();
      this._resizeObserver = null;
    }
  }

  protected override updated(changed: Map<string, unknown>): void {
    super.updated(changed);

    if (!this._didSetup) {
      this._didSetup = true;
      this._distributeSlots();
      this._setupResizeObserver();
    }

    if (changed.has('collapsed')) {
      if (this.collapsed) {
        this.classList.add('wf-nav-bar--collapsed');
      } else {
        this.classList.remove('wf-nav-bar--collapsed');
      }
    }
  }

  /** Distribute captured children into their respective slot containers. */
  private _distributeSlots(): void {
    const brandSlot = this.querySelector('.wf-nav-bar__brand');
    const tabsSlot = this.querySelector('.wf-nav-bar__tabs');
    const actionsSlot = this.querySelector('.wf-nav-bar__actions');

    for (const node of this._userContent) {
      if (node instanceof Element) {
        const slot = node.getAttribute('slot');
        if (slot === 'brand' && brandSlot) {
          brandSlot.appendChild(node);
        } else if (slot === 'actions' && actionsSlot) {
          actionsSlot.appendChild(node);
        } else if (tabsSlot) {
          tabsSlot.appendChild(node);
        }
      } else if (node.textContent?.trim() && tabsSlot) {
        tabsSlot.appendChild(node);
      }
    }
  }

  /** Set up a ResizeObserver on the tabs area to detect overflow. */
  private _setupResizeObserver(): void {
    if (typeof ResizeObserver === 'undefined') return;

    const tabs = this.querySelector('.wf-nav-bar__tabs') as HTMLElement;
    if (!tabs) return;

    this._resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const target = entry.target as HTMLElement;
        // Check if the nav bar itself is narrower than the breakpoint
        const navWidth = this.getBoundingClientRect().width;
        if (navWidth > 0 && navWidth < this.breakpoint) {
          this.collapsed = true;
        } else if (navWidth >= this.breakpoint) {
          this.collapsed = false;
        }

        // Check for tab overflow
        this.hasOverflow = target.scrollWidth > target.clientWidth;
      }
    });

    this._resizeObserver.observe(tabs);
  }

  private _onHamburgerToggle(e: CustomEvent<{ open: boolean }>): void {
    this._hamburgerOpen = e.detail.open;
    this.requestUpdate();
  }

  render() {
    return html`
      <div class="wf-nav-bar__brand" ?hidden=${this.collapsed}></div>
      <div class="wf-nav-bar__tabs" ?hidden=${this.collapsed}></div>
      <div class="wf-nav-bar__overflow" ?hidden=${!this.hasOverflow || this.collapsed}>
        <button class="wf-nav-bar__overflow-btn" aria-label="More tabs">
          &hellip;
        </button>
      </div>
      <div class="wf-nav-bar__actions" ?hidden=${this.collapsed}></div>
      <wf-hamburger
        position=${this.hamburgerPosition}
        ?open=${this._hamburgerOpen}
        ?hidden=${!this.collapsed}
        @wf-toggle=${(e: CustomEvent) => this._onHamburgerToggle(e)}
      ></wf-hamburger>
    `;
  }
}

customElements.define('wf-nav-bar', WfNavBar);

declare global {
  interface HTMLElementTagNameMap {
    'wf-nav-bar': WfNavBar;
  }
}
