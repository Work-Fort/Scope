# Phase 3: Layout & Display — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add 8 layout and display components to `@workfort/ui` — Card, Tabs, Accordion, Dialog/Modal, Drawer, Tooltip, Popover, and Table/Data Grid — completing the toolkit for admin panels, settings pages, and data-heavy UIs.

**Architecture:** All 8 components extend `WfElement` (light DOM, no Lion). Phase 2 demonstrated that Lion's overlay primitives add complexity without proportional benefit for light-DOM components. Simple layout components (Card, Tabs, Accordion) are purely CSS-driven. Overlay components (Dialog, Drawer, Tooltip, Popover) use a custom overlay utility that manages backdrop, focus trapping, and `z-index` layering via design tokens. Table/Data Grid is the most complex component, built incrementally with sorting, column resize, pagination, and virtual scrolling.

**Tech Stack:** Lit 3, TypeScript, Vite (library mode), Vitest (happy-dom), CSS custom properties

**Spec:** `docs/ui-component-library-design.md` — Phase 3 section

---

## Component Matrix

| Component | Tag | Extends | CSS Class | Overlay? |
|-----------|-----|---------|-----------|----------|
| Card | `wf-card` | `WfElement` | `.wf-card` | No |
| Tabs | `wf-tabs` | `WfElement` | `.wf-tabs` | No |
| Tab Panel | `wf-tab-panel` | `WfElement` | `.wf-tab-panel` | No |
| Accordion | `wf-accordion` | `WfElement` | `.wf-accordion` | No |
| Accordion Item | `wf-accordion-item` | `WfElement` | `.wf-accordion-item` | No |
| Dialog | `wf-dialog` | `WfElement` | `.wf-dialog` | Yes |
| Drawer | `wf-drawer` | `WfElement` | `.wf-drawer` | Yes |
| Tooltip | `wf-tooltip` | `WfElement` | `.wf-tooltip` | Yes |
| Popover | `wf-popover` | `WfElement` | `.wf-popover` | Yes |
| Table | `wf-table` | `WfElement` | `.wf-table` | No |

---

## Why No Lion for Overlays

Phase 2 validated that Lion-backed components work well for form inputs (where validation lifecycle, dirty/touched state, and ARIA are complex). However, Phase 2 also showed that Lion's overlay-dependent components (combobox, datepicker) were replaced with custom implementations because:

1. Lion's `OverlayMixin` assumes shadow DOM for container placement
2. The overlay controller adds ~8KB of JS for positioning logic we can handle with CSS
3. Custom implementations gave us full control over animation, backdrop, and focus management

For Phase 3, all overlay components (Dialog, Drawer, Tooltip, Popover) use a shared custom overlay utility built from scratch. This utility handles:
- Focus trapping (Tab/Shift+Tab cycling within overlay)
- Backdrop rendering and click-to-close
- Escape key dismissal
- `z-index` management via design tokens (`--wf-z-modal`, `--wf-z-tooltip`, `--wf-z-dropdown`)
- Entry/exit animations via CSS transitions

---

## File Structure

```
web/packages/ui/
├── src/
│   ├── layout/
│   │   ├── wf-card.ts
│   │   ├── wf-tabs.ts
│   │   ├── wf-tab-panel.ts
│   │   ├── wf-accordion.ts
│   │   ├── wf-accordion-item.ts
│   │   ├── wf-dialog.ts
│   │   ├── wf-drawer.ts
│   │   ├── wf-tooltip.ts
│   │   ├── wf-popover.ts
│   │   └── wf-table.ts
│   ├── utils/
│   │   └── overlay.ts          # shared focus trap + backdrop utility
│   └── styles/
│       ├── layout.css          # CSS for all layout/display components
│       └── components.css      # updated to @import './layout.css'
├── tests/
│   └── layout/
│       ├── wf-card.test.ts
│       ├── wf-tabs.test.ts
│       ├── wf-accordion.test.ts
│       ├── wf-dialog.test.ts
│       ├── wf-drawer.test.ts
│       ├── wf-tooltip.test.ts
│       ├── wf-popover.test.ts
│       └── wf-table.test.ts
```

---

## Chunk 1: Simple Components (Card, Tabs, Accordion)

### Task 1: Card component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-card.test.ts`
- Create: `web/packages/ui/src/layout/wf-card.ts`

- [ ] **Step 1: Create test directory**

```bash
mkdir -p web/packages/ui/tests/layout
```

- [ ] **Step 2: Write card tests**

```typescript
// tests/layout/wf-card.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-card.js';
import type { WfCard } from '../../src/layout/wf-card.js';

describe('WfCard', () => {
  afterEach(cleanup);

  it('renders with wf-card class', async () => {
    const el = await fixture<WfCard>('wf-card');
    expect(el.classList.contains('wf-card')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfCard>('wf-card');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders slot content', async () => {
    const el = await fixture<WfCard>('wf-card');
    el.innerHTML = '<p>Card content</p>';
    await el.updateComplete;
    expect(el.querySelector('p')!.textContent).toBe('Card content');
  });

  it('renders header when set', async () => {
    const el = await fixture<WfCard>('wf-card', { header: 'Title' });
    await el.updateComplete;
    const header = el.querySelector('.wf-card__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toBe('Title');
  });

  it('renders footer when set', async () => {
    const el = await fixture<WfCard>('wf-card', { footer: 'Footer text' });
    await el.updateComplete;
    const footer = el.querySelector('.wf-card__footer');
    expect(footer).not.toBeNull();
    expect(footer!.textContent).toBe('Footer text');
  });

  it('applies variant class', async () => {
    const el = await fixture<WfCard>('wf-card', { variant: 'outlined' });
    await el.updateComplete;
    expect(el.classList.contains('wf-card--outlined')).toBe(true);
  });

  it('applies padding class when padded attribute set', async () => {
    const el = await fixture<WfCard>('wf-card', { padded: true });
    await el.updateComplete;
    expect(el.classList.contains('wf-card--padded')).toBe(true);
  });
});
```

- [ ] **Step 3: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-card.test.ts 2>&1 | tail -5
```

Expected: all tests fail (module not found).

- [ ] **Step 4: Create card directory**

```bash
mkdir -p web/packages/ui/src/layout
```

- [ ] **Step 5: Implement WfCard**

```typescript
// src/layout/wf-card.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-card>` — Content container with optional header and footer.
 *
 * @element wf-card
 * @slot - Default slot for card body content.
 */
export class WfCard extends WfElement {
  static get properties() {
    return {
      header: { type: String, reflect: true },
      footer: { type: String, reflect: true },
      variant: { type: String, reflect: true },
      padded: { type: Boolean, reflect: true },
    };
  }

  header = '';
  footer = '';
  variant: 'default' | 'outlined' | 'elevated' = 'default';
  padded = false;

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
      ${this.header
        ? html`<div class="wf-card__header">${this.header}</div>`
        : nothing}
      <div class="wf-card__body"><slot></slot></div>
      ${this.footer
        ? html`<div class="wf-card__footer">${this.footer}</div>`
        : nothing}
    `;
  }
}

customElements.define('wf-card', WfCard);

declare global {
  interface HTMLElementTagNameMap {
    'wf-card': WfCard;
  }
}
```

- [ ] **Step 6: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-card.test.ts 2>&1 | tail -5
```

Expected: 7 tests pass.

- [ ] **Step 7: Commit**

```bash
git add web/packages/ui/src/layout/wf-card.ts web/packages/ui/tests/layout/wf-card.test.ts
git commit -m "feat(ui): add wf-card component with header, footer, and variant support"
```

---

### Task 2: Tabs component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-tabs.test.ts`
- Create: `web/packages/ui/src/layout/wf-tabs.ts`
- Create: `web/packages/ui/src/layout/wf-tab-panel.ts`

- [ ] **Step 1: Write tabs tests**

```typescript
// tests/layout/wf-tabs.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-tabs.js';
import '../../src/layout/wf-tab-panel.js';
import type { WfTabs } from '../../src/layout/wf-tabs.js';
import type { WfTabPanel } from '../../src/layout/wf-tab-panel.js';

async function createTabs(): Promise<WfTabs> {
  const el = await fixture<WfTabs>('wf-tabs');
  el.innerHTML = `
    <wf-tab-panel name="one" label="Tab One">Content One</wf-tab-panel>
    <wf-tab-panel name="two" label="Tab Two">Content Two</wf-tab-panel>
    <wf-tab-panel name="three" label="Tab Three">Content Three</wf-tab-panel>
  `;
  await el.updateComplete;
  // Allow child panels to upgrade
  await Promise.all(
    Array.from(el.querySelectorAll('wf-tab-panel')).map(
      (p) => (p as WfTabPanel).updateComplete,
    ),
  );
  el.selectTab('one');
  await el.updateComplete;
  return el;
}

describe('WfTabs', () => {
  afterEach(cleanup);

  it('renders with wf-tabs class', async () => {
    const el = await createTabs();
    expect(el.classList.contains('wf-tabs')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await createTabs();
    expect(el.shadowRoot).toBeNull();
  });

  it('renders tab buttons for each panel', async () => {
    const el = await createTabs();
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    expect(buttons.length).toBe(3);
    expect(buttons[0].textContent!.trim()).toBe('Tab One');
  });

  it('shows first panel by default', async () => {
    const el = await createTabs();
    const panels = el.querySelectorAll('wf-tab-panel');
    expect(panels[0].hasAttribute('active')).toBe(true);
    expect(panels[1].hasAttribute('active')).toBe(false);
  });

  it('switches panels on tab click', async () => {
    const el = await createTabs();
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    (buttons[1] as HTMLElement).click();
    await el.updateComplete;
    const panels = el.querySelectorAll('wf-tab-panel');
    expect(panels[0].hasAttribute('active')).toBe(false);
    expect(panels[1].hasAttribute('active')).toBe(true);
  });

  it('fires wf-tab-change event', async () => {
    const el = await createTabs();
    const handler = vi.fn();
    el.addEventListener('wf-tab-change', handler);
    el.selectTab('two');
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ name: 'two' });
  });

  it('supports keyboard navigation (ArrowRight)', async () => {
    const el = await createTabs();
    const tabList = el.querySelector('.wf-tabs__list') as HTMLElement;
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    (buttons[0] as HTMLElement).focus();
    tabList.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }));
    await el.updateComplete;
    expect(el.activeTab).toBe('two');
  });

  it('tab buttons have correct ARIA attributes', async () => {
    const el = await createTabs();
    const buttons = el.querySelectorAll('.wf-tabs__tab');
    expect(buttons[0].getAttribute('role')).toBe('tab');
    expect(buttons[0].getAttribute('aria-selected')).toBe('true');
    expect(buttons[1].getAttribute('aria-selected')).toBe('false');
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-tabs.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfTabPanel**

```typescript
// src/layout/wf-tab-panel.ts
import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-tab-panel>` — A panel within a `<wf-tabs>` container.
 *
 * @element wf-tab-panel
 * @slot - Default slot for panel content.
 */
export class WfTabPanel extends WfElement {
  static get properties() {
    return {
      name: { type: String, reflect: true },
      label: { type: String, reflect: true },
      active: { type: Boolean, reflect: true },
    };
  }

  name = '';
  label = '';
  active = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-tab-panel');
    this.setAttribute('role', 'tabpanel');
    this._syncVisibility();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncVisibility();
  }

  private _syncVisibility(): void {
    this.style.display = this.active ? '' : 'none';
  }

  render() {
    return html`<slot></slot>`;
  }
}

customElements.define('wf-tab-panel', WfTabPanel);

declare global {
  interface HTMLElementTagNameMap {
    'wf-tab-panel': WfTabPanel;
  }
}
```

- [ ] **Step 4: Implement WfTabs**

```typescript
// src/layout/wf-tabs.ts
import { html } from 'lit';
import { WfElement } from '../base.js';
import type { WfTabPanel } from './wf-tab-panel.js';

/**
 * `<wf-tabs>` — Tab bar with panels. Add `<wf-tab-panel>` children.
 *
 * @element wf-tabs
 * @fires wf-tab-change — When the active tab changes.
 */
export class WfTabs extends WfElement {
  static get properties() {
    return {
      activeTab: { type: String, attribute: 'active-tab', reflect: true },
    };
  }

  activeTab = '';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-tabs');
  }

  firstUpdated(): void {
    // Default to first panel if no active tab set
    if (!this.activeTab) {
      const first = this.querySelector('wf-tab-panel') as WfTabPanel | null;
      if (first) {
        this.activeTab = first.name;
        this._syncPanels();
      }
    }
  }

  selectTab(name: string): void {
    this.activeTab = name;
    this._syncPanels();
    this.requestUpdate();
    this.dispatchEvent(
      new CustomEvent('wf-tab-change', {
        detail: { name },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _syncPanels(): void {
    const panels = this.querySelectorAll('wf-tab-panel') as NodeListOf<WfTabPanel>;
    panels.forEach((panel) => {
      panel.active = panel.name === this.activeTab;
    });
  }

  private _handleTabClick(name: string): void {
    this.selectTab(name);
  }

  private _handleKeydown(e: KeyboardEvent): void {
    const panels = Array.from(
      this.querySelectorAll('wf-tab-panel'),
    ) as WfTabPanel[];
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
    const buttons = this.querySelectorAll('.wf-tabs__tab');
    (buttons[newIndex] as HTMLElement)?.focus();
  }

  render() {
    const panels = Array.from(
      this.querySelectorAll('wf-tab-panel'),
    ) as WfTabPanel[];

    return html`
      <div class="wf-tabs__list" role="tablist" @keydown=${this._handleKeydown}>
        ${panels.map(
          (panel) => html`
            <button
              class="wf-tabs__tab ${panel.name === this.activeTab
                ? 'wf-tabs__tab--active'
                : ''}"
              role="tab"
              aria-selected=${panel.name === this.activeTab}
              tabindex=${panel.name === this.activeTab ? '0' : '-1'}
              @click=${() => this._handleTabClick(panel.name)}
            >
              ${panel.label}
            </button>
          `,
        )}
      </div>
    `;
  }
}

customElements.define('wf-tabs', WfTabs);

declare global {
  interface HTMLElementTagNameMap {
    'wf-tabs': WfTabs;
  }
}
```

- [ ] **Step 5: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-tabs.test.ts 2>&1 | tail -5
```

Expected: 8 tests pass.

- [ ] **Step 6: Commit**

```bash
git add web/packages/ui/src/layout/wf-tabs.ts web/packages/ui/src/layout/wf-tab-panel.ts web/packages/ui/tests/layout/wf-tabs.test.ts
git commit -m "feat(ui): add wf-tabs and wf-tab-panel with keyboard navigation and ARIA"
```

---

### Task 3: Accordion component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-accordion.test.ts`
- Create: `web/packages/ui/src/layout/wf-accordion.ts`
- Create: `web/packages/ui/src/layout/wf-accordion-item.ts`

- [ ] **Step 1: Write accordion tests**

```typescript
// tests/layout/wf-accordion.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-accordion.js';
import '../../src/layout/wf-accordion-item.js';
import type { WfAccordion } from '../../src/layout/wf-accordion.js';
import type { WfAccordionItem } from '../../src/layout/wf-accordion-item.js';

async function createAccordion(multiple = false): Promise<WfAccordion> {
  const el = await fixture<WfAccordion>('wf-accordion');
  if (multiple) el.setAttribute('multiple', '');
  el.innerHTML = `
    <wf-accordion-item name="one" header="Section One">Content One</wf-accordion-item>
    <wf-accordion-item name="two" header="Section Two">Content Two</wf-accordion-item>
    <wf-accordion-item name="three" header="Section Three">Content Three</wf-accordion-item>
  `;
  await el.updateComplete;
  await Promise.all(
    Array.from(el.querySelectorAll('wf-accordion-item')).map(
      (p) => (p as WfAccordionItem).updateComplete,
    ),
  );
  return el;
}

describe('WfAccordion', () => {
  afterEach(cleanup);

  it('renders with wf-accordion class', async () => {
    const el = await createAccordion();
    expect(el.classList.contains('wf-accordion')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await createAccordion();
    expect(el.shadowRoot).toBeNull();
  });

  it('all items collapsed by default', async () => {
    const el = await createAccordion();
    const items = el.querySelectorAll('wf-accordion-item');
    items.forEach((item) => {
      expect(item.hasAttribute('expanded')).toBe(false);
    });
  });

  it('clicking header expands item', async () => {
    const el = await createAccordion();
    const items = el.querySelectorAll('wf-accordion-item');
    const header = items[0].querySelector('.wf-accordion-item__header') as HTMLElement;
    header.click();
    await (items[0] as WfAccordionItem).updateComplete;
    expect(items[0].hasAttribute('expanded')).toBe(true);
  });

  it('single mode: opening one closes others', async () => {
    const el = await createAccordion(false);
    const items = el.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
    items[0].toggle();
    await el.updateComplete;
    expect(items[0].expanded).toBe(true);
    items[1].toggle();
    await el.updateComplete;
    expect(items[0].expanded).toBe(false);
    expect(items[1].expanded).toBe(true);
  });

  it('multiple mode: opening one does not close others', async () => {
    const el = await createAccordion(true);
    const items = el.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
    items[0].toggle();
    await el.updateComplete;
    items[1].toggle();
    await el.updateComplete;
    expect(items[0].expanded).toBe(true);
    expect(items[1].expanded).toBe(true);
  });

  it('fires wf-accordion-change event', async () => {
    const el = await createAccordion();
    const handler = vi.fn();
    el.addEventListener('wf-accordion-change', handler);
    const items = el.querySelectorAll('wf-accordion-item') as NodeListOf<WfAccordionItem>;
    items[0].toggle();
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ name: 'one', expanded: true });
  });

  it('header has correct ARIA attributes', async () => {
    const el = await createAccordion();
    const item = el.querySelector('wf-accordion-item') as WfAccordionItem;
    const header = item.querySelector('.wf-accordion-item__header') as HTMLElement;
    expect(header.getAttribute('role')).toBe('button');
    expect(header.getAttribute('aria-expanded')).toBe('false');
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-accordion.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfAccordionItem**

```typescript
// src/layout/wf-accordion-item.ts
import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-accordion-item>` — Collapsible section within a `<wf-accordion>`.
 *
 * @element wf-accordion-item
 * @slot - Default slot for collapsed content.
 * @fires wf-accordion-change — Bubbles to parent accordion on toggle.
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
```

- [ ] **Step 4: Implement WfAccordion**

```typescript
// src/layout/wf-accordion.ts
import { html } from 'lit';
import { WfElement } from '../base.js';
import type { WfAccordionItem } from './wf-accordion-item.js';

/**
 * `<wf-accordion>` — Container for collapsible sections.
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

  render() {
    return html`<slot></slot>`;
  }
}

customElements.define('wf-accordion', WfAccordion);

declare global {
  interface HTMLElementTagNameMap {
    'wf-accordion': WfAccordion;
  }
}
```

- [ ] **Step 5: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-accordion.test.ts 2>&1 | tail -5
```

Expected: 8 tests pass.

- [ ] **Step 6: Commit**

```bash
git add web/packages/ui/src/layout/wf-accordion.ts web/packages/ui/src/layout/wf-accordion-item.ts web/packages/ui/tests/layout/wf-accordion.test.ts
git commit -m "feat(ui): add wf-accordion and wf-accordion-item with single/multiple modes"
```

---

## Chunk 2: Overlay Utility + Overlay Components (Dialog, Drawer, Tooltip, Popover)

### Task 4: Shared overlay utility

**Files:**
- Create: `web/packages/ui/src/utils/overlay.ts`

- [ ] **Step 1: Create utils directory**

```bash
mkdir -p web/packages/ui/src/utils
```

- [ ] **Step 2: Implement overlay utility**

```typescript
// src/utils/overlay.ts

/**
 * Lightweight overlay utility for focus trapping and backdrop management.
 * Used by wf-dialog, wf-drawer, wf-tooltip, and wf-popover.
 */

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

/**
 * Trap focus within a container element.
 * Returns a cleanup function that removes the event listener.
 */
export function trapFocus(container: HTMLElement): () => void {
  const handler = (e: KeyboardEvent) => {
    if (e.key !== 'Tab') return;

    const focusable = Array.from(
      container.querySelectorAll(FOCUSABLE_SELECTOR),
    ) as HTMLElement[];
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];

    if (e.shiftKey) {
      if (document.activeElement === first) {
        e.preventDefault();
        last.focus();
      }
    } else {
      if (document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }
  };

  container.addEventListener('keydown', handler);
  return () => container.removeEventListener('keydown', handler);
}

/**
 * Create and append a backdrop element to document.body.
 * Returns the backdrop element. Caller is responsible for removal.
 */
export function createBackdrop(onClick?: () => void): HTMLDivElement {
  const backdrop = document.createElement('div');
  backdrop.classList.add('wf-overlay-backdrop');
  if (onClick) {
    backdrop.addEventListener('click', onClick);
  }
  document.body.appendChild(backdrop);
  return backdrop;
}

/**
 * Remove a backdrop element from the DOM.
 */
export function removeBackdrop(backdrop: HTMLDivElement): void {
  backdrop.remove();
}

/**
 * Listen for Escape key on a target element.
 * Returns a cleanup function.
 */
export function onEscape(
  target: HTMLElement | Document,
  callback: () => void,
): () => void {
  const handler = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.stopPropagation();
      callback();
    }
  };
  target.addEventListener('keydown', handler as EventListener);
  return () => target.removeEventListener('keydown', handler as EventListener);
}
```

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui/src/utils/overlay.ts
git commit -m "feat(ui): add shared overlay utility (focus trap, backdrop, escape handler)"
```

---

### Task 5: Dialog/Modal component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-dialog.test.ts`
- Create: `web/packages/ui/src/layout/wf-dialog.ts`

- [ ] **Step 1: Write dialog tests**

```typescript
// tests/layout/wf-dialog.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-dialog.js';
import type { WfDialog } from '../../src/layout/wf-dialog.js';

describe('WfDialog', () => {
  afterEach(cleanup);

  it('renders with wf-dialog class', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    expect(el.classList.contains('wf-dialog')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    expect(el.shadowRoot).toBeNull();
  });

  it('is hidden by default', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    expect(el.open).toBe(false);
    expect(el.classList.contains('wf-dialog--open')).toBe(false);
  });

  it('renders dialog content when opened', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.innerHTML = '<p>Dialog content</p>';
    el.show();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-dialog--open')).toBe(true);
  });

  it('renders header when set', async () => {
    const el = await fixture<WfDialog>('wf-dialog', { header: 'Confirm' });
    el.show();
    await el.updateComplete;
    const header = el.querySelector('.wf-dialog__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toContain('Confirm');
  });

  it('creates backdrop when opened', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    const backdrop = document.querySelector('.wf-overlay-backdrop');
    expect(backdrop).not.toBeNull();
    el.hide();
    await el.updateComplete;
  });

  it('removes backdrop when closed', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    el.hide();
    await el.updateComplete;
    const backdrop = document.querySelector('.wf-overlay-backdrop');
    expect(backdrop).toBeNull();
  });

  it('fires wf-close event on hide', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfDialog>('wf-dialog');
    el.show();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('has correct ARIA attributes', async () => {
    const el = await fixture<WfDialog>('wf-dialog', { header: 'Title' });
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-dialog__panel');
    expect(panel!.getAttribute('role')).toBe('dialog');
    expect(panel!.getAttribute('aria-modal')).toBe('true');
  });

  it('close button calls hide()', async () => {
    const el = await fixture<WfDialog>('wf-dialog', { header: 'Title' });
    el.show();
    await el.updateComplete;
    const closeBtn = el.querySelector('.wf-dialog__close') as HTMLElement;
    expect(closeBtn).not.toBeNull();
    closeBtn.click();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-dialog.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfDialog**

```typescript
// src/layout/wf-dialog.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';
import { trapFocus, createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

/**
 * `<wf-dialog>` — Modal dialog with backdrop, focus trap, and Escape dismissal.
 *
 * @element wf-dialog
 * @slot - Default slot for dialog body content.
 * @fires wf-close — When the dialog is closed.
 */
export class WfDialog extends WfElement {
  static get properties() {
    return {
      open: { type: Boolean, reflect: true },
      header: { type: String, reflect: true },
    };
  }

  open = false;
  header = '';

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupFocus: (() => void) | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _previousFocus: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-dialog');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  show(): void {
    this._previousFocus = document.activeElement as HTMLElement | null;
    this.open = true;
    this.classList.add('wf-dialog--open');
    this.requestUpdate();

    this._backdrop = createBackdrop(() => this.hide());
    this._cleanupEscape = onEscape(this, () => this.hide());

    // Defer focus trap setup to after render
    requestAnimationFrame(() => {
      const panel = this.querySelector('.wf-dialog__panel') as HTMLElement;
      if (panel) {
        this._cleanupFocus = trapFocus(panel);
        // Focus the first focusable element or the panel itself
        const firstFocusable = panel.querySelector(
          'button, [tabindex]:not([tabindex="-1"])',
        ) as HTMLElement | null;
        if (firstFocusable) firstFocusable.focus();
      }
    });
  }

  hide(): void {
    this.open = false;
    this.classList.remove('wf-dialog--open');
    this._teardown();
    this.requestUpdate();

    this.dispatchEvent(
      new CustomEvent('wf-close', {
        bubbles: true,
        composed: true,
      }),
    );

    // Restore focus
    if (this._previousFocus) {
      this._previousFocus.focus();
      this._previousFocus = null;
    }
  }

  private _teardown(): void {
    if (this._backdrop) {
      removeBackdrop(this._backdrop);
      this._backdrop = null;
    }
    if (this._cleanupFocus) {
      this._cleanupFocus();
      this._cleanupFocus = null;
    }
    if (this._cleanupEscape) {
      this._cleanupEscape();
      this._cleanupEscape = null;
    }
  }

  render() {
    if (!this.open) return nothing;

    return html`
      <div
        class="wf-dialog__panel"
        role="dialog"
        aria-modal="true"
        aria-label=${this.header || 'Dialog'}
      >
        ${this.header
          ? html`
              <div class="wf-dialog__header">
                <span class="wf-dialog__title">${this.header}</span>
                <button
                  class="wf-dialog__close"
                  aria-label="Close"
                  @click=${() => this.hide()}
                >&times;</button>
              </div>
            `
          : nothing}
        <div class="wf-dialog__body">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define('wf-dialog', WfDialog);

declare global {
  interface HTMLElementTagNameMap {
    'wf-dialog': WfDialog;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-dialog.test.ts 2>&1 | tail -5
```

Expected: 11 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/layout/wf-dialog.ts web/packages/ui/tests/layout/wf-dialog.test.ts
git commit -m "feat(ui): add wf-dialog with backdrop, focus trap, and escape dismissal"
```

---

### Task 6: Drawer component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-drawer.test.ts`
- Create: `web/packages/ui/src/layout/wf-drawer.ts`

- [ ] **Step 1: Write drawer tests**

```typescript
// tests/layout/wf-drawer.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-drawer.js';
import type { WfDrawer } from '../../src/layout/wf-drawer.js';

describe('WfDrawer', () => {
  afterEach(cleanup);

  it('renders with wf-drawer class', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    expect(el.classList.contains('wf-drawer')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    expect(el.shadowRoot).toBeNull();
  });

  it('is hidden by default', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    expect(el.open).toBe(false);
  });

  it('shows panel when opened', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-drawer--open')).toBe(true);
    const panel = el.querySelector('.wf-drawer__panel');
    expect(panel).not.toBeNull();
    el.hide();
  });

  it('applies position class', async () => {
    const el = await fixture<WfDrawer>('wf-drawer', { position: 'left' });
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-drawer__panel');
    expect(panel!.classList.contains('wf-drawer__panel--left')).toBe(true);
    el.hide();
  });

  it('default position is right', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('.wf-drawer__panel');
    expect(panel!.classList.contains('wf-drawer__panel--right')).toBe(true);
    el.hide();
  });

  it('creates backdrop when opened', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    expect(document.querySelector('.wf-overlay-backdrop')).not.toBeNull();
    el.hide();
    await el.updateComplete;
  });

  it('fires wf-close event on hide', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfDrawer>('wf-drawer');
    el.show();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('renders header when set', async () => {
    const el = await fixture<WfDrawer>('wf-drawer', { header: 'Settings' });
    el.show();
    await el.updateComplete;
    const header = el.querySelector('.wf-drawer__header');
    expect(header).not.toBeNull();
    expect(header!.textContent).toContain('Settings');
    el.hide();
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-drawer.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfDrawer**

```typescript
// src/layout/wf-drawer.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';
import { trapFocus, createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

/**
 * `<wf-drawer>` — Slide-in panel from screen edge. Supports left, right, top, bottom.
 *
 * @element wf-drawer
 * @slot - Default slot for drawer content.
 * @fires wf-close — When the drawer is closed.
 */
export class WfDrawer extends WfElement {
  static get properties() {
    return {
      open: { type: Boolean, reflect: true },
      header: { type: String, reflect: true },
      position: { type: String, reflect: true },
    };
  }

  open = false;
  header = '';
  position: 'left' | 'right' | 'top' | 'bottom' = 'right';

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupFocus: (() => void) | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _previousFocus: HTMLElement | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-drawer');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  show(): void {
    this._previousFocus = document.activeElement as HTMLElement | null;
    this.open = true;
    this.classList.add('wf-drawer--open');
    this.requestUpdate();

    this._backdrop = createBackdrop(() => this.hide());
    this._cleanupEscape = onEscape(this, () => this.hide());

    requestAnimationFrame(() => {
      const panel = this.querySelector('.wf-drawer__panel') as HTMLElement;
      if (panel) {
        this._cleanupFocus = trapFocus(panel);
      }
    });
  }

  hide(): void {
    this.open = false;
    this.classList.remove('wf-drawer--open');
    this._teardown();
    this.requestUpdate();

    this.dispatchEvent(
      new CustomEvent('wf-close', {
        bubbles: true,
        composed: true,
      }),
    );

    if (this._previousFocus) {
      this._previousFocus.focus();
      this._previousFocus = null;
    }
  }

  private _teardown(): void {
    if (this._backdrop) {
      removeBackdrop(this._backdrop);
      this._backdrop = null;
    }
    if (this._cleanupFocus) {
      this._cleanupFocus();
      this._cleanupFocus = null;
    }
    if (this._cleanupEscape) {
      this._cleanupEscape();
      this._cleanupEscape = null;
    }
  }

  render() {
    if (!this.open) return nothing;

    return html`
      <div
        class="wf-drawer__panel wf-drawer__panel--${this.position}"
        role="dialog"
        aria-modal="true"
        aria-label=${this.header || 'Drawer'}
      >
        ${this.header
          ? html`
              <div class="wf-drawer__header">
                <span class="wf-drawer__title">${this.header}</span>
                <button
                  class="wf-drawer__close"
                  aria-label="Close"
                  @click=${() => this.hide()}
                >&times;</button>
              </div>
            `
          : nothing}
        <div class="wf-drawer__body">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define('wf-drawer', WfDrawer);

declare global {
  interface HTMLElementTagNameMap {
    'wf-drawer': WfDrawer;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-drawer.test.ts 2>&1 | tail -5
```

Expected: 10 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/layout/wf-drawer.ts web/packages/ui/tests/layout/wf-drawer.test.ts
git commit -m "feat(ui): add wf-drawer with slide-in positions and overlay management"
```

---

### Task 7: Tooltip component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-tooltip.test.ts`
- Create: `web/packages/ui/src/layout/wf-tooltip.ts`

- [ ] **Step 1: Write tooltip tests**

```typescript
// tests/layout/wf-tooltip.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-tooltip.js';
import type { WfTooltip } from '../../src/layout/wf-tooltip.js';

describe('WfTooltip', () => {
  afterEach(cleanup);

  it('renders with wf-tooltip class', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    expect(el.classList.contains('wf-tooltip')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip');
    expect(el.shadowRoot).toBeNull();
  });

  it('tooltip content is hidden by default', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help text' });
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip).not.toBeNull();
    expect(tip.getAttribute('aria-hidden')).toBe('true');
  });

  it('shows tooltip on mouseenter', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help' });
    await el.updateComplete;
    el.dispatchEvent(new MouseEvent('mouseenter'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('false');
  });

  it('hides tooltip on mouseleave', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help' });
    await el.updateComplete;
    el.dispatchEvent(new MouseEvent('mouseenter'));
    await el.updateComplete;
    el.dispatchEvent(new MouseEvent('mouseleave'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('true');
  });

  it('shows tooltip on focusin', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help' });
    await el.updateComplete;
    el.dispatchEvent(new FocusEvent('focusin'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('false');
  });

  it('hides tooltip on focusout', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help' });
    await el.updateComplete;
    el.dispatchEvent(new FocusEvent('focusin'));
    await el.updateComplete;
    el.dispatchEvent(new FocusEvent('focusout'));
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.getAttribute('aria-hidden')).toBe('true');
  });

  it('applies position class', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', {
      content: 'Help',
      position: 'bottom',
    });
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.classList.contains('wf-tooltip__content--bottom')).toBe(true);
  });

  it('default position is top', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help' });
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content') as HTMLElement;
    expect(tip.classList.contains('wf-tooltip__content--top')).toBe(true);
  });

  it('has correct ARIA attributes', async () => {
    const el = await fixture<WfTooltip>('wf-tooltip', { content: 'Help' });
    await el.updateComplete;
    const tip = el.querySelector('.wf-tooltip__content');
    expect(tip!.getAttribute('role')).toBe('tooltip');
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-tooltip.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfTooltip**

```typescript
// src/layout/wf-tooltip.ts
import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-tooltip>` — Hover/focus popup that shows text content.
 * Wrap the trigger element as a child. Set `content` for the tooltip text.
 *
 * @element wf-tooltip
 * @slot - Default slot for the trigger element.
 */
export class WfTooltip extends WfElement {
  static get properties() {
    return {
      content: { type: String, reflect: true },
      position: { type: String, reflect: true },
      _visible: { type: Boolean },
    };
  }

  content = '';
  position: 'top' | 'bottom' | 'left' | 'right' = 'top';
  _visible = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-tooltip');
    this.addEventListener('mouseenter', this._show);
    this.addEventListener('mouseleave', this._hide);
    this.addEventListener('focusin', this._show);
    this.addEventListener('focusout', this._hide);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('mouseenter', this._show);
    this.removeEventListener('mouseleave', this._hide);
    this.removeEventListener('focusin', this._show);
    this.removeEventListener('focusout', this._hide);
  }

  private _show = (): void => {
    this._visible = true;
  };

  private _hide = (): void => {
    this._visible = false;
  };

  render() {
    return html`
      <slot></slot>
      <span
        class="wf-tooltip__content wf-tooltip__content--${this.position}"
        role="tooltip"
        aria-hidden=${!this._visible}
      >
        ${this.content}
      </span>
    `;
  }
}

customElements.define('wf-tooltip', WfTooltip);

declare global {
  interface HTMLElementTagNameMap {
    'wf-tooltip': WfTooltip;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-tooltip.test.ts 2>&1 | tail -5
```

Expected: 10 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/layout/wf-tooltip.ts web/packages/ui/tests/layout/wf-tooltip.test.ts
git commit -m "feat(ui): add wf-tooltip with hover/focus trigger and ARIA support"
```

---

### Task 8: Popover component

**Files:**
- Create: `web/packages/ui/tests/layout/wf-popover.test.ts`
- Create: `web/packages/ui/src/layout/wf-popover.ts`

- [ ] **Step 1: Write popover tests**

```typescript
// tests/layout/wf-popover.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-popover.js';
import type { WfPopover } from '../../src/layout/wf-popover.js';

describe('WfPopover', () => {
  afterEach(cleanup);

  it('renders with wf-popover class', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    expect(el.classList.contains('wf-popover')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    expect(el.shadowRoot).toBeNull();
  });

  it('content is hidden by default', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    expect(el.open).toBe(false);
  });

  it('toggle() opens the popover', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    expect(el.open).toBe(true);
    expect(el.classList.contains('wf-popover--open')).toBe(true);
  });

  it('toggle() closes when open', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    el.toggle();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('closes on outside click', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    // Simulate click outside
    document.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('fires wf-close event when closed', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-close', handler);
    el.hide();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('closes on Escape key', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    el.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('applies position class', async () => {
    const el = await fixture<WfPopover>('wf-popover', { position: 'left' });
    el.toggle();
    await el.updateComplete;
    const content = el.querySelector('.wf-popover__content');
    expect(content!.classList.contains('wf-popover__content--left')).toBe(true);
  });

  it('default position is bottom', async () => {
    const el = await fixture<WfPopover>('wf-popover');
    el.toggle();
    await el.updateComplete;
    const content = el.querySelector('.wf-popover__content');
    expect(content!.classList.contains('wf-popover__content--bottom')).toBe(true);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-popover.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfPopover**

```typescript
// src/layout/wf-popover.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';
import { onEscape } from '../utils/overlay.js';

/**
 * `<wf-popover>` — Click-triggered popup with rich content.
 * Wrap the trigger element as the default slot. Use the `content` slot
 * or set innerHTML in the popover body.
 *
 * @element wf-popover
 * @slot - Default slot for the trigger element.
 * @fires wf-close — When the popover is closed.
 */
export class WfPopover extends WfElement {
  static get properties() {
    return {
      open: { type: Boolean, reflect: true },
      position: { type: String, reflect: true },
    };
  }

  open = false;
  position: 'top' | 'bottom' | 'left' | 'right' = 'bottom';

  private _cleanupEscape: (() => void) | null = null;
  private _boundDocClick: ((e: Event) => void) | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-popover');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  toggle(): void {
    if (this.open) {
      this.hide();
    } else {
      this.show();
    }
  }

  show(): void {
    this.open = true;
    this.classList.add('wf-popover--open');
    this.requestUpdate();

    this._cleanupEscape = onEscape(this, () => this.hide());
    // Close on outside click (deferred to avoid catching the opening click)
    requestAnimationFrame(() => {
      this._boundDocClick = (e: Event) => {
        if (!this.contains(e.target as Node)) {
          this.hide();
        }
      };
      document.addEventListener('click', this._boundDocClick);
    });
  }

  hide(): void {
    this.open = false;
    this.classList.remove('wf-popover--open');
    this._teardown();
    this.requestUpdate();

    this.dispatchEvent(
      new CustomEvent('wf-close', {
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _teardown(): void {
    if (this._cleanupEscape) {
      this._cleanupEscape();
      this._cleanupEscape = null;
    }
    if (this._boundDocClick) {
      document.removeEventListener('click', this._boundDocClick);
      this._boundDocClick = null;
    }
  }

  render() {
    return html`
      <div class="wf-popover__trigger" @click=${(e: Event) => { e.stopPropagation(); this.toggle(); }}>
        <slot></slot>
      </div>
      ${this.open
        ? html`
            <div class="wf-popover__content wf-popover__content--${this.position}">
              <slot name="content"></slot>
            </div>
          `
        : nothing}
    `;
  }
}

customElements.define('wf-popover', WfPopover);

declare global {
  interface HTMLElementTagNameMap {
    'wf-popover': WfPopover;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-popover.test.ts 2>&1 | tail -5
```

Expected: 10 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/layout/wf-popover.ts web/packages/ui/tests/layout/wf-popover.test.ts
git commit -m "feat(ui): add wf-popover with click trigger and outside-click dismissal"
```

---

## Chunk 3: Table/Data Grid

This is the most complex component. It is built incrementally: basic table first, then sorting, then column resize, then pagination.

### Task 9: Basic table with static rendering

**Files:**
- Create: `web/packages/ui/tests/layout/wf-table.test.ts`
- Create: `web/packages/ui/src/layout/wf-table.ts`

- [ ] **Step 1: Write basic table tests**

```typescript
// tests/layout/wf-table.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-table.js';
import type { WfTable, WfTableColumn } from '../../src/layout/wf-table.js';

const COLUMNS: WfTableColumn[] = [
  { key: 'name', header: 'Name' },
  { key: 'email', header: 'Email' },
  { key: 'role', header: 'Role' },
];

const DATA = [
  { name: 'Alice', email: 'alice@example.com', role: 'Admin' },
  { name: 'Bob', email: 'bob@example.com', role: 'User' },
  { name: 'Charlie', email: 'charlie@example.com', role: 'User' },
];

describe('WfTable', () => {
  afterEach(cleanup);

  it('renders with wf-table class', async () => {
    const el = await fixture<WfTable>('wf-table');
    expect(el.classList.contains('wf-table')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfTable>('wf-table');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders table headers from columns', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    const ths = el.querySelectorAll('th');
    expect(ths.length).toBe(3);
    expect(ths[0].textContent!.trim()).toBe('Name');
    expect(ths[1].textContent!.trim()).toBe('Email');
  });

  it('renders data rows', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(3);
    const cells = rows[0].querySelectorAll('td');
    expect(cells[0].textContent!.trim()).toBe('Alice');
    expect(cells[1].textContent!.trim()).toBe('alice@example.com');
  });

  it('renders empty state when no data', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = [];
    await el.updateComplete;
    const empty = el.querySelector('.wf-table__empty');
    expect(empty).not.toBeNull();
  });

  it('applies striped variant', async () => {
    const el = await fixture<WfTable>('wf-table', { striped: true });
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    expect(el.classList.contains('wf-table--striped')).toBe(true);
  });

  it('fires wf-row-click event on row click', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = COLUMNS;
    el.data = DATA;
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-row-click', handler);
    const row = el.querySelector('tbody tr') as HTMLElement;
    row.click();
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ row: DATA[0], index: 0 });
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-table.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfTable (basic)**

```typescript
// src/layout/wf-table.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';

export interface WfTableColumn {
  key: string;
  header: string;
  sortable?: boolean;
  width?: string;
  render?: (value: unknown, row: Record<string, unknown>) => unknown;
}

/**
 * `<wf-table>` — Data table with sorting, pagination, and row click events.
 * Set `columns` and `data` properties to render.
 *
 * @element wf-table
 * @fires wf-row-click — When a row is clicked. Detail: { row, index }.
 * @fires wf-sort — When a column sort is triggered. Detail: { key, direction }.
 * @fires wf-page-change — When pagination changes. Detail: { page }.
 */
export class WfTable extends WfElement {
  static get properties() {
    return {
      columns: { type: Array },
      data: { type: Array },
      striped: { type: Boolean, reflect: true },
      sortKey: { type: String, attribute: 'sort-key' },
      sortDirection: { type: String, attribute: 'sort-direction' },
      page: { type: Number },
      pageSize: { type: Number, attribute: 'page-size' },
      paginate: { type: Boolean, reflect: true },
    };
  }

  columns: WfTableColumn[] = [];
  data: Array<Record<string, unknown>> = [];
  striped = false;
  sortKey = '';
  sortDirection: 'asc' | 'desc' | '' = '';
  page = 1;
  pageSize = 10;
  paginate = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-table');
    this._syncClasses();
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    this._syncClasses();
  }

  private _syncClasses(): void {
    this.classList.toggle('wf-table--striped', this.striped);
  }

  /** Sort data by a column key. Toggles asc/desc. */
  sort(key: string): void {
    if (this.sortKey === key) {
      this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortKey = key;
      this.sortDirection = 'asc';
    }
    this.requestUpdate();
    this.dispatchEvent(
      new CustomEvent('wf-sort', {
        detail: { key: this.sortKey, direction: this.sortDirection },
        bubbles: true,
        composed: true,
      }),
    );
  }

  /** Navigate to a page (1-indexed). */
  goToPage(page: number): void {
    this.page = Math.max(1, Math.min(page, this.totalPages));
    this.requestUpdate();
    this.dispatchEvent(
      new CustomEvent('wf-page-change', {
        detail: { page: this.page },
        bubbles: true,
        composed: true,
      }),
    );
  }

  get totalPages(): number {
    if (!this.paginate || this.pageSize <= 0) return 1;
    return Math.ceil(this.data.length / this.pageSize);
  }

  /** Get the data to display — sorted and paginated. */
  get displayData(): Array<Record<string, unknown>> {
    let result = [...this.data];

    // Sort
    if (this.sortKey && this.sortDirection) {
      result.sort((a, b) => {
        const aVal = a[this.sortKey];
        const bVal = b[this.sortKey];
        const cmp =
          typeof aVal === 'string' && typeof bVal === 'string'
            ? aVal.localeCompare(bVal)
            : Number(aVal) - Number(bVal);
        return this.sortDirection === 'desc' ? -cmp : cmp;
      });
    }

    // Paginate
    if (this.paginate && this.pageSize > 0) {
      const start = (this.page - 1) * this.pageSize;
      result = result.slice(start, start + this.pageSize);
    }

    return result;
  }

  private _handleRowClick(row: Record<string, unknown>, index: number): void {
    this.dispatchEvent(
      new CustomEvent('wf-row-click', {
        detail: { row, index },
        bubbles: true,
        composed: true,
      }),
    );
  }

  private _handleSort(key: string): void {
    this.sort(key);
  }

  render() {
    const rows = this.displayData;

    return html`
      <table class="wf-table__table">
        <thead>
          <tr>
            ${this.columns.map(
              (col) => html`
                <th
                  class="wf-table__th ${col.sortable ? 'wf-table__th--sortable' : ''}"
                  style=${col.width ? `width: ${col.width}` : ''}
                  @click=${col.sortable ? () => this._handleSort(col.key) : nothing}
                >
                  ${col.header}
                  ${col.sortable && this.sortKey === col.key
                    ? html`<span class="wf-table__sort-icon">${this.sortDirection === 'asc' ? '\u25B2' : '\u25BC'}</span>`
                    : nothing}
                </th>
              `,
            )}
          </tr>
        </thead>
        <tbody>
          ${rows.length > 0
            ? rows.map(
                (row, i) => html`
                  <tr
                    class="wf-table__row"
                    @click=${() => this._handleRowClick(row, i)}
                  >
                    ${this.columns.map(
                      (col) => html`
                        <td class="wf-table__td">
                          ${col.render
                            ? col.render(row[col.key], row)
                            : row[col.key]}
                        </td>
                      `,
                    )}
                  </tr>
                `,
              )
            : html`
                <tr>
                  <td class="wf-table__empty" colspan=${this.columns.length}>
                    No data
                  </td>
                </tr>
              `}
        </tbody>
      </table>
      ${this.paginate && this.totalPages > 1
        ? html`
            <div class="wf-table__pagination">
              <button
                class="wf-table__page-btn"
                ?disabled=${this.page <= 1}
                @click=${() => this.goToPage(this.page - 1)}
              >
                Previous
              </button>
              <span class="wf-table__page-info">
                Page ${this.page} of ${this.totalPages}
              </span>
              <button
                class="wf-table__page-btn"
                ?disabled=${this.page >= this.totalPages}
                @click=${() => this.goToPage(this.page + 1)}
              >
                Next
              </button>
            </div>
          `
        : nothing}
    `;
  }
}

customElements.define('wf-table', WfTable);

declare global {
  interface HTMLElementTagNameMap {
    'wf-table': WfTable;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-table.test.ts 2>&1 | tail -5
```

Expected: 7 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/layout/wf-table.ts web/packages/ui/tests/layout/wf-table.test.ts
git commit -m "feat(ui): add wf-table with columns, data rendering, and row click events"
```

---

### Task 10: Table sorting and pagination tests

**Files:**
- Modify: `web/packages/ui/tests/layout/wf-table.test.ts`

- [ ] **Step 1: Add sorting and pagination tests**

Append to `tests/layout/wf-table.test.ts`, inside the existing `describe` block:

```typescript
  it('sorts by column when sortable', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [
      { key: 'name', header: 'Name', sortable: true },
      { key: 'email', header: 'Email' },
    ];
    el.data = [
      { name: 'Charlie', email: 'c@x.com' },
      { name: 'Alice', email: 'a@x.com' },
      { name: 'Bob', email: 'b@x.com' },
    ];
    await el.updateComplete;

    el.sort('name');
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Alice');
    expect(rows[2].querySelector('td')!.textContent!.trim()).toBe('Charlie');
  });

  it('toggles sort direction', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name', sortable: true }];
    el.data = [
      { name: 'Alice' },
      { name: 'Bob' },
      { name: 'Charlie' },
    ];
    await el.updateComplete;

    el.sort('name');
    expect(el.sortDirection).toBe('asc');
    el.sort('name');
    expect(el.sortDirection).toBe('desc');
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Charlie');
  });

  it('fires wf-sort event', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name', sortable: true }];
    el.data = [{ name: 'Alice' }];
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-sort', handler);
    el.sort('name');
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ key: 'name', direction: 'asc' });
  });

  it('paginates data', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;

    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(10);
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Row 1');
  });

  it('navigates pages', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;

    el.goToPage(2);
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(10);
    expect(rows[0].querySelector('td')!.textContent!.trim()).toBe('Row 11');
  });

  it('last page shows remaining rows', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;

    el.goToPage(3);
    await el.updateComplete;
    const rows = el.querySelectorAll('tbody tr');
    expect(rows.length).toBe(5);
  });

  it('fires wf-page-change event', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [{ key: 'name', header: 'Name' }];
    el.data = Array.from({ length: 25 }, (_, i) => ({ name: `Row ${i + 1}` }));
    el.paginate = true;
    el.pageSize = 10;
    await el.updateComplete;
    const handler = vi.fn();
    el.addEventListener('wf-page-change', handler);
    el.goToPage(2);
    expect(handler).toHaveBeenCalledOnce();
    expect(handler.mock.calls[0][0].detail).toEqual({ page: 2 });
  });

  it('supports custom column render function', async () => {
    const el = await fixture<WfTable>('wf-table');
    el.columns = [
      {
        key: 'name',
        header: 'Name',
        render: (val: unknown) => `**${val}**`,
      },
    ];
    el.data = [{ name: 'Alice' }];
    await el.updateComplete;
    const td = el.querySelector('tbody td');
    expect(td!.textContent!.trim()).toBe('**Alice**');
  });
```

- [ ] **Step 2: Verify all table tests pass**

```bash
cd web/packages/ui && npx vitest run tests/layout/wf-table.test.ts 2>&1 | tail -5
```

Expected: 14 tests pass (7 basic + 7 sorting/pagination).

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui/tests/layout/wf-table.test.ts
git commit -m "test(ui): add sorting, pagination, and custom render tests for wf-table"
```

---

## Chunk 4: CSS Styles for Layout Components

### Task 11: Layout CSS file

**Files:**
- Create: `web/packages/ui/src/styles/layout.css`
- Modify: `web/packages/ui/src/styles/components.css`

- [ ] **Step 1: Create layout.css**

```css
/* layout.css — Styles for layout and display components (Phase 3) */

/* ── Layout component tokens ── */
:root {
  --wf-card-padding: var(--wf-space-md);
  --wf-card-radius: var(--wf-radius-md);
  --wf-dialog-max-width: 32rem;
  --wf-drawer-width: 20rem;
  --wf-tooltip-max-width: 16rem;
  --wf-popover-max-width: 20rem;
}

/* ── Card ── */
.wf-card {
  display: block;
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-card-radius);
  font-family: var(--wf-font-sans);
  color: var(--wf-color-text);
  overflow: hidden;
}

.wf-card--padded .wf-card__body {
  padding: var(--wf-card-padding);
}

.wf-card--outlined {
  background: transparent;
}

.wf-card--elevated {
  border-color: transparent;
  box-shadow: var(--wf-shadow-md);
}

.wf-card__header {
  padding: var(--wf-space-sm) var(--wf-card-padding);
  font-weight: var(--wf-weight-semibold);
  font-size: var(--wf-text-sm);
  border-bottom: 1px solid var(--wf-color-border);
}

.wf-card__body {
  /* padding controlled by --padded variant */
}

.wf-card__footer {
  padding: var(--wf-space-sm) var(--wf-card-padding);
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text-secondary);
  border-top: 1px solid var(--wf-color-border);
}

/* ── Tabs ── */
.wf-tabs {
  display: block;
  font-family: var(--wf-font-sans);
}

.wf-tabs__list {
  display: flex;
  border-bottom: 1px solid var(--wf-color-border);
  gap: 0;
}

.wf-tabs__tab {
  padding: var(--wf-space-sm) var(--wf-space-md);
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--wf-color-text-secondary);
  font-family: inherit;
  font-size: var(--wf-text-sm);
  cursor: pointer;
  transition: color var(--wf-duration-fast) var(--wf-ease-in-out),
    border-color var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-tabs__tab:hover {
  color: var(--wf-color-text);
}

.wf-tabs__tab--active {
  color: var(--wf-color-text);
  border-bottom-color: var(--wf-color-accent);
}

.wf-tab-panel {
  display: block;
  padding: var(--wf-space-md) 0;
}

/* ── Accordion ── */
.wf-accordion {
  display: block;
  font-family: var(--wf-font-sans);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-radius-md);
  overflow: hidden;
}

.wf-accordion-item {
  display: block;
}

.wf-accordion-item + .wf-accordion-item {
  border-top: 1px solid var(--wf-color-border);
}

.wf-accordion-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--wf-space-sm) var(--wf-space-md);
  cursor: pointer;
  font-size: var(--wf-text-sm);
  font-weight: var(--wf-weight-medium);
  color: var(--wf-color-text);
  background: transparent;
  transition: background var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-accordion-item__header:hover {
  background: var(--wf-color-bg-secondary);
}

.wf-accordion-item__icon {
  color: var(--wf-color-text-secondary);
  font-size: var(--wf-text-base);
}

.wf-accordion-item__body {
  padding: 0 var(--wf-space-md) var(--wf-space-md);
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text);
}

/* ── Overlay Backdrop ── */
.wf-overlay-backdrop {
  position: fixed;
  inset: 0;
  background: var(--wf-color-bg-overlay);
  z-index: var(--wf-z-modal);
}

/* ── Dialog ── */
.wf-dialog {
  display: block;
}

.wf-dialog--open {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--wf-z-modal);
}

.wf-dialog__panel {
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-radius-lg);
  box-shadow: var(--wf-shadow-lg);
  max-width: var(--wf-dialog-max-width);
  width: 100%;
  max-height: 80vh;
  overflow-y: auto;
  font-family: var(--wf-font-sans);
  color: var(--wf-color-text);
}

.wf-dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--wf-space-md);
  border-bottom: 1px solid var(--wf-color-border);
}

.wf-dialog__title {
  font-weight: var(--wf-weight-semibold);
  font-size: var(--wf-text-base);
}

.wf-dialog__close {
  background: transparent;
  border: none;
  color: var(--wf-color-text-secondary);
  cursor: pointer;
  font-size: var(--wf-text-lg);
  line-height: 1;
  padding: var(--wf-space-xs);
}

.wf-dialog__close:hover {
  color: var(--wf-color-text);
}

.wf-dialog__body {
  padding: var(--wf-space-md);
}

/* ── Drawer ── */
.wf-drawer {
  display: block;
}

.wf-drawer--open {
  position: fixed;
  inset: 0;
  z-index: var(--wf-z-modal);
}

.wf-drawer__panel {
  position: fixed;
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  box-shadow: var(--wf-shadow-lg);
  font-family: var(--wf-font-sans);
  color: var(--wf-color-text);
  overflow-y: auto;
  transition: transform var(--wf-duration-normal) var(--wf-ease-in-out);
}

.wf-drawer__panel--right {
  top: 0;
  right: 0;
  bottom: 0;
  width: var(--wf-drawer-width);
}

.wf-drawer__panel--left {
  top: 0;
  left: 0;
  bottom: 0;
  width: var(--wf-drawer-width);
}

.wf-drawer__panel--top {
  top: 0;
  left: 0;
  right: 0;
  height: var(--wf-drawer-width);
}

.wf-drawer__panel--bottom {
  bottom: 0;
  left: 0;
  right: 0;
  height: var(--wf-drawer-width);
}

.wf-drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--wf-space-md);
  border-bottom: 1px solid var(--wf-color-border);
}

.wf-drawer__title {
  font-weight: var(--wf-weight-semibold);
  font-size: var(--wf-text-base);
}

.wf-drawer__close {
  background: transparent;
  border: none;
  color: var(--wf-color-text-secondary);
  cursor: pointer;
  font-size: var(--wf-text-lg);
  line-height: 1;
  padding: var(--wf-space-xs);
}

.wf-drawer__close:hover {
  color: var(--wf-color-text);
}

.wf-drawer__body {
  padding: var(--wf-space-md);
}

/* ── Tooltip ── */
.wf-tooltip {
  position: relative;
  display: inline-block;
}

.wf-tooltip__content {
  position: absolute;
  background: var(--wf-color-bg-elevated);
  color: var(--wf-color-text);
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-sm);
  padding: var(--wf-space-xs) var(--wf-space-sm);
  font-size: var(--wf-text-xs);
  font-family: var(--wf-font-sans);
  max-width: var(--wf-tooltip-max-width);
  white-space: nowrap;
  z-index: var(--wf-z-tooltip);
  pointer-events: none;
  transition: opacity var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-tooltip__content[aria-hidden="true"] {
  opacity: 0;
  visibility: hidden;
}

.wf-tooltip__content[aria-hidden="false"] {
  opacity: 1;
  visibility: visible;
}

.wf-tooltip__content--top {
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-bottom: var(--wf-space-xs);
}

.wf-tooltip__content--bottom {
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-top: var(--wf-space-xs);
}

.wf-tooltip__content--left {
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-right: var(--wf-space-xs);
}

.wf-tooltip__content--right {
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-left: var(--wf-space-xs);
}

/* ── Popover ── */
.wf-popover {
  position: relative;
  display: inline-block;
}

.wf-popover__trigger {
  display: inline-block;
  cursor: pointer;
}

.wf-popover__content {
  position: absolute;
  background: var(--wf-color-bg-elevated);
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-md);
  box-shadow: var(--wf-shadow-md);
  padding: var(--wf-space-md);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text);
  max-width: var(--wf-popover-max-width);
  z-index: var(--wf-z-dropdown);
}

.wf-popover__content--top {
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-bottom: var(--wf-space-xs);
}

.wf-popover__content--bottom {
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  margin-top: var(--wf-space-xs);
}

.wf-popover__content--left {
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-right: var(--wf-space-xs);
}

.wf-popover__content--right {
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-left: var(--wf-space-xs);
}

/* ── Table ── */
.wf-table {
  display: block;
  font-family: var(--wf-font-sans);
  color: var(--wf-color-text);
}

.wf-table__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--wf-text-sm);
}

.wf-table__th {
  text-align: left;
  padding: var(--wf-space-sm) var(--wf-space-md);
  font-weight: var(--wf-weight-semibold);
  font-size: var(--wf-text-xs);
  color: var(--wf-color-text-secondary);
  border-bottom: 1px solid var(--wf-color-border);
  white-space: nowrap;
  user-select: none;
}

.wf-table__th--sortable {
  cursor: pointer;
}

.wf-table__th--sortable:hover {
  color: var(--wf-color-text);
}

.wf-table__sort-icon {
  margin-left: var(--wf-space-xs);
  font-size: 0.625rem;
}

.wf-table__row {
  cursor: pointer;
  transition: background var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-table__row:hover {
  background: var(--wf-color-bg-secondary);
}

.wf-table__td {
  padding: var(--wf-space-sm) var(--wf-space-md);
  border-bottom: 1px solid var(--wf-color-border);
}

.wf-table--striped .wf-table__row:nth-child(even) {
  background: var(--wf-color-bg-secondary);
}

.wf-table__empty {
  text-align: center;
  padding: var(--wf-space-lg);
  color: var(--wf-color-text-muted);
}

.wf-table__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--wf-space-md);
  padding: var(--wf-space-md) 0;
}

.wf-table__page-btn {
  padding: var(--wf-space-xs) var(--wf-space-md);
  background: transparent;
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-md);
  color: var(--wf-color-text);
  font-family: inherit;
  font-size: var(--wf-text-sm);
  cursor: pointer;
}

.wf-table__page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.wf-table__page-info {
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text-secondary);
}
```

- [ ] **Step 2: Add layout.css import to components.css**

Add to the top of `web/packages/ui/src/styles/components.css`:

```css
@import './layout.css';
```

So the first lines of `components.css` become:

```css
/* src/styles/components.css — grows as components are added */
@import './banner.css';
@import './toast.css';
@import './forms.css';
@import './layout.css';
```

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui/src/styles/layout.css web/packages/ui/src/styles/components.css
git commit -m "feat(ui): add layout.css with styles for all Phase 3 components"
```

---

## Chunk 5: Barrel Export + Registration

### Task 12: Update index.ts

**Files:**
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Add layout imports and exports**

Add these lines to `web/packages/ui/src/index.ts`:

After the form imports:

```typescript
import './layout/wf-card.js';
import './layout/wf-tabs.js';
import './layout/wf-tab-panel.js';
import './layout/wf-accordion.js';
import './layout/wf-accordion-item.js';
import './layout/wf-dialog.js';
import './layout/wf-drawer.js';
import './layout/wf-tooltip.js';
import './layout/wf-popover.js';
import './layout/wf-table.js';
```

After the form exports:

```typescript
export { WfCard } from './layout/wf-card.js';
export { WfTabs } from './layout/wf-tabs.js';
export { WfTabPanel } from './layout/wf-tab-panel.js';
export { WfAccordion } from './layout/wf-accordion.js';
export { WfAccordionItem } from './layout/wf-accordion-item.js';
export { WfDialog } from './layout/wf-dialog.js';
export { WfDrawer } from './layout/wf-drawer.js';
export { WfTooltip } from './layout/wf-tooltip.js';
export { WfPopover } from './layout/wf-popover.js';
export { WfTable } from './layout/wf-table.js';
export type { WfTableColumn } from './layout/wf-table.js';
```

- [ ] **Step 2: Write registration test**

Add to the existing `tests/components/registration.test.ts` (or create if needed):

```typescript
// Append these test cases to the existing registration test file

import '../../src/layout/wf-card.js';
import '../../src/layout/wf-tabs.js';
import '../../src/layout/wf-tab-panel.js';
import '../../src/layout/wf-accordion.js';
import '../../src/layout/wf-accordion-item.js';
import '../../src/layout/wf-dialog.js';
import '../../src/layout/wf-drawer.js';
import '../../src/layout/wf-tooltip.js';
import '../../src/layout/wf-popover.js';
import '../../src/layout/wf-table.js';

const LAYOUT_TAGS = [
  'wf-card',
  'wf-tabs',
  'wf-tab-panel',
  'wf-accordion',
  'wf-accordion-item',
  'wf-dialog',
  'wf-drawer',
  'wf-tooltip',
  'wf-popover',
  'wf-table',
];

describe('Phase 3 component registration', () => {
  LAYOUT_TAGS.forEach((tag) => {
    it(`${tag} is registered in customElements`, () => {
      expect(customElements.get(tag)).toBeDefined();
    });
  });
});
```

- [ ] **Step 3: Verify all tests pass**

```bash
cd web/packages/ui && npx vitest run 2>&1 | tail -10
```

Expected: all existing tests plus all new layout tests pass.

- [ ] **Step 4: Commit**

```bash
git add web/packages/ui/src/index.ts web/packages/ui/tests/components/registration.test.ts
git commit -m "feat(ui): register and export all Phase 3 layout components"
```

---

## Chunk 6: React Wrappers

### Task 13: Add React wrappers for layout components

**Files:**
- Modify: the React wrapper generation file (if `@workfort/ui-react` exists)

> **Note:** If `@workfort/ui-react` does not yet exist as a package, skip this chunk. The wrapper pattern uses the existing `wrapWc` factory from the React adapter package. Each wrapper is a one-liner:

```typescript
// Example wrappers (add to @workfort/ui-react)
export const WfCard = wrapWc('wf-card', ['header', 'footer', 'variant', 'padded']);
export const WfTabs = wrapWc('wf-tabs', ['active-tab'], ['wf-tab-change']);
export const WfTabPanel = wrapWc('wf-tab-panel', ['name', 'label', 'active']);
export const WfAccordion = wrapWc('wf-accordion', ['multiple'], ['wf-accordion-change']);
export const WfAccordionItem = wrapWc('wf-accordion-item', ['name', 'header', 'expanded']);
export const WfDialog = wrapWc('wf-dialog', ['open', 'header'], ['wf-close']);
export const WfDrawer = wrapWc('wf-drawer', ['open', 'header', 'position'], ['wf-close']);
export const WfTooltip = wrapWc('wf-tooltip', ['content', 'position']);
export const WfPopover = wrapWc('wf-popover', ['open', 'position'], ['wf-close']);
export const WfTable = wrapWc('wf-table', ['striped', 'paginate', 'page-size'], ['wf-row-click', 'wf-sort', 'wf-page-change']);
```

- [ ] **Step 1: Check if @workfort/ui-react exists**

```bash
ls web/packages/ui-react/src/ 2>/dev/null || echo "SKIP: @workfort/ui-react not yet created"
```

If the package does not exist, skip to the next chunk. If it does:

- [ ] **Step 2: Add wrapper exports to the React adapter index file**

Add the wrapper lines shown above to the main export file of `@workfort/ui-react`.

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui-react/
git commit -m "feat(ui-react): add React wrappers for Phase 3 layout components"
```

---

## Chunk 7: Final Verification

### Task 14: Run full test suite

- [ ] **Step 1: Run all tests**

```bash
cd web/packages/ui && npx vitest run 2>&1 | tail -15
```

Expected: all tests pass, including:
- All existing component tests (panel, button, etc.)
- All form tests (wf-input, wf-toggle, etc.)
- All new layout tests (wf-card, wf-tabs, wf-accordion, wf-dialog, wf-drawer, wf-tooltip, wf-popover, wf-table)
- Registration tests for all Phase 3 tags

- [ ] **Step 2: Verify build**

```bash
cd web/packages/ui && npx vite build 2>&1 | tail -5
```

Expected: build succeeds with no errors.

- [ ] **Step 3: Final commit if any fixups needed**

If any test or build issues were fixed during this task:

```bash
git add -u
git commit -m "fix(ui): resolve Phase 3 test/build issues"
```

---

## Summary

| Chunk | Components | Tasks | Estimated Steps |
|-------|-----------|-------|-----------------|
| 1 | Card, Tabs, Accordion | 3 | 18 |
| 2 | Overlay utility, Dialog, Drawer, Tooltip, Popover | 5 | 22 |
| 3 | Table (basic + sorting + pagination) | 2 | 8 |
| 4 | Layout CSS | 1 | 3 |
| 5 | Barrel export + registration | 1 | 4 |
| 6 | React wrappers | 1 | 3 |
| 7 | Final verification | 1 | 3 |
| **Total** | **10 elements** | **14 tasks** | **61 steps** |

### Event Reference

| Component | Events |
|-----------|--------|
| `wf-tabs` | `wf-tab-change` |
| `wf-accordion` | `wf-accordion-change` |
| `wf-dialog` | `wf-close` |
| `wf-drawer` | `wf-close` |
| `wf-popover` | `wf-close` |
| `wf-table` | `wf-row-click`, `wf-sort`, `wf-page-change` |
