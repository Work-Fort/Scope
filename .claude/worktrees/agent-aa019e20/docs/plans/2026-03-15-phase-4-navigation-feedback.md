# Phase 4: Navigation & Feedback — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add 6 navigation and feedback components to `@workfort/ui` — Breadcrumbs, Pagination, Stepper, Progress Bar, Spinner, and Alert/Confirm Dialog — completing the toolkit for AI agents to compose full UIs.

**Architecture:** Three Lion-backed components (Pagination, Stepper, Progress Bar) extend Lion classes, override `createRenderRoot()` for light DOM, and apply WorkFort CSS classes. Two custom components (Breadcrumbs, Spinner) extend `WfElement` directly. Alert/Confirm Dialog builds on Phase 3's `WfDialog` component.

**Tech Stack:** @lion/ui 0.16.x, Lit 3, TypeScript, Vite (library mode), Vitest (happy-dom), CSS custom properties

**Spec:** `docs/ui-component-library-design.md` — Phase 4: Navigation & Feedback section

---

## Lion Integration Notes

### LionPagination (`@lion/ui/pagination.js`)

Provides: page calculation with ellipsis, prev/next/goto methods, ARIA labels (`aria-current`, `aria-label` for page buttons), localization via `LocalizeMixin`, `current-changed` event.

**Integration approach:** Extend `LionPagination`, override `createRenderRoot()` → `this`, override `render()` to produce WorkFort-styled markup. Lion's page calculation (`__calculateNavList()`) and navigation methods (`next()`, `previous()`, `goto()`, `first()`, `last()`) are inherited as-is.

### LionSteps / LionStep (`@lion/ui/steps.js`)

Provides: multi-step controller with `current` tracking, `next()`/`previous()` methods, conditional steps (`passesCondition()`), `forward-only` support, `transition` event, step lifecycle (`enter`/`leave`/`skip` events).

**Critical light DOM issue:** `LionSteps.steps` getter uses `this.shadowRoot.querySelector('slot').assignedNodes()` to find child steps. In light DOM (no shadow root), this breaks. **Solution:** Override `get steps()` to use `this.querySelectorAll('wf-step')` instead of slot assignment. Also override `render()` since the slot-based rendering is unnecessary in light DOM.

### LionProgressIndicator (`@lion/ui/progress-indicator.js`)

Provides: `role="progressbar"` management, `aria-valuenow`/`aria-valuemin`/`aria-valuemax` synchronization, `indeterminate` mode detection, `_progressPercentage` calculation, localized default label for indeterminate state.

**Integration approach:** Extend `LionProgressIndicator`, override `createRenderRoot()` → `this`, override `_graphicTemplate()` to render a WorkFort-styled progress bar with CSS-driven width. Lion handles all ARIA attribute management.

---

## Component Matrix

| Component | Tag | Extends | CSS Class | Lion? |
|-----------|-----|---------|-----------|-------|
| Breadcrumbs | `wf-breadcrumbs` | `WfElement` | `.wf-breadcrumbs` | No |
| Spinner | `wf-spinner` | `WfElement` | `.wf-spinner` | No |
| Pagination | `wf-pagination` | `LionPagination` | `.wf-pagination` | Yes |
| Stepper | `wf-stepper` | `LionSteps` | `.wf-stepper` | Yes |
| Step | `wf-step` | `LionStep` | `.wf-step` | Yes |
| Progress Bar | `wf-progress` | `LionProgressIndicator` | `.wf-progress` | Yes |
| Alert Dialog | `wf-alert-dialog` | `WfElement` | `.wf-alert-dialog` | No (uses WfDialog) |

---

## File Structure

```
web/packages/ui/
├── src/
│   ├── navigation/
│   │   ├── wf-breadcrumbs.ts
│   │   ├── wf-pagination.ts    # extends LionPagination
│   │   ├── wf-stepper.ts       # extends LionSteps
│   │   ├── wf-step.ts          # extends LionStep
│   │   ├── wf-progress.ts      # extends LionProgressIndicator
│   │   ├── wf-spinner.ts
│   │   └── wf-alert-dialog.ts  # composition with WfDialog
│   └── styles/
│       └── navigation.css      # CSS for all Phase 4 components
├── tests/
│   └── navigation/
│       ├── wf-breadcrumbs.test.ts
│       ├── wf-pagination.test.ts
│       ├── wf-stepper.test.ts
│       ├── wf-progress.test.ts
│       ├── wf-spinner.test.ts
│       └── wf-alert-dialog.test.ts
```

---

## Chunk 1: Simple Components (Breadcrumbs, Spinner)

### Task 1: Breadcrumbs component

**Files:**
- Create: `web/packages/ui/tests/navigation/wf-breadcrumbs.test.ts`
- Create: `web/packages/ui/src/navigation/wf-breadcrumbs.ts`

- [ ] **Step 1: Create test directory**

```bash
mkdir -p web/packages/ui/tests/navigation
```

- [ ] **Step 2: Write breadcrumbs tests**

```typescript
// tests/navigation/wf-breadcrumbs.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/navigation/wf-breadcrumbs.js';
import type { WfBreadcrumbs } from '../../src/navigation/wf-breadcrumbs.js';

describe('WfBreadcrumbs', () => {
  afterEach(cleanup);

  it('renders with wf-breadcrumbs class', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    expect(el.classList.contains('wf-breadcrumbs')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders nav with aria-label', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    await el.updateComplete;
    const nav = el.querySelector('nav');
    expect(nav).not.toBeNull();
    expect(nav!.getAttribute('aria-label')).toBe('Breadcrumb');
  });

  it('renders items from property', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Settings', href: '/settings' },
      { label: 'Profile' },
    ];
    await el.updateComplete;
    const links = el.querySelectorAll('.wf-breadcrumbs__link');
    expect(links.length).toBe(2);
    expect(links[0].textContent).toBe('Home');
    expect(links[0].getAttribute('href')).toBe('/');
  });

  it('marks last item as current with aria-current', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Profile' },
    ];
    await el.updateComplete;
    const current = el.querySelector('[aria-current="page"]');
    expect(current).not.toBeNull();
    expect(current!.textContent).toBe('Profile');
  });

  it('renders separator between items', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Settings', href: '/settings' },
      { label: 'Profile' },
    ];
    await el.updateComplete;
    const separators = el.querySelectorAll('.wf-breadcrumbs__separator');
    expect(separators.length).toBe(2);
  });

  it('fires wf-breadcrumb-click on link click', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [
      { label: 'Home', href: '/' },
      { label: 'Profile' },
    ];
    await el.updateComplete;

    let detail: unknown = null;
    el.addEventListener('wf-breadcrumb-click', ((e: CustomEvent) => {
      detail = e.detail;
    }) as EventListener);

    const link = el.querySelector('.wf-breadcrumbs__link') as HTMLElement;
    link.click();
    expect(detail).toEqual({ item: { label: 'Home', href: '/' }, index: 0 });
  });

  it('renders ordered list for semantic structure', async () => {
    const el = await fixture<WfBreadcrumbs>('wf-breadcrumbs');
    el.items = [{ label: 'Home', href: '/' }];
    await el.updateComplete;
    const ol = el.querySelector('ol');
    expect(ol).not.toBeNull();
    expect(ol!.classList.contains('wf-breadcrumbs__list')).toBe(true);
  });
});
```

- [ ] **Step 3: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-breadcrumbs.test.ts 2>&1 | tail -5
```

- [ ] **Step 4: Create navigation directory and implement WfBreadcrumbs**

```bash
mkdir -p web/packages/ui/src/navigation
```

```typescript
// src/navigation/wf-breadcrumbs.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';

export interface WfBreadcrumbItem {
  label: string;
  href?: string;
}

/**
 * `<wf-breadcrumbs>` — Navigation breadcrumb trail.
 * Set the `items` property to an array of { label, href? } objects.
 * The last item is rendered as the current page (no link).
 *
 * @element wf-breadcrumbs
 * @fires wf-breadcrumb-click — When a breadcrumb link is clicked. Detail: { item, index }
 */
export class WfBreadcrumbs extends WfElement {
  static get properties() {
    return {
      items: { type: Array },
    };
  }

  items: WfBreadcrumbItem[] = [];

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-breadcrumbs');
  }

  private _onItemClick(e: Event, item: WfBreadcrumbItem, index: number): void {
    e.preventDefault();
    this.dispatchEvent(
      new CustomEvent('wf-breadcrumb-click', {
        bubbles: true,
        composed: true,
        detail: { item, index },
      }),
    );
  }

  render() {
    if (!this.items.length) return nothing;

    return html`
      <nav aria-label="Breadcrumb">
        <ol class="wf-breadcrumbs__list">
          ${this.items.map((item, i) => {
            const isLast = i === this.items.length - 1;
            return html`
              <li class="wf-breadcrumbs__item">
                ${isLast
                  ? html`<span class="wf-breadcrumbs__current" aria-current="page">${item.label}</span>`
                  : html`<a
                      class="wf-breadcrumbs__link"
                      href=${item.href || '#'}
                      @click=${(e: Event) => this._onItemClick(e, item, i)}
                    >${item.label}</a>`}
                ${!isLast
                  ? html`<span class="wf-breadcrumbs__separator" aria-hidden="true">/</span>`
                  : nothing}
              </li>
            `;
          })}
        </ol>
      </nav>
    `;
  }
}

customElements.define('wf-breadcrumbs', WfBreadcrumbs);

declare global {
  interface HTMLElementTagNameMap {
    'wf-breadcrumbs': WfBreadcrumbs;
  }
}
```

- [ ] **Step 5: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-breadcrumbs.test.ts 2>&1 | tail -5
```

Expected: 8 tests pass.

- [ ] **Step 6: Commit**

```bash
git add web/packages/ui/src/navigation/wf-breadcrumbs.ts web/packages/ui/tests/navigation/wf-breadcrumbs.test.ts
git commit -m "feat(ui): add wf-breadcrumbs with semantic nav, ARIA, and click events"
```

---

### Task 2: Spinner component

**Files:**
- Create: `web/packages/ui/tests/navigation/wf-spinner.test.ts`
- Create: `web/packages/ui/src/navigation/wf-spinner.ts`

- [ ] **Step 1: Write spinner tests**

```typescript
// tests/navigation/wf-spinner.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/navigation/wf-spinner.js';
import type { WfSpinner } from '../../src/navigation/wf-spinner.js';

describe('WfSpinner', () => {
  afterEach(cleanup);

  it('renders with wf-spinner class', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    expect(el.classList.contains('wf-spinner')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    expect(el.shadowRoot).toBeNull();
  });

  it('has role="status" for accessibility', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    expect(el.getAttribute('role')).toBe('status');
  });

  it('has default aria-label', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    await el.updateComplete;
    const srText = el.querySelector('.wf-spinner__sr');
    expect(srText).not.toBeNull();
    expect(srText!.textContent).toBe('Loading');
  });

  it('uses custom label when set', async () => {
    const el = await fixture<WfSpinner>('wf-spinner', { label: 'Saving...' });
    await el.updateComplete;
    const srText = el.querySelector('.wf-spinner__sr');
    expect(srText!.textContent).toBe('Saving...');
  });

  it('applies size variant class', async () => {
    const el = await fixture<WfSpinner>('wf-spinner', { size: 'lg' });
    await el.updateComplete;
    expect(el.classList.contains('wf-spinner--lg')).toBe(true);
  });

  it('default size is md', async () => {
    const el = await fixture<WfSpinner>('wf-spinner');
    await el.updateComplete;
    expect(el.classList.contains('wf-spinner--md')).toBe(true);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-spinner.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfSpinner**

```typescript
// src/navigation/wf-spinner.ts
import { html } from 'lit';
import { WfElement } from '../base.js';

/**
 * `<wf-spinner>` — CSS-animated loading spinner.
 * Renders a circular animation with a screen-reader-accessible label.
 *
 * @element wf-spinner
 * @attr {string} size - Size variant: 'sm' | 'md' | 'lg'. Default: 'md'.
 * @attr {string} label - Screen reader text. Default: 'Loading'.
 */
export class WfSpinner extends WfElement {
  static get properties() {
    return {
      size: { type: String, reflect: true },
      label: { type: String, reflect: true },
    };
  }

  size: 'sm' | 'md' | 'lg' = 'md';
  label = 'Loading';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-spinner');
    this.setAttribute('role', 'status');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    // Update size class
    this.classList.remove('wf-spinner--sm', 'wf-spinner--md', 'wf-spinner--lg');
    this.classList.add(`wf-spinner--${this.size}`);
  }

  render() {
    return html`
      <span class="wf-spinner__circle" aria-hidden="true"></span>
      <span class="wf-spinner__sr">${this.label}</span>
    `;
  }
}

customElements.define('wf-spinner', WfSpinner);

declare global {
  interface HTMLElementTagNameMap {
    'wf-spinner': WfSpinner;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-spinner.test.ts 2>&1 | tail -5
```

Expected: 7 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/navigation/wf-spinner.ts web/packages/ui/tests/navigation/wf-spinner.test.ts
git commit -m "feat(ui): add wf-spinner with CSS animation and screen-reader support"
```

---

## Chunk 2: Lion-Backed Components (Pagination, Stepper, Progress Bar)

### Task 3: Pagination component (extends LionPagination)

**Files:**
- Create: `web/packages/ui/tests/navigation/wf-pagination.test.ts`
- Create: `web/packages/ui/src/navigation/wf-pagination.ts`

- [ ] **Step 1: Write pagination tests**

```typescript
// tests/navigation/wf-pagination.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/navigation/wf-pagination.js';
import type { WfPagination } from '../../src/navigation/wf-pagination.js';

describe('WfPagination', () => {
  afterEach(cleanup);

  it('renders with wf-pagination class', async () => {
    const el = await fixture<WfPagination>('wf-pagination');
    expect(el.classList.contains('wf-pagination')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfPagination>('wf-pagination');
    expect(el.shadowRoot).toBeNull();
  });

  it('renders nav with role="navigation"', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    await el.updateComplete;
    const nav = el.querySelector('nav');
    expect(nav).not.toBeNull();
    expect(nav!.getAttribute('role')).toBe('navigation');
  });

  it('renders page buttons matching count', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    await el.updateComplete;
    const pageButtons = el.querySelectorAll('.wf-pagination__page');
    expect(pageButtons.length).toBe(5);
  });

  it('marks current page with aria-current="true"', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 3 });
    await el.updateComplete;
    const current = el.querySelector('[aria-current="true"]');
    expect(current).not.toBeNull();
    expect(current!.textContent!.trim()).toBe('3');
  });

  it('current page button has active class', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 2 });
    await el.updateComplete;
    const active = el.querySelector('.wf-pagination__page--active');
    expect(active).not.toBeNull();
    expect(active!.textContent!.trim()).toBe('2');
  });

  it('disables previous button on first page', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    await el.updateComplete;
    const prev = el.querySelector('.wf-pagination__prev') as HTMLButtonElement;
    expect(prev.disabled).toBe(true);
  });

  it('disables next button on last page', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 5 });
    await el.updateComplete;
    const next = el.querySelector('.wf-pagination__next') as HTMLButtonElement;
    expect(next.disabled).toBe(true);
  });

  it('navigates via next()', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    await el.updateComplete;
    el.next();
    await el.updateComplete;
    expect(el.current).toBe(2);
  });

  it('navigates via previous()', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 3 });
    await el.updateComplete;
    el.previous();
    await el.updateComplete;
    expect(el.current).toBe(2);
  });

  it('navigates via goto()', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    await el.updateComplete;
    el.goto(4);
    await el.updateComplete;
    expect(el.current).toBe(4);
  });

  it('fires current-changed event on page change', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 5, current: 1 });
    await el.updateComplete;

    let fired = false;
    el.addEventListener('current-changed', () => { fired = true; });
    el.next();
    expect(fired).toBe(true);
  });

  it('renders ellipsis for large page counts', async () => {
    const el = await fixture<WfPagination>('wf-pagination', { count: 20, current: 10 });
    await el.updateComplete;
    const ellipses = el.querySelectorAll('.wf-pagination__ellipsis');
    expect(ellipses.length).toBeGreaterThan(0);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-pagination.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfPagination**

```typescript
// src/navigation/wf-pagination.ts
// @ts-expect-error — @lion/ui has no bundled type declarations
import { LionPagination } from '@lion/ui/pagination.js';
import { html } from 'lit';

/**
 * `<wf-pagination>` — Page navigation with prev/next, page numbers, and ellipsis.
 * Extends LionPagination for page calculation, ARIA labels, and localization.
 * Lion provides: __calculateNavList(), next(), previous(), goto(), first(), last(),
 * current-changed event, aria-current on active page, localized button labels.
 *
 * @element wf-pagination
 */
export class WfPagination extends LionPagination {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-pagination');
  }

  render() {
    const navList = (this as any).__calculateNavList();

    return html`
      <nav role="navigation" aria-label="Pagination" class="wf-pagination__nav">
        <ul class="wf-pagination__list">
          <li>
            <button
              class="wf-pagination__prev"
              aria-label="Previous page"
              ?disabled=${this.current <= 1}
              @click=${() => this.previous()}
            >&lsaquo;</button>
          </li>
          ${navList.map((page: number | '...') =>
            page === '...'
              ? html`<li><span class="wf-pagination__ellipsis">&hellip;</span></li>`
              : html`
                  <li>
                    <button
                      class="wf-pagination__page ${page === this.current ? 'wf-pagination__page--active' : ''}"
                      aria-current=${page === this.current}
                      aria-label="Page ${page}"
                      @click=${() => this.goto(page as number)}
                    >${page}</button>
                  </li>
                `,
          )}
          <li>
            <button
              class="wf-pagination__next"
              aria-label="Next page"
              ?disabled=${this.current >= this.count}
              @click=${() => this.next()}
            >&rsaquo;</button>
          </li>
        </ul>
      </nav>
    `;
  }
}

customElements.define('wf-pagination', WfPagination);

declare global {
  interface HTMLElementTagNameMap {
    'wf-pagination': WfPagination;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-pagination.test.ts 2>&1 | tail -5
```

Expected: 13 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/navigation/wf-pagination.ts web/packages/ui/tests/navigation/wf-pagination.test.ts
git commit -m "feat(ui): add wf-pagination extending LionPagination with WorkFort styling"
```

---

### Task 4: Stepper component (extends LionSteps/LionStep)

**Files:**
- Create: `web/packages/ui/tests/navigation/wf-stepper.test.ts`
- Create: `web/packages/ui/src/navigation/wf-stepper.ts`
- Create: `web/packages/ui/src/navigation/wf-step.ts`

- [ ] **Step 1: Write stepper tests**

```typescript
// tests/navigation/wf-stepper.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/navigation/wf-stepper.js';
import '../../src/navigation/wf-step.js';
import type { WfStepper } from '../../src/navigation/wf-stepper.js';
import type { WfStep } from '../../src/navigation/wf-step.js';

/** Helper: create a stepper with N steps, wait for render. */
async function createStepper(stepCount: number): Promise<WfStepper> {
  const el = document.createElement('wf-stepper') as WfStepper;
  for (let i = 0; i < stepCount; i++) {
    const step = document.createElement('wf-step') as WfStep;
    step.textContent = `Step ${i + 1} content`;
    el.appendChild(step);
  }
  document.body.appendChild(el);
  await el.updateComplete;
  // Wait for firstUpdated to process steps
  await new Promise(r => requestAnimationFrame(r));
  await el.updateComplete;
  return el;
}

describe('WfStepper', () => {
  afterEach(cleanup);

  it('renders with wf-stepper class', async () => {
    const el = await createStepper(3);
    expect(el.classList.contains('wf-stepper')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await createStepper(3);
    expect(el.shadowRoot).toBeNull();
  });

  it('starts at step 0', async () => {
    const el = await createStepper(3);
    expect(el.current).toBe(0);
  });

  it('first step has status "entered"', async () => {
    const el = await createStepper(3);
    const steps = el.querySelectorAll('wf-step');
    expect(steps[0].getAttribute('status')).toBe('entered');
  });

  it('navigates to next step', async () => {
    const el = await createStepper(3);
    el.next();
    await el.updateComplete;
    expect(el.current).toBe(1);
    const steps = el.querySelectorAll('wf-step');
    expect(steps[1].getAttribute('status')).toBe('entered');
  });

  it('navigates to previous step', async () => {
    const el = await createStepper(3);
    el.next();
    await el.updateComplete;
    el.previous();
    await el.updateComplete;
    expect(el.current).toBe(0);
  });

  it('fires transition event on step change', async () => {
    const el = await createStepper(3);
    let detail: unknown = null;
    el.addEventListener('transition', ((e: CustomEvent) => {
      detail = e.detail;
    }) as EventListener);
    el.next();
    expect(detail).not.toBeNull();
  });

  it('throws on out-of-bounds navigation', async () => {
    const el = await createStepper(3);
    expect(() => el.next()).not.toThrow(); // 0 -> 1
    el.current = 2;
    await el.updateComplete;
    expect(() => (el as any)._goTo(3, 2)).toThrow();
  });
});

describe('WfStep', () => {
  afterEach(cleanup);

  it('renders with wf-step class', async () => {
    const el = await fixture<WfStep>('wf-step');
    expect(el.classList.contains('wf-step')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfStep>('wf-step');
    expect(el.shadowRoot).toBeNull();
  });

  it('defaults to untouched status', async () => {
    const el = await fixture<WfStep>('wf-step');
    expect(el.status).toBe('untouched');
  });

  it('only shows content when status is entered', async () => {
    const el = await createStepper(2);
    const steps = el.querySelectorAll('wf-step');
    expect(steps[0].classList.contains('wf-step--active')).toBe(true);
    expect(steps[1].classList.contains('wf-step--active')).toBe(false);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-stepper.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfStep**

```typescript
// src/navigation/wf-step.ts
// @ts-expect-error — @lion/ui has no bundled type declarations
import { LionStep } from '@lion/ui/steps.js';
import { html } from 'lit';

/**
 * `<wf-step>` — A single step within a `<wf-stepper>`.
 * Extends LionStep for lifecycle management (enter/leave/skip events),
 * conditional display, and forward-only support.
 *
 * @element wf-step
 * @slot - Default slot for step content.
 */
export class WfStep extends LionStep {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-step');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    // Toggle visibility class based on status (replaces Lion's :host([status]) CSS)
    this.classList.toggle('wf-step--active', this.status === 'entered');
  }

  /**
   * Override firstUpdated to set controller reference without relying on
   * shadow DOM parentNode traversal.
   */
  firstUpdated(changedProperties: import('lit').PropertyValues): void {
    // Skip LionStep.firstUpdated which sets this.controller = this.parentNode
    // We do the same but explicitly, since we're in light DOM
    this.controller = this.parentElement as any;
    // Note: we intentionally skip super.firstUpdated() because LionStep's
    // implementation only sets this.controller = this.parentNode, which we
    // just did. Calling LitElement.firstUpdated directly:
    LionStep.prototype.firstUpdated.call(this, changedProperties);
  }

  render() {
    return html`<slot></slot>`;
  }
}

customElements.define('wf-step', WfStep);

declare global {
  interface HTMLElementTagNameMap {
    'wf-step': WfStep;
  }
}
```

- [ ] **Step 4: Implement WfStepper**

```typescript
// src/navigation/wf-stepper.ts
// @ts-expect-error — @lion/ui has no bundled type declarations
import { LionSteps } from '@lion/ui/steps.js';
import { html } from 'lit';

/**
 * `<wf-stepper>` — Multi-step controller.
 * Extends LionSteps for step management, conditional navigation,
 * and transition events.
 *
 * Lion provides: current tracking, next()/previous(), conditional steps
 * (passesCondition), forward-only support, transition event with
 * fromStep/toStep detail.
 *
 * @element wf-stepper
 * @fires transition — When transitioning between steps. Detail: { fromStep, toStep }
 */
export class WfStepper extends LionSteps {
  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-stepper');
  }

  /**
   * Override steps getter to work in light DOM.
   * LionSteps uses shadowRoot.querySelector('slot').assignedNodes()
   * which breaks without shadow DOM. Instead, query child wf-step elements directly.
   */
  get steps(): any[] {
    return Array.from(this.querySelectorAll(':scope > wf-step'));
  }

  render() {
    // No slot needed in light DOM — children render naturally
    return html`<slot></slot>`;
  }
}

customElements.define('wf-stepper', WfStepper);

declare global {
  interface HTMLElementTagNameMap {
    'wf-stepper': WfStepper;
  }
}
```

- [ ] **Step 5: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-stepper.test.ts 2>&1 | tail -5
```

Expected: 12 tests pass.

- [ ] **Step 6: Commit**

```bash
git add web/packages/ui/src/navigation/wf-stepper.ts web/packages/ui/src/navigation/wf-step.ts web/packages/ui/tests/navigation/wf-stepper.test.ts
git commit -m "feat(ui): add wf-stepper/wf-step extending LionSteps with light DOM support"
```

---

### Task 5: Progress Bar component (extends LionProgressIndicator)

**Files:**
- Create: `web/packages/ui/tests/navigation/wf-progress.test.ts`
- Create: `web/packages/ui/src/navigation/wf-progress.ts`

- [ ] **Step 1: Write progress bar tests**

```typescript
// tests/navigation/wf-progress.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/navigation/wf-progress.js';
import type { WfProgress } from '../../src/navigation/wf-progress.js';

describe('WfProgress', () => {
  afterEach(cleanup);

  it('renders with wf-progress class', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    expect(el.classList.contains('wf-progress')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    expect(el.shadowRoot).toBeNull();
  });

  it('has role="progressbar"', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    expect(el.getAttribute('role')).toBe('progressbar');
  });

  it('renders track and fill elements', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 50 });
    await el.updateComplete;
    expect(el.querySelector('.wf-progress__track')).not.toBeNull();
    expect(el.querySelector('.wf-progress__fill')).not.toBeNull();
  });

  it('sets fill width to match percentage', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 75, min: 0, max: 100 });
    await el.updateComplete;
    const fill = el.querySelector('.wf-progress__fill') as HTMLElement;
    expect(fill.style.width).toBe('75%');
  });

  it('sets aria-valuenow from Lion', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 30 });
    await el.updateComplete;
    expect(el.getAttribute('aria-valuenow')).toBe('30');
  });

  it('sets aria-valuemin and aria-valuemax', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 50, min: 0, max: 200 });
    await el.updateComplete;
    expect(el.getAttribute('aria-valuemin')).toBe('0');
    expect(el.getAttribute('aria-valuemax')).toBe('200');
  });

  it('clamps value to min', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: -10, min: 0, max: 100 });
    await el.updateComplete;
    expect(el.value).toBe(0);
  });

  it('clamps value to max', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 150, min: 0, max: 100 });
    await el.updateComplete;
    expect(el.value).toBe(100);
  });

  it('renders indeterminate state when no value set', async () => {
    const el = await fixture<WfProgress>('wf-progress');
    // Do not set value — Lion treats missing value attribute as indeterminate
    el.removeAttribute('value');
    await el.updateComplete;
    expect(el.indeterminate).toBe(true);
    expect(el.classList.contains('wf-progress--indeterminate')).toBe(true);
  });

  it('renders label when set', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 60 });
    el.label = '60%';
    await el.updateComplete;
    const label = el.querySelector('.wf-progress__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toBe('60%');
  });

  it('applies variant class', async () => {
    const el = await fixture<WfProgress>('wf-progress', { value: 50, variant: 'success' });
    await el.updateComplete;
    expect(el.classList.contains('wf-progress--success')).toBe(true);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-progress.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfProgress**

```typescript
// src/navigation/wf-progress.ts
// @ts-expect-error — @lion/ui has no bundled type declarations
import { LionProgressIndicator } from '@lion/ui/progress-indicator.js';
import { html, nothing } from 'lit';

/**
 * `<wf-progress>` — Determinate or indeterminate progress bar.
 * Extends LionProgressIndicator for ARIA attribute management
 * (role="progressbar", aria-valuenow/min/max), indeterminate detection,
 * and percentage calculation.
 *
 * @element wf-progress
 * @attr {number} value - Current progress value. Omit for indeterminate.
 * @attr {number} min - Minimum value. Default: 0.
 * @attr {number} max - Maximum value. Default: 100.
 * @attr {string} variant - Visual variant: 'default' | 'success' | 'warning' | 'error'.
 */
export class WfProgress extends LionProgressIndicator {
  static get properties() {
    return {
      ...super.properties,
      label: { type: String },
      variant: { type: String, reflect: true },
    };
  }

  label = '';
  variant = '';

  createRenderRoot(): this {
    return this;
  }

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-progress');
  }

  updated(changedProperties: Map<string, unknown>): void {
    super.updated(changedProperties);
    // Toggle indeterminate class
    this.classList.toggle('wf-progress--indeterminate', this.indeterminate);
    // Toggle variant class
    this.classList.remove('wf-progress--success', 'wf-progress--warning', 'wf-progress--error');
    if (this.variant) {
      this.classList.add(`wf-progress--${this.variant}`);
    }
  }

  /**
   * Override Lion's _graphicTemplate to render WorkFort-styled progress bar.
   * Lion calls this from its render() method.
   */
  _graphicTemplate() {
    const percentage = this._progressPercentage;
    return html`
      <div class="wf-progress__track">
        <div
          class="wf-progress__fill"
          style=${this.indeterminate ? '' : `width: ${percentage}%`}
        ></div>
      </div>
      ${this.label
        ? html`<span class="wf-progress__label">${this.label}</span>`
        : nothing}
    `;
  }
}

customElements.define('wf-progress', WfProgress);

declare global {
  interface HTMLElementTagNameMap {
    'wf-progress': WfProgress;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-progress.test.ts 2>&1 | tail -5
```

Expected: 12 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/navigation/wf-progress.ts web/packages/ui/tests/navigation/wf-progress.test.ts
git commit -m "feat(ui): add wf-progress extending LionProgressIndicator with determinate/indeterminate modes"
```

---

## Chunk 3: Alert/Confirm Dialog

### Task 6: Alert/Confirm Dialog component

**Files:**
- Create: `web/packages/ui/tests/navigation/wf-alert-dialog.test.ts`
- Create: `web/packages/ui/src/navigation/wf-alert-dialog.ts`

> **Dependency note:** This component builds on Phase 3's `WfDialog` via composition. It creates a `wf-dialog` internally with confirm/cancel button slots. If Phase 3's Dialog is later refactored to use Lion overlays, this component should be updated to match — but the public API will remain the same.

- [ ] **Step 1: Write alert dialog tests**

```typescript
// tests/navigation/wf-alert-dialog.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/layout/wf-dialog.js';
import '../../src/navigation/wf-alert-dialog.js';
import type { WfAlertDialog } from '../../src/navigation/wf-alert-dialog.js';

describe('WfAlertDialog', () => {
  afterEach(cleanup);

  it('renders with wf-alert-dialog class', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    expect(el.classList.contains('wf-alert-dialog')).toBe(true);
  });

  it('has no shadow DOM', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    expect(el.shadowRoot).toBeNull();
  });

  it('is closed by default', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    expect(el.open).toBe(false);
  });

  it('renders dialog panel when shown', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { header: 'Confirm' });
    el.message = 'Are you sure?';
    el.show();
    await el.updateComplete;
    expect(el.open).toBe(true);
    const panel = el.querySelector('.wf-alert-dialog__panel');
    expect(panel).not.toBeNull();
  });

  it('renders message text', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    el.message = 'Delete this item?';
    el.show();
    await el.updateComplete;
    const msg = el.querySelector('.wf-alert-dialog__message');
    expect(msg).not.toBeNull();
    expect(msg!.textContent).toBe('Delete this item?');
  });

  it('renders confirm button with default text', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    el.show();
    await el.updateComplete;
    const confirm = el.querySelector('.wf-alert-dialog__confirm') as HTMLButtonElement;
    expect(confirm).not.toBeNull();
    expect(confirm.textContent!.trim()).toBe('OK');
  });

  it('renders cancel button for confirm variant', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { type: 'confirm' });
    el.show();
    await el.updateComplete;
    const cancel = el.querySelector('.wf-alert-dialog__cancel') as HTMLButtonElement;
    expect(cancel).not.toBeNull();
    expect(cancel.textContent!.trim()).toBe('Cancel');
  });

  it('does not render cancel button for alert variant', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { type: 'alert' });
    el.show();
    await el.updateComplete;
    const cancel = el.querySelector('.wf-alert-dialog__cancel');
    expect(cancel).toBeNull();
  });

  it('fires wf-confirm on confirm click', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    el.show();
    await el.updateComplete;

    let fired = false;
    el.addEventListener('wf-confirm', () => { fired = true; });
    const confirm = el.querySelector('.wf-alert-dialog__confirm') as HTMLElement;
    confirm.click();
    expect(fired).toBe(true);
  });

  it('fires wf-cancel on cancel click', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { type: 'confirm' });
    el.show();
    await el.updateComplete;

    let fired = false;
    el.addEventListener('wf-cancel', () => { fired = true; });
    const cancel = el.querySelector('.wf-alert-dialog__cancel') as HTMLElement;
    cancel.click();
    expect(fired).toBe(true);
  });

  it('closes after confirm click', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    el.show();
    await el.updateComplete;
    const confirm = el.querySelector('.wf-alert-dialog__confirm') as HTMLElement;
    confirm.click();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('closes after cancel click', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { type: 'confirm' });
    el.show();
    await el.updateComplete;
    const cancel = el.querySelector('.wf-alert-dialog__cancel') as HTMLElement;
    cancel.click();
    await el.updateComplete;
    expect(el.open).toBe(false);
  });

  it('uses custom button labels', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { type: 'confirm' });
    el.confirmLabel = 'Delete';
    el.cancelLabel = 'Keep';
    el.show();
    await el.updateComplete;
    const confirm = el.querySelector('.wf-alert-dialog__confirm') as HTMLElement;
    const cancel = el.querySelector('.wf-alert-dialog__cancel') as HTMLElement;
    expect(confirm.textContent!.trim()).toBe('Delete');
    expect(cancel.textContent!.trim()).toBe('Keep');
  });

  it('renders header when set', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { header: 'Warning' });
    el.show();
    await el.updateComplete;
    const header = el.querySelector('.wf-alert-dialog__header');
    expect(header).not.toBeNull();
    expect(header!.textContent!.trim()).toBe('Warning');
  });

  it('sets role="alertdialog"', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    el.show();
    await el.updateComplete;
    const panel = el.querySelector('[role="alertdialog"]');
    expect(panel).not.toBeNull();
  });

  it('applies variant class for destructive actions', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { variant: 'destructive' });
    el.show();
    await el.updateComplete;
    expect(el.classList.contains('wf-alert-dialog--destructive')).toBe(true);
  });

  it('returns promise from show() that resolves on confirm', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog');
    const promise = el.show();
    await el.updateComplete;
    const confirm = el.querySelector('.wf-alert-dialog__confirm') as HTMLElement;
    confirm.click();
    const result = await promise;
    expect(result).toBe(true);
  });

  it('returns promise from show() that resolves false on cancel', async () => {
    const el = await fixture<WfAlertDialog>('wf-alert-dialog', { type: 'confirm' });
    const promise = el.show();
    await el.updateComplete;
    const cancel = el.querySelector('.wf-alert-dialog__cancel') as HTMLElement;
    cancel.click();
    const result = await promise;
    expect(result).toBe(false);
  });
});
```

- [ ] **Step 2: Verify tests fail**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-alert-dialog.test.ts 2>&1 | tail -5
```

- [ ] **Step 3: Implement WfAlertDialog**

```typescript
// src/navigation/wf-alert-dialog.ts
import { html, nothing } from 'lit';
import { WfElement } from '../base.js';
import { trapFocus, createBackdrop, removeBackdrop, onEscape } from '../utils/overlay.js';

/**
 * `<wf-alert-dialog>` — Modal alert or confirmation dialog.
 * Builds on the overlay utilities from Phase 3 (same as WfDialog).
 *
 * Two modes:
 * - `type="alert"` — single OK button (default)
 * - `type="confirm"` — OK + Cancel buttons
 *
 * The `show()` method returns a Promise<boolean> that resolves to:
 * - `true` when the user clicks confirm
 * - `false` when the user clicks cancel or presses Escape
 *
 * @element wf-alert-dialog
 * @fires wf-confirm — When the confirm button is clicked.
 * @fires wf-cancel — When the cancel button is clicked.
 * @fires wf-close — When the dialog is closed (either confirm or cancel).
 */
export class WfAlertDialog extends WfElement {
  static get properties() {
    return {
      open: { type: Boolean, reflect: true },
      type: { type: String, reflect: true },
      header: { type: String, reflect: true },
      message: { type: String },
      confirmLabel: { type: String, attribute: 'confirm-label' },
      cancelLabel: { type: String, attribute: 'cancel-label' },
      variant: { type: String, reflect: true },
    };
  }

  open = false;
  type: 'alert' | 'confirm' = 'alert';
  header = '';
  message = '';
  confirmLabel = 'OK';
  cancelLabel = 'Cancel';
  variant = '';

  private _backdrop: HTMLDivElement | null = null;
  private _cleanupFocus: (() => void) | null = null;
  private _cleanupEscape: (() => void) | null = null;
  private _previousFocus: HTMLElement | null = null;
  private _resolve: ((value: boolean) => void) | null = null;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-alert-dialog');
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this._teardown();
  }

  /**
   * Show the dialog. Returns a promise that resolves to true (confirm) or false (cancel).
   */
  show(): Promise<boolean> {
    this._previousFocus = document.activeElement as HTMLElement | null;
    this.open = true;

    if (this.variant) {
      this.classList.add(`wf-alert-dialog--${this.variant}`);
    }

    this.requestUpdate();

    this._backdrop = createBackdrop(() => this._cancel());
    this._cleanupEscape = onEscape(this, () => this._cancel());

    // Defer focus trap to after render
    requestAnimationFrame(() => {
      const panel = this.querySelector('.wf-alert-dialog__panel') as HTMLElement;
      if (panel) {
        this._cleanupFocus = trapFocus(panel);
        // Focus confirm button by default
        const confirm = this.querySelector('.wf-alert-dialog__confirm') as HTMLElement;
        if (confirm) confirm.focus();
      }
    });

    return new Promise<boolean>((resolve) => {
      this._resolve = resolve;
    });
  }

  hide(): void {
    this.open = false;
    this._teardown();
    this.requestUpdate();

    this.dispatchEvent(
      new CustomEvent('wf-close', { bubbles: true, composed: true }),
    );

    if (this._previousFocus) {
      this._previousFocus.focus();
      this._previousFocus = null;
    }
  }

  private _confirm(): void {
    this.dispatchEvent(
      new CustomEvent('wf-confirm', { bubbles: true, composed: true }),
    );
    if (this._resolve) {
      this._resolve(true);
      this._resolve = null;
    }
    this.hide();
  }

  private _cancel(): void {
    this.dispatchEvent(
      new CustomEvent('wf-cancel', { bubbles: true, composed: true }),
    );
    if (this._resolve) {
      this._resolve(false);
      this._resolve = null;
    }
    this.hide();
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
        class="wf-alert-dialog__panel"
        role="alertdialog"
        aria-modal="true"
        aria-label=${this.header || 'Alert'}
      >
        ${this.header
          ? html`<div class="wf-alert-dialog__header">${this.header}</div>`
          : nothing}
        ${this.message
          ? html`<div class="wf-alert-dialog__message">${this.message}</div>`
          : nothing}
        <div class="wf-alert-dialog__actions">
          ${this.type === 'confirm'
            ? html`
                <button
                  class="wf-alert-dialog__cancel"
                  @click=${() => this._cancel()}
                >${this.cancelLabel}</button>
              `
            : nothing}
          <button
            class="wf-alert-dialog__confirm"
            @click=${() => this._confirm()}
          >${this.confirmLabel}</button>
        </div>
      </div>
    `;
  }
}

customElements.define('wf-alert-dialog', WfAlertDialog);

declare global {
  interface HTMLElementTagNameMap {
    'wf-alert-dialog': WfAlertDialog;
  }
}
```

- [ ] **Step 4: Verify tests pass**

```bash
cd web/packages/ui && npx vitest run tests/navigation/wf-alert-dialog.test.ts 2>&1 | tail -5
```

Expected: 17 tests pass.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/navigation/wf-alert-dialog.ts web/packages/ui/tests/navigation/wf-alert-dialog.test.ts
git commit -m "feat(ui): add wf-alert-dialog with alert/confirm modes and promise-based API"
```

---

## Chunk 4: CSS + Exports + Registration Tests

### Task 7: Navigation CSS

**Files:**
- Create: `web/packages/ui/src/styles/navigation.css`
- Modify: `web/packages/ui/src/styles/components.css`

- [ ] **Step 1: Create navigation.css**

```css
/* src/styles/navigation.css — Phase 4: Navigation & Feedback */

/* ── Breadcrumbs ── */
.wf-breadcrumbs {
  display: block;
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text);
}

.wf-breadcrumbs__list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--wf-space-xs);
}

.wf-breadcrumbs__item {
  display: flex;
  align-items: center;
  gap: var(--wf-space-xs);
}

.wf-breadcrumbs__link {
  color: var(--wf-color-text-secondary);
  text-decoration: none;
  transition: color var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-breadcrumbs__link:hover {
  color: var(--wf-color-text);
  text-decoration: underline;
}

.wf-breadcrumbs__separator {
  color: var(--wf-color-text-muted);
  user-select: none;
}

.wf-breadcrumbs__current {
  color: var(--wf-color-text);
  font-weight: var(--wf-weight-medium);
}

/* ── Spinner ── */
.wf-spinner {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.wf-spinner__circle {
  display: block;
  border-radius: var(--wf-radius-full);
  border: 2px solid var(--wf-color-border);
  border-top-color: var(--wf-color-text);
  animation: wf-spin 0.6s linear infinite;
}

.wf-spinner--sm .wf-spinner__circle {
  width: 1rem;
  height: 1rem;
}

.wf-spinner--md .wf-spinner__circle {
  width: 1.5rem;
  height: 1.5rem;
}

.wf-spinner--lg .wf-spinner__circle {
  width: 2.5rem;
  height: 2.5rem;
  border-width: 3px;
}

.wf-spinner__sr {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@keyframes wf-spin {
  to {
    transform: rotate(360deg);
  }
}

/* ── Pagination ── */
.wf-pagination {
  display: block;
  font-family: var(--wf-font-sans);
}

.wf-pagination__nav {
  display: flex;
  justify-content: center;
}

.wf-pagination__list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  align-items: center;
  gap: var(--wf-space-xs);
}

.wf-pagination__page,
.wf-pagination__prev,
.wf-pagination__next {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  height: 2rem;
  padding: 0 var(--wf-space-xs);
  border: 1px solid var(--wf-color-border-strong);
  border-radius: var(--wf-radius-md);
  background: transparent;
  color: var(--wf-color-text);
  font-family: inherit;
  font-size: var(--wf-text-sm);
  cursor: pointer;
  transition: background var(--wf-duration-fast) var(--wf-ease-in-out),
              color var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-pagination__page:hover,
.wf-pagination__prev:hover,
.wf-pagination__next:hover {
  background: var(--wf-color-bg-secondary);
}

.wf-pagination__page--active {
  background: var(--wf-color-text);
  color: var(--wf-color-bg);
  border-color: var(--wf-color-text);
  font-weight: var(--wf-weight-semibold);
}

.wf-pagination__page--active:hover {
  background: var(--wf-color-text);
  color: var(--wf-color-bg);
}

.wf-pagination__prev:disabled,
.wf-pagination__next:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.wf-pagination__ellipsis {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  height: 2rem;
  color: var(--wf-color-text-muted);
  font-size: var(--wf-text-sm);
}

/* ── Stepper ── */
.wf-stepper {
  display: block;
}

.wf-step {
  display: none;
}

.wf-step--active {
  display: block;
}

/* ── Progress Bar ── */
.wf-progress {
  display: block;
  font-family: var(--wf-font-sans);
}

.wf-progress__track {
  width: 100%;
  height: 0.5rem;
  background: var(--wf-color-bg-secondary);
  border-radius: var(--wf-radius-full);
  overflow: hidden;
}

.wf-progress__fill {
  height: 100%;
  background: var(--wf-color-text);
  border-radius: var(--wf-radius-full);
  transition: width var(--wf-duration-normal) var(--wf-ease-in-out);
}

.wf-progress--success .wf-progress__fill {
  background: var(--wf-color-success);
}

.wf-progress--warning .wf-progress__fill {
  background: var(--wf-color-warning, var(--wf-amber-500));
}

.wf-progress--error .wf-progress__fill {
  background: var(--wf-color-error);
}

.wf-progress--indeterminate .wf-progress__fill {
  width: 30% !important;
  animation: wf-progress-indeterminate 1.5s var(--wf-ease-in-out) infinite;
}

@keyframes wf-progress-indeterminate {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(400%);
  }
}

.wf-progress__label {
  display: block;
  margin-top: var(--wf-space-xs);
  font-size: var(--wf-text-xs);
  color: var(--wf-color-text-secondary);
}

/* ── Alert Dialog ── */
.wf-alert-dialog {
  display: block;
}

.wf-alert-dialog__panel {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-radius-lg);
  box-shadow: var(--wf-shadow-lg);
  padding: var(--wf-space-lg);
  min-width: 20rem;
  max-width: 28rem;
  z-index: var(--wf-z-modal);
  font-family: var(--wf-font-sans);
}

.wf-alert-dialog__header {
  font-size: var(--wf-text-base);
  font-weight: var(--wf-weight-semibold);
  color: var(--wf-color-text);
  margin-bottom: var(--wf-space-sm);
}

.wf-alert-dialog__message {
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text-secondary);
  line-height: var(--wf-leading-normal);
  margin-bottom: var(--wf-space-lg);
}

.wf-alert-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--wf-space-sm);
}

.wf-alert-dialog__confirm,
.wf-alert-dialog__cancel {
  padding: var(--wf-space-xs) var(--wf-space-md);
  border-radius: var(--wf-radius-md);
  font-family: inherit;
  font-size: var(--wf-text-sm);
  font-weight: var(--wf-weight-medium);
  cursor: pointer;
  transition: background var(--wf-duration-fast) var(--wf-ease-in-out);
}

.wf-alert-dialog__confirm {
  background: var(--wf-color-text);
  color: var(--wf-color-bg);
  border: 1px solid var(--wf-color-text);
}

.wf-alert-dialog__confirm:hover {
  opacity: 0.9;
}

.wf-alert-dialog__cancel {
  background: transparent;
  color: var(--wf-color-text);
  border: 1px solid var(--wf-color-border-strong);
}

.wf-alert-dialog__cancel:hover {
  background: var(--wf-color-bg-secondary);
}

.wf-alert-dialog--destructive .wf-alert-dialog__confirm {
  background: var(--wf-color-error);
  border-color: var(--wf-color-error);
}
```

- [ ] **Step 2: Add navigation.css import to components.css**

Add to the top of `web/packages/ui/src/styles/components.css`:

```css
@import './navigation.css';
```

So the import lines in `components.css` become:

```css
@import './banner.css';
@import './toast.css';
@import './forms.css';
@import './layout.css';
@import './navigation.css';
```

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui/src/styles/navigation.css web/packages/ui/src/styles/components.css
git commit -m "feat(ui): add navigation.css with styles for all Phase 4 components"
```

---

### Task 8: Update index.ts

**Files:**
- Modify: `web/packages/ui/src/index.ts`

- [ ] **Step 1: Add navigation imports and exports**

Add these lines to `web/packages/ui/src/index.ts`:

After the layout imports (or form imports, if layout is not yet added):

```typescript
import './navigation/wf-breadcrumbs.js';
import './navigation/wf-pagination.js';
import './navigation/wf-stepper.js';
import './navigation/wf-step.js';
import './navigation/wf-progress.js';
import './navigation/wf-spinner.js';
import './navigation/wf-alert-dialog.js';
```

After the layout exports (or form exports):

```typescript
export { WfBreadcrumbs } from './navigation/wf-breadcrumbs.js';
export type { WfBreadcrumbItem } from './navigation/wf-breadcrumbs.js';
export { WfPagination } from './navigation/wf-pagination.js';
export { WfStepper } from './navigation/wf-stepper.js';
export { WfStep } from './navigation/wf-step.js';
export { WfProgress } from './navigation/wf-progress.js';
export { WfSpinner } from './navigation/wf-spinner.js';
export { WfAlertDialog } from './navigation/wf-alert-dialog.js';
```

- [ ] **Step 2: Write registration test**

Add to `tests/components/registration.test.ts`:

```typescript
// Append to the existing EXPECTED array:
const NAVIGATION_TAGS = [
  'wf-breadcrumbs',
  'wf-pagination',
  'wf-stepper',
  'wf-step',
  'wf-progress',
  'wf-spinner',
  'wf-alert-dialog',
];

describe('Phase 4 component registration', () => {
  NAVIGATION_TAGS.forEach((tag) => {
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

Expected: all existing tests plus all new navigation tests pass.

- [ ] **Step 4: Commit**

```bash
git add web/packages/ui/src/index.ts web/packages/ui/tests/components/registration.test.ts
git commit -m "feat(ui): register and export all Phase 4 navigation components"
```

---

## Chunk 5: React Wrappers + Framework Adapter Updates

### Task 9: Add React wrappers for navigation components

**Files:**
- Modify: `web/packages/react/src/components.tsx`
- Modify: `web/packages/react/src/index.tsx`

- [ ] **Step 1: Add type imports to components.tsx**

Add to the type import block in `web/packages/react/src/components.tsx`:

```typescript
import type {
  // ... existing imports ...
  WfBreadcrumbs, WfPagination, WfStepper, WfStep,
  WfProgress, WfSpinner, WfAlertDialog,
} from '@workfort/ui';
```

- [ ] **Step 2: Add wrapper exports to components.tsx**

Add after the existing form component wrappers:

```typescript
// Navigation & Feedback components
export const Breadcrumbs = wrapWc<WfBreadcrumbs, { items?: unknown[] }>('wf-breadcrumbs', 'Breadcrumbs');
export const Pagination = wrapWc<WfPagination, { count?: number; current?: number }>('wf-pagination', 'Pagination');
export const Stepper = wrapWc<WfStepper, { current?: number }>('wf-stepper', 'Stepper');
export const Step = wrapWc<WfStep, { status?: string; 'forward-only'?: boolean; 'initial-step'?: boolean }>('wf-step', 'Step');
export const Progress = wrapWc<WfProgress, { value?: number; min?: number; max?: number; variant?: string; label?: string }>('wf-progress', 'Progress');
export const Spinner = wrapWc<WfSpinner, { size?: 'sm' | 'md' | 'lg'; label?: string }>('wf-spinner', 'Spinner');
export const AlertDialog = wrapWc<WfAlertDialog, { type?: 'alert' | 'confirm'; header?: string; message?: string; 'confirm-label'?: string; 'cancel-label'?: string; variant?: string }>('wf-alert-dialog', 'AlertDialog');
```

- [ ] **Step 3: Update index.tsx**

Add navigation components to the re-exports in `web/packages/react/src/index.tsx`:

```typescript
export {
  Panel, Button, Badge, StatusDot, Skeleton, Divider,
  TextInput, List, ListItem, ScrollArea, ErrorFallback,
  // Form components
  Input, Textarea, Select, Checkbox, CheckboxGroup,
  Radio, RadioGroup, Toggle, Slider, Combobox,
  DatePicker, FileUpload, Form,
  // Navigation & Feedback components
  Breadcrumbs, Pagination, Stepper, Step,
  Progress, Spinner, AlertDialog,
} from './components.js';
export { useAuth } from './use-auth.js';
export { useTheme } from './use-theme.js';
```

> **Note:** `@workfort/ui-solid`, `@workfort/ui-vue`, `@workfort/ui-svelte` do not need component wrappers — Solid, Vue, and Svelte handle custom elements natively. No changes required for those packages.

- [ ] **Step 4: Commit**

```bash
git add web/packages/react/src/components.tsx web/packages/react/src/index.tsx
git commit -m "feat(ui-react): add React wrappers for Phase 4 navigation components"
```

---

## Chunk 6: Storybook Stories (for Documentation repo)

### Task 10: Story specifications

> **Note:** Stories live in the `Work-Fort/Documentation` repository, not in the main monorepo. This task documents the story specs for implementation in that repo.

Each component gets one story file per renderer (HTML, React, Solid where applicable).

- [ ] **Step 1: Document story specs**

The following stories should be created in the Documentation repo:

**`wf-breadcrumbs`:**
- Default with 3 items
- Single item (no links)
- Long trail (5+ items)
- Interactive: click handler logs navigation

**`wf-spinner`:**
- Sizes: sm, md, lg
- Custom label
- Inline with text

**`wf-pagination`:**
- Small page count (5 pages)
- Large page count (50 pages, showing ellipsis)
- Interactive: current page changes on click

**`wf-stepper` + `wf-step`:**
- 3-step wizard with content
- Navigation via next/previous buttons
- Conditional step (skip based on data)

**`wf-progress`:**
- Determinate: 0%, 25%, 50%, 75%, 100%
- Indeterminate
- Variants: success, warning, error
- With label

**`wf-alert-dialog`:**
- Alert mode (single OK button)
- Confirm mode (OK + Cancel)
- Destructive variant
- Custom button labels
- Promise-based usage example

- [ ] **Step 2: No commit needed** — stories are implemented in the Documentation repo.

---

## Chunk 7: Final Verification

### Task 11: Run full test suite

- [ ] **Step 1: Run all tests**

```bash
cd web/packages/ui && npx vitest run 2>&1 | tail -15
```

Expected: all tests pass, including:
- All existing component tests (panel, button, etc.)
- All form tests (wf-input, wf-toggle, etc.)
- All layout tests (wf-card, wf-tabs, etc., if Phase 3 is implemented)
- All new navigation tests (wf-breadcrumbs, wf-pagination, wf-stepper, wf-progress, wf-spinner, wf-alert-dialog)
- Registration tests for all Phase 4 tags

- [ ] **Step 2: Verify build**

```bash
cd web/packages/ui && npx vite build 2>&1 | tail -5
```

Expected: build succeeds with no errors.

- [ ] **Step 3: Final commit if any fixups needed**

If any test or build issues were fixed during this task:

```bash
git add -u
git commit -m "fix(ui): resolve Phase 4 test/build issues"
```

---

## Summary

| Chunk | Components | Tasks | Estimated Steps |
|-------|-----------|-------|-----------------|
| 1 | Breadcrumbs, Spinner | 2 | 11 |
| 2 | Pagination (Lion), Stepper (Lion), Progress Bar (Lion) | 3 | 15 |
| 3 | Alert/Confirm Dialog | 1 | 5 |
| 4 | CSS, exports, registration | 2 | 7 |
| 5 | React wrappers | 1 | 4 |
| 6 | Storybook story specs | 1 | 2 |
| 7 | Final verification | 1 | 3 |
| **Total** | **7 elements** | **11 tasks** | **47 steps** |

### Event Reference

| Component | Events |
|-----------|--------|
| `wf-breadcrumbs` | `wf-breadcrumb-click` |
| `wf-pagination` | `current-changed` (inherited from Lion) |
| `wf-stepper` | `transition` (inherited from Lion) |
| `wf-step` | `enter`, `leave`, `skip` (inherited from Lion) |
| `wf-alert-dialog` | `wf-confirm`, `wf-cancel`, `wf-close` |

### Lion Integration Summary

| Component | Lion Class | What Lion Provides | What We Override |
|-----------|-----------|-------------------|-----------------|
| `wf-pagination` | `LionPagination` | Page calculation, ellipsis logic, nav methods, ARIA labels, localization | `createRenderRoot()`, `render()` |
| `wf-stepper` | `LionSteps` | Step controller, `current` tracking, conditional navigation, transition events | `createRenderRoot()`, `render()`, `get steps()` (light DOM fix) |
| `wf-step` | `LionStep` | Step lifecycle (enter/leave/skip), status management, condition evaluation | `createRenderRoot()`, `render()`, `firstUpdated()`, `updated()` |
| `wf-progress` | `LionProgressIndicator` | `role="progressbar"`, `aria-value*` management, indeterminate detection, percentage calc | `createRenderRoot()`, `_graphicTemplate()` |
