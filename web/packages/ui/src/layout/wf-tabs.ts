import { WfElement } from '../base.js';
import type { WfTabPanel } from './wf-tab-panel.js';

/**
 * `<wf-tabs>` -- Tab bar with panels. Add `<wf-tab-panel>` children.
 *
 * @element wf-tabs
 * @fires wf-tab-change -- When the active tab changes.
 */
export class WfTabs extends WfElement {
  static get properties() {
    return {
      activeTab: { type: String, attribute: 'active-tab', reflect: true },
    };
  }

  activeTab = '';

  private _tabList: HTMLDivElement | null = null;
  private _observer: MutationObserver | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-tabs');

    // Observe child changes to rebuild tabs
    this._observer = new MutationObserver(() => this._buildTabList());
    this._observer.observe(this, { childList: true });
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._observer?.disconnect();
  }

  // Skip Lit's template rendering -- DOM is managed imperatively.
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  protected override update(_changedProperties: Map<string, unknown>): void {
    // Intentionally empty to prevent Lit from rendering into this
    // element's DOM, which conflicts with externally-set innerHTML.
  }

  selectTab(name: string): void {
    // Ensure tab list is built if it hasn't been yet
    if (!this._tabList) {
      this._buildTabList();
    }
    this.activeTab = name;
    this._syncPanels();
    this._updateTabListActive();
    this.dispatchEvent(
      new CustomEvent('wf-tab-change', {
        detail: { name },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _getPanels(): WfTabPanel[] {
    return Array.from(this.querySelectorAll('wf-tab-panel')) as WfTabPanel[];
  }

  private _syncPanels(): void {
    this._getPanels().forEach((panel) => {
      panel.active = panel.name === this.activeTab;
    });
  }

  private _buildTabList(): void {
    const panels = this._getPanels();
    if (panels.length === 0) return;

    // Default to first panel if no active tab set
    if (!this.activeTab && panels.length > 0) {
      this.activeTab = panels[0].name;
    }

    // Remove old tab list if it exists
    if (this._tabList) {
      this._tabList.remove();
    }

    // Create tab list
    const tabList = document.createElement('div');
    tabList.classList.add('wf-tabs__list');
    tabList.setAttribute('role', 'tablist');
    tabList.addEventListener('keydown', (e) => this._handleKeydown(e));

    panels.forEach((panel) => {
      const button = document.createElement('button');
      button.classList.add('wf-tabs__tab');
      if (panel.name === this.activeTab) {
        button.classList.add('wf-tabs__tab--active');
      }
      button.setAttribute('role', 'tab');
      button.setAttribute('aria-selected', String(panel.name === this.activeTab));
      button.setAttribute('tabindex', panel.name === this.activeTab ? '0' : '-1');
      button.textContent = panel.label;
      button.addEventListener('click', () => this.selectTab(panel.name));
      tabList.appendChild(button);
    });

    this._tabList = tabList;
    this.insertBefore(tabList, this.firstChild);

    this._syncPanels();
  }

  private _updateTabListActive(): void {
    if (!this._tabList) return;
    const buttons = this._tabList.querySelectorAll('.wf-tabs__tab');
    const panels = this._getPanels();
    buttons.forEach((btn, i) => {
      const isActive = panels[i]?.name === this.activeTab;
      btn.classList.toggle('wf-tabs__tab--active', isActive);
      btn.setAttribute('aria-selected', String(isActive));
      btn.setAttribute('tabindex', isActive ? '0' : '-1');
    });
  }

  private _handleKeydown(e: KeyboardEvent): void {
    const panels = this._getPanels();
    const names = panels.map((p) => p.name);
    const currentIndex = names.indexOf(this.activeTab);
    let newIndex = currentIndex;

    switch (e.key) {
      case 'ArrowRight':
        e.preventDefault();
        newIndex = (currentIndex + 1) % names.length;
        break;
      case 'ArrowLeft':
        e.preventDefault();
        newIndex = (currentIndex - 1 + names.length) % names.length;
        break;
      case 'Home':
        e.preventDefault();
        newIndex = 0;
        break;
      case 'End':
        e.preventDefault();
        newIndex = names.length - 1;
        break;
      default:
        return;
    }

    this.selectTab(names[newIndex]);
    const buttons = this._tabList?.querySelectorAll('.wf-tabs__tab');
    (buttons?.[newIndex] as HTMLElement)?.focus();
  }
}

customElements.define('wf-tabs', WfTabs);

declare global {
  interface HTMLElementTagNameMap {
    'wf-tabs': WfTabs;
  }
}
