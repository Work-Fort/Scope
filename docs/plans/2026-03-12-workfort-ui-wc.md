# @workfort/ui Web Components Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `@workfort/ui`, a Lit-based Web Components library with a framework-agnostic auth client and adapters for React, Vue, Svelte, and SolidJS.

**Architecture:** Single npm package with sub-path exports. Core components are Lit Custom Elements rendered in light DOM, themed via `--wf-*` CSS custom properties. AuthClient manages session state with zero framework dependencies. React gets wrapper components (needed for proper WC interop); Vue, Svelte, and SolidJS get type augmentations + framework-native state hooks only (they handle Custom Elements natively).

**Tech Stack:** Lit 3, TypeScript 5, Vite (library mode), Vitest (happy-dom), React 18+, Vue 3+, Svelte 5+, SolidJS 1.8+

**Spec:** `docs/2026-03-12-workfort-ui-wc-design.md`

---

## File Structure

All source lives at `web/ui/` in the workfort repo:

```
web/ui/
├── package.json              # name: @workfort/ui
├── tsconfig.json
├── tsconfig.build.json       # Excludes tests
├── vite.config.ts            # Library mode, 6 entry points
├── vitest.workspace.ts       # Test projects (core, react, vue, svelte, solid)
├── src/
│   ├── index.ts              # Registers all custom elements, re-exports classes
│   ├── base.ts               # WfElement base class (light DOM override)
│   ├── components/
│   │   ├── panel.ts
│   │   ├── button.ts
│   │   ├── badge.ts
│   │   ├── status-dot.ts
│   │   ├── skeleton.ts
│   │   ├── divider.ts
│   │   ├── text-input.ts
│   │   ├── list.ts
│   │   ├── list-item.ts
│   │   ├── scroll-area.ts
│   │   └── error-fallback.ts
│   ├── styles/
│   │   ├── tokens.css        # --wf-* custom property fallback values
│   │   └── components.css    # All component structural styles
│   ├── auth/
│   │   ├── index.ts          # Re-exports AuthClient, types, getAuthClient
│   │   ├── client.ts         # AuthClient implementation
│   │   └── types.ts          # User, Session, AuthInitError, AuthEvents
│   ├── react/
│   │   ├── index.tsx         # Re-exports wrapper components + hooks
│   │   ├── components.tsx    # React wrapper components (forwardRef)
│   │   ├── use-auth.ts       # useAuth() hook
│   │   └── use-theme.ts     # useTheme() hook
│   ├── vue/
│   │   ├── index.ts          # Re-exports composables + type augmentations
│   │   ├── use-auth.ts       # useAuth() composable
│   │   └── use-theme.ts     # useTheme() composable
│   ├── svelte/
│   │   ├── index.ts          # Re-exports stores + type augmentations
│   │   ├── auth.ts           # Auth Svelte store
│   │   └── theme.ts          # Theme Svelte store
│   └── solid/
│       ├── index.ts          # Re-exports primitives + type augmentations
│       ├── use-auth.ts       # useAuth() signal-based primitive
│       └── use-theme.ts     # useTheme() signal-based primitive
└── tests/
    ├── helpers.ts            # Shared test utilities (fixture, cleanup)
    ├── components/           # One test file per component
    ├── auth/
    │   ├── client.test.ts
    │   └── singleton.test.ts
    ├── react/
    │   ├── components.test.tsx
    │   └── use-auth.test.tsx
    ├── vue/
    │   └── use-auth.test.ts
    ├── svelte/
    │   └── auth.test.ts
    └── solid/
        └── use-auth.test.ts
```

**Key decisions:**
- **React is the only adapter with wrapper components** — React's Custom Element support is poor (event handling, boolean attributes). Vue, Svelte, and SolidJS handle Custom Elements natively and only need type augmentations + state hooks.
- **Vitest workspace** — React adapter tests need `@vitejs/plugin-react` for JSX transform. Other tests don't. Workspace config isolates this.
- **No Svelte compiler in build** — Svelte adapter exports plain `.ts` files (stores + type augmentations), not `.svelte` components. Consumers use `<wf-panel>` directly in Svelte templates.

---

## Chunk 1: Project Scaffold + Core Components

### Task 1: Initialize project

**Files:**
- Create: `web/ui/package.json`
- Create: `web/ui/tsconfig.json`
- Create: `web/ui/tsconfig.build.json`
- Create: `web/ui/vite.config.ts`
- Create: `web/ui/vitest.workspace.ts`
- Modify: `.gitignore`

- [ ] **Step 1: Create directory and package.json**

```bash
mkdir -p web/ui
```

```json
{
  "name": "@workfort/ui",
  "version": "0.0.1",
  "type": "module",
  "license": "Apache-2.0",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "default": "./dist/index.js"
    },
    "./auth": {
      "types": "./dist/auth/index.d.ts",
      "default": "./dist/auth/index.js"
    },
    "./react": {
      "types": "./dist/react/index.d.ts",
      "default": "./dist/react/index.js"
    },
    "./vue": {
      "types": "./dist/vue/index.d.ts",
      "default": "./dist/vue/index.js"
    },
    "./svelte": {
      "types": "./dist/svelte/index.d.ts",
      "default": "./dist/svelte/index.js"
    },
    "./solid": {
      "types": "./dist/solid/index.d.ts",
      "default": "./dist/solid/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "vite build",
    "test": "vitest run",
    "test:watch": "vitest",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "lit": "^3.2.0"
  },
  "peerDependencies": {
    "react": "^18.0.0 || ^19.0.0",
    "vue": "^3.3.0",
    "svelte": "^4.0.0 || ^5.0.0",
    "solid-js": "^1.8.0"
  },
  "peerDependenciesMeta": {
    "react": { "optional": true },
    "vue": { "optional": true },
    "svelte": { "optional": true },
    "solid-js": { "optional": true }
  },
  "devDependencies": {
    "@testing-library/react": "^16.0.0",
    "@types/react": "^18.3.0",
    "@vitejs/plugin-react": "^4.3.0",
    "happy-dom": "^15.0.0",
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "solid-js": "^1.9.0",
    "svelte": "^5.0.0",
    "typescript": "^5.6.0",
    "vite": "^6.0.0",
    "vite-plugin-dts": "^4.0.0",
    "vitest": "^2.1.0",
    "vue": "^3.5.0"
  }
}
```

- [ ] **Step 2: Create tsconfig.json**

`experimentalDecorators` + `useDefineForClassFields: false` required for Lit's `@property()` decorator. `jsx: react-jsx` enables React JSX for `.tsx` files only.

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "experimentalDecorators": true,
    "useDefineForClassFields": false,
    "jsx": "react-jsx",
    "paths": {
      "@workfort/ui": ["./src/index.ts"],
      "@workfort/ui/auth": ["./src/auth/index.ts"],
      "@workfort/ui/react": ["./src/react/index.tsx"],
      "@workfort/ui/vue": ["./src/vue/index.ts"],
      "@workfort/ui/svelte": ["./src/svelte/index.ts"],
      "@workfort/ui/solid": ["./src/solid/index.ts"]
    }
  },
  "include": ["src", "tests"]
}
```

- [ ] **Step 3: Create tsconfig.build.json**

```json
{
  "extends": "./tsconfig.json",
  "exclude": ["tests"]
}
```

- [ ] **Step 4: Create vite.config.ts**

`react({ include: /\.tsx$/ })` scopes the React JSX transform to `.tsx` files only. `cssCodeSplit: false` bundles all CSS into a single `style.css`.

```ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import dts from 'vite-plugin-dts';

export default defineConfig({
  plugins: [
    react({ include: /\.tsx$/ }),
    dts({ tsconfigPath: './tsconfig.build.json' }),
  ],
  build: {
    lib: {
      entry: {
        index: 'src/index.ts',
        'auth/index': 'src/auth/index.ts',
        'react/index': 'src/react/index.tsx',
        'vue/index': 'src/vue/index.ts',
        'svelte/index': 'src/svelte/index.ts',
        'solid/index': 'src/solid/index.ts',
      },
      formats: ['es'],
    },
    rollupOptions: {
      external: [
        'lit',
        /^lit\//,
        'react',
        'react/jsx-runtime',
        'react-dom',
        'vue',
        /^svelte/,
        'solid-js',
        /^solid-js\//,
      ],
    },
    cssCodeSplit: false,
  },
});
```

- [ ] **Step 5: Create vitest.workspace.ts**

```ts
import { defineWorkspace } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineWorkspace([
  {
    test: {
      name: 'core',
      include: ['tests/components/**/*.test.ts', 'tests/auth/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
  {
    test: {
      name: 'react',
      include: ['tests/react/**/*.test.tsx'],
      environment: 'happy-dom',
    },
    plugins: [react()],
  },
  {
    test: {
      name: 'vue',
      include: ['tests/vue/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
  {
    test: {
      name: 'svelte',
      include: ['tests/svelte/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
  {
    test: {
      name: 'solid',
      include: ['tests/solid/**/*.test.ts'],
      environment: 'happy-dom',
    },
  },
]);
```

- [ ] **Step 6: Update root .gitignore**

Add these lines to the project's `.gitignore`:

```
web/ui/node_modules/
web/ui/dist/
```

- [ ] **Step 7: Install dependencies and verify**

```bash
cd web/ui && npm install
```

- [ ] **Step 8: Commit**

```bash
git add web/ui/package.json web/ui/package-lock.json web/ui/tsconfig.json web/ui/tsconfig.build.json web/ui/vite.config.ts web/ui/vitest.workspace.ts .gitignore
git commit -m "feat(ui): initialize @workfort/ui package scaffold"
```

Do NOT commit `node_modules/`. DO commit `package-lock.json` (lockfile ensures reproducible installs).

---

### Task 2: Test helpers and base class

**Files:**
- Create: `web/ui/tests/helpers.ts`
- Create: `web/ui/src/base.ts`

- [ ] **Step 1: Create test helpers**

```ts
// tests/helpers.ts

/**
 * Create and connect a Custom Element for testing.
 * Waits for Lit's updateComplete before returning.
 */
export async function fixture<T extends HTMLElement>(
  tag: string,
  attrs?: Record<string, string | number | boolean>,
): Promise<T> {
  const el = document.createElement(tag) as T;
  if (attrs) {
    for (const [key, val] of Object.entries(attrs)) {
      if (typeof val === 'boolean') {
        if (val) el.setAttribute(key, '');
      } else {
        el.setAttribute(key, String(val));
      }
    }
  }
  document.body.appendChild(el);
  if ('updateComplete' in el) {
    await (el as any).updateComplete;
  }
  return el;
}

/** Remove all children from document.body between tests. */
export function cleanup(): void {
  while (document.body.firstChild) {
    document.body.removeChild(document.body.firstChild);
  }
}
```

- [ ] **Step 2: Create WfElement base class**

```ts
// src/base.ts
import { LitElement } from 'lit';

/**
 * Base class for all @workfort/ui components.
 * Renders in light DOM (no Shadow DOM).
 */
export class WfElement extends LitElement {
  createRenderRoot(): this {
    return this;
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add web/ui/tests/helpers.ts web/ui/src/base.ts
git commit -m "feat(ui): add test helpers and WfElement base class"
```

---

### Task 3: CSS tokens and component styles

**Files:**
- Create: `web/ui/src/styles/tokens.css`
- Create: `web/ui/src/styles/components.css`

- [ ] **Step 1: Create CSS tokens with fallback values**

```css
/* src/styles/tokens.css */
:root {
  --wf-bg: #1a1a2e;
  --wf-bg-secondary: #16213e;
  --wf-text: #e0e0e0;
  --wf-text-secondary: #a0a0a0;
  --wf-text-muted: #606060;
  --wf-border: #2a2a4a;
  --wf-accent: #5865f2;
  --wf-space-xs: 4px;
  --wf-space-sm: 8px;
  --wf-space-md: 12px;
  --wf-space-lg: 16px;
  --wf-space-xl: 24px;
  --wf-font-sans: system-ui, -apple-system, sans-serif;
  --wf-font-mono: 'SF Mono', 'Fira Code', monospace;
  --wf-font-size-xs: 0.75rem;
  --wf-font-size-sm: 0.875rem;
  --wf-font-size-base: 1rem;
  --wf-font-size-lg: 1.125rem;
  --wf-radius-sm: 4px;
  --wf-radius-md: 6px;
  --wf-radius-lg: 8px;
}
```

- [ ] **Step 2: Create component structural styles**

See `src/styles/components.css` — all component CSS using `--wf-*` tokens. Each component has a `.wf-<name>` base class and optional modifier classes. Full CSS provided in the component tasks below; create the file empty here and add styles as each component is implemented.

```css
/* src/styles/components.css — grows as components are added */
```

- [ ] **Step 3: Commit**

```bash
git add web/ui/src/styles/
git commit -m "feat(ui): add CSS tokens and component styles scaffold"
```

---

### Task 4: WfPanel component (exemplar — full TDD)

This is the first component. It establishes the pattern all others follow. Pay close attention to how Lit light DOM rendering interacts with consumer-provided children.

**Files:**
- Create: `web/ui/src/components/panel.ts`
- Create: `web/ui/tests/components/panel.test.ts`
- Modify: `web/ui/src/styles/components.css`

**Ref:** Spec lines 81, 93-109, 119

- [ ] **Step 1: Write the failing test**

```ts
// tests/components/panel.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/panel.js';
import type { WfPanel } from '../../src/components/panel.js';

describe('WfPanel', () => {
  afterEach(cleanup);

  it('is registered as a custom element', () => {
    expect(customElements.get('wf-panel')).toBeDefined();
  });

  it('renders with wf-panel class', async () => {
    const el = await fixture<WfPanel>('wf-panel');
    expect(el.classList.contains('wf-panel')).toBe(true);
  });

  it('renders label when provided', async () => {
    const el = await fixture<WfPanel>('wf-panel', { label: 'Channels' });
    const label = el.querySelector('.wf-panel__label');
    expect(label).not.toBeNull();
    expect(label!.textContent).toBe('Channels');
  });

  it('does not render label when empty', async () => {
    const el = await fixture<WfPanel>('wf-panel');
    const label = el.querySelector('.wf-panel__label');
    expect(label).toBeNull();
  });

  it('preserves consumer-provided children', async () => {
    const el = await fixture<WfPanel>('wf-panel');
    const child = document.createElement('div');
    child.className = 'user-content';
    child.textContent = 'Hello';
    el.appendChild(child);
    await el.updateComplete;
    const found = el.querySelector('.user-content');
    expect(found).not.toBeNull();
    expect(found!.textContent).toBe('Hello');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd web/ui && npx vitest run tests/components/panel.test.ts
```

Expected: FAIL — module not found.

- [ ] **Step 3: Implement WfPanel**

```ts
// src/components/panel.ts
import { html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfPanel extends WfElement {
  @property({ type: String }) label = '';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-panel');
  }

  render() {
    return this.label
      ? html`<div class="wf-panel__label">${this.label}</div>`
      : nothing;
  }
}

customElements.define('wf-panel', WfPanel);

declare global {
  interface HTMLElementTagNameMap {
    'wf-panel': WfPanel;
  }
}
```

**Critical:** Lit light DOM rendering inserts template content via comment markers — it does NOT replace existing children. The "preserves consumer-provided children" test verifies this. If it fails with happy-dom, switch to `@web/test-runner` with a real browser for component tests.

- [ ] **Step 4: Add panel styles to components.css**

```css
/* Append to src/styles/components.css */
.wf-panel {
  display: block;
  background: var(--wf-bg);
  border: 1px solid var(--wf-border);
  border-radius: var(--wf-radius-md);
  padding: var(--wf-space-md);
  font-family: var(--wf-font-sans);
  color: var(--wf-text);
}
.wf-panel__label {
  font-size: var(--wf-font-size-sm);
  color: var(--wf-text-secondary);
  margin-bottom: var(--wf-space-sm);
  font-weight: 600;
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd web/ui && npx vitest run tests/components/panel.test.ts
```

Expected: 5 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add web/ui/src/components/panel.ts web/ui/tests/components/panel.test.ts web/ui/src/styles/components.css
git commit -m "feat(ui): add WfPanel component with tests"
```

---

### Task 5: WfButton component

**Files:**
- Create: `web/ui/src/components/button.ts`
- Create: `web/ui/tests/components/button.test.ts`
- Modify: `web/ui/src/styles/components.css`

- [ ] **Step 1: Write the failing test**

```ts
// tests/components/button.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/button.js';
import type { WfButton } from '../../src/components/button.js';

describe('WfButton', () => {
  afterEach(cleanup);

  it('renders with wf-button class', async () => {
    const el = await fixture<WfButton>('wf-button');
    expect(el.classList.contains('wf-button')).toBe(true);
  });

  it('applies filled variant class', async () => {
    const el = await fixture<WfButton>('wf-button', { variant: 'filled' });
    expect(el.classList.contains('wf-button--filled')).toBe(true);
  });

  it('dispatches wf-click event on click', async () => {
    const el = await fixture<WfButton>('wf-button');
    const handler = vi.fn();
    el.addEventListener('wf-click', handler);
    el.click();
    expect(handler).toHaveBeenCalledOnce();
  });

  it('does not dispatch wf-click when disabled', async () => {
    const el = await fixture<WfButton>('wf-button', { disabled: true });
    const handler = vi.fn();
    el.addEventListener('wf-click', handler);
    el.click();
    expect(handler).not.toHaveBeenCalled();
  });

  it('has button role and tabindex', async () => {
    const el = await fixture<WfButton>('wf-button');
    expect(el.getAttribute('role')).toBe('button');
    expect(el.getAttribute('tabindex')).toBe('0');
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd web/ui && npx vitest run tests/components/button.test.ts
```

- [ ] **Step 3: Implement WfButton**

```ts
// src/components/button.ts
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfButton extends WfElement {
  @property({ type: String }) variant: 'text' | 'filled' = 'text';
  @property({ type: Boolean, reflect: true }) disabled = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-button');
    this.setAttribute('role', 'button');
    this.setAttribute('tabindex', '0');
    this.addEventListener('click', this._handleClick);
    this.addEventListener('keydown', this._handleKeydown);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('click', this._handleClick);
    this.removeEventListener('keydown', this._handleKeydown);
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('variant')) {
      this.classList.toggle('wf-button--filled', this.variant === 'filled');
    }
    if (changed.has('disabled')) {
      this.setAttribute('aria-disabled', String(this.disabled));
      this.setAttribute('tabindex', this.disabled ? '-1' : '0');
    }
  }

  private _handleClick = (e: Event): void => {
    if (this.disabled) {
      e.stopImmediatePropagation();
      return;
    }
    this.dispatchEvent(new CustomEvent('wf-click', { bubbles: true, composed: true }));
  };

  private _handleKeydown = (e: KeyboardEvent): void => {
    if ((e.key === 'Enter' || e.key === ' ') && !this.disabled) {
      e.preventDefault();
      this.dispatchEvent(new CustomEvent('wf-click', { bubbles: true, composed: true }));
    }
  };
}

customElements.define('wf-button', WfButton);

declare global {
  interface HTMLElementTagNameMap {
    'wf-button': WfButton;
  }
}
```

- [ ] **Step 4: Add button styles to components.css**

```css
/* Append to src/styles/components.css */
.wf-button {
  display: inline-flex;
  align-items: center;
  gap: var(--wf-space-xs);
  padding: var(--wf-space-sm) var(--wf-space-md);
  border: 1px solid var(--wf-border);
  border-radius: var(--wf-radius-sm);
  background: transparent;
  color: var(--wf-text);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-font-size-sm);
  cursor: pointer;
}
.wf-button[disabled] { opacity: 0.5; cursor: not-allowed; pointer-events: none; }
.wf-button--filled { background: var(--wf-accent); border-color: var(--wf-accent); color: white; }
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd web/ui && npx vitest run tests/components/button.test.ts
```

- [ ] **Step 6: Commit**

```bash
git add web/ui/src/components/button.ts web/ui/tests/components/button.test.ts web/ui/src/styles/components.css
git commit -m "feat(ui): add WfButton component with tests"
```

---

### Task 6: Leaf components (Badge, StatusDot, Skeleton, Divider)

Four simple components with no children or complex interactions. Batched.

**Files:**
- Create: `web/ui/src/components/badge.ts`, `status-dot.ts`, `skeleton.ts`, `divider.ts`
- Create: `web/ui/tests/components/leaf.test.ts`
- Modify: `web/ui/src/styles/components.css`

- [ ] **Step 1: Write the failing tests**

```ts
// tests/components/leaf.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/badge.js';
import '../../src/components/status-dot.js';
import '../../src/components/skeleton.js';
import '../../src/components/divider.js';
import type { WfBadge } from '../../src/components/badge.js';
import type { WfStatusDot } from '../../src/components/status-dot.js';
import type { WfSkeleton } from '../../src/components/skeleton.js';

describe('WfBadge', () => {
  afterEach(cleanup);

  it('renders count', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 5 });
    expect(el.textContent).toContain('5');
  });

  it('hides when count is 0', async () => {
    const el = await fixture<WfBadge>('wf-badge', { count: 0 });
    expect(el.style.display).toBe('none');
  });
});

describe('WfStatusDot', () => {
  afterEach(cleanup);

  it('applies status class', async () => {
    const el = await fixture<WfStatusDot>('wf-status-dot', { status: 'online' });
    expect(el.classList.contains('wf-status-dot--online')).toBe(true);
  });

  it('defaults to offline', async () => {
    const el = await fixture<WfStatusDot>('wf-status-dot');
    expect(el.classList.contains('wf-status-dot--offline')).toBe(true);
  });
});

describe('WfSkeleton', () => {
  afterEach(cleanup);

  it('applies dimensions from attributes', async () => {
    const el = await fixture<WfSkeleton>('wf-skeleton', { width: '100px', height: '20px' });
    expect(el.style.width).toBe('100px');
    expect(el.style.height).toBe('20px');
  });
});

describe('WfDivider', () => {
  afterEach(cleanup);

  it('renders with wf-divider class', async () => {
    const el = await fixture('wf-divider');
    expect(el.classList.contains('wf-divider')).toBe(true);
  });

  it('has separator role', async () => {
    const el = await fixture('wf-divider');
    expect(el.getAttribute('role')).toBe('separator');
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web/ui && npx vitest run tests/components/leaf.test.ts
```

- [ ] **Step 3: Implement all four components**

```ts
// src/components/badge.ts
import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfBadge extends WfElement {
  @property({ type: Number }) count = 0;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-badge');
  }

  updated(): void {
    this.style.display = this.count > 0 ? '' : 'none';
  }

  render() {
    return html`${this.count}`;
  }
}

customElements.define('wf-badge', WfBadge);
declare global { interface HTMLElementTagNameMap { 'wf-badge': WfBadge; } }
```

```ts
// src/components/status-dot.ts
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfStatusDot extends WfElement {
  @property({ type: String }) status: 'online' | 'offline' | 'away' = 'offline';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-status-dot');
    this._applyStatus();
  }

  updated(): void { this._applyStatus(); }

  private _applyStatus(): void {
    this.classList.remove('wf-status-dot--online', 'wf-status-dot--away', 'wf-status-dot--offline');
    this.classList.add(`wf-status-dot--${this.status}`);
    this.setAttribute('aria-label', this.status);
  }
}

customElements.define('wf-status-dot', WfStatusDot);
declare global { interface HTMLElementTagNameMap { 'wf-status-dot': WfStatusDot; } }
```

```ts
// src/components/skeleton.ts
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfSkeleton extends WfElement {
  @property({ type: String }) width = '100%';
  @property({ type: String }) height = '1em';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-skeleton');
    this.setAttribute('aria-hidden', 'true');
    this._applyDimensions();
  }

  updated(): void { this._applyDimensions(); }

  private _applyDimensions(): void {
    this.style.width = this.width;
    this.style.height = this.height;
  }
}

customElements.define('wf-skeleton', WfSkeleton);
declare global { interface HTMLElementTagNameMap { 'wf-skeleton': WfSkeleton; } }
```

```ts
// src/components/divider.ts
import { WfElement } from '../base.js';

export class WfDivider extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-divider');
    this.setAttribute('role', 'separator');
  }
}

customElements.define('wf-divider', WfDivider);
declare global { interface HTMLElementTagNameMap { 'wf-divider': WfDivider; } }
```

- [ ] **Step 4: Add styles to components.css**

Append badge, status-dot, skeleton, and divider styles:

```css
/* Badge */
.wf-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.25em;
  height: 1.25em;
  padding: 0 var(--wf-space-xs, 0.25rem);
  border-radius: var(--wf-radius-lg, 9999px);
  background: var(--wf-accent, #6366f1);
  color: #fff;
  font-size: var(--wf-font-size-xs, 0.75rem);
  font-family: var(--wf-font-sans, system-ui);
  line-height: 1;
}
.wf-badge:empty { display: none; }

/* StatusDot */
.wf-status-dot {
  display: inline-block;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: var(--wf-text-muted, #6b7280);
}
.wf-status-dot--online { background: #22c55e; }
.wf-status-dot--away { background: #eab308; }
.wf-status-dot--offline { background: var(--wf-text-muted, #6b7280); }

/* Skeleton */
.wf-skeleton {
  display: block;
  background: var(--wf-bg-secondary, #f3f4f6);
  border-radius: var(--wf-radius-sm, 0.25rem);
  animation: wf-skeleton-pulse 1.5s ease-in-out infinite;
}
@keyframes wf-skeleton-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* Divider */
.wf-divider {
  display: block;
  height: 1px;
  background: var(--wf-border, #e5e7eb);
  margin: var(--wf-space-sm, 0.5rem) 0;
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd web/ui && npx vitest run tests/components/leaf.test.ts
```

- [ ] **Step 6: Commit**

```bash
git add web/ui/src/components/badge.ts web/ui/src/components/status-dot.ts web/ui/src/components/skeleton.ts web/ui/src/components/divider.ts web/ui/tests/components/leaf.test.ts web/ui/src/styles/components.css
git commit -m "feat(ui): add Badge, StatusDot, Skeleton, Divider components"
```

---

### Task 7: WfTextInput component

**Files:**
- Create: `web/ui/src/components/text-input.ts`
- Create: `web/ui/tests/components/text-input.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
// tests/components/text-input.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/text-input.js';
import type { WfTextInput } from '../../src/components/text-input.js';

describe('WfTextInput', () => {
  afterEach(cleanup);

  it('renders an input element', async () => {
    const el = await fixture<WfTextInput>('wf-text-input');
    expect(el.querySelector('input')).not.toBeNull();
  });

  it('sets placeholder', async () => {
    const el = await fixture<WfTextInput>('wf-text-input', { placeholder: 'Type here...' });
    expect(el.querySelector('input')!.placeholder).toBe('Type here...');
  });

  it('dispatches wf-input on input', async () => {
    const el = await fixture<WfTextInput>('wf-text-input');
    const handler = vi.fn();
    el.addEventListener('wf-input', handler);
    const input = el.querySelector('input')!;
    input.value = 'hello';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    expect(handler).toHaveBeenCalledOnce();
    expect((handler.mock.calls[0][0] as CustomEvent).detail.value).toBe('hello');
  });

  it('reflects value property', async () => {
    const el = await fixture<WfTextInput>('wf-text-input');
    el.value = 'test';
    await el.updateComplete;
    expect(el.querySelector('input')!.value).toBe('test');
  });
});
```

- [ ] **Step 2: Run test, verify it fails**

```bash
cd web/ui && npx vitest run tests/components/text-input.test.ts
```

- [ ] **Step 3: Implement WfTextInput**

```ts
// src/components/text-input.ts
import { html } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfTextInput extends WfElement {
  @property({ type: String }) placeholder = '';
  @property({ type: String }) value = '';
  @property({ type: Boolean, reflect: true }) disabled = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-text-input');
  }

  render() {
    return html`
      <input
        class="wf-text-input__input"
        .value=${this.value}
        placeholder=${this.placeholder}
        ?disabled=${this.disabled}
        @input=${this._onInput}
        @change=${this._onChange}
      />
    `;
  }

  private _onInput(e: Event): void {
    const input = e.target as HTMLInputElement;
    this.value = input.value;
    this.dispatchEvent(new CustomEvent('wf-input', {
      bubbles: true, composed: true, detail: { value: input.value },
    }));
  }

  private _onChange(e: Event): void {
    const input = e.target as HTMLInputElement;
    this.dispatchEvent(new CustomEvent('wf-change', {
      bubbles: true, composed: true, detail: { value: input.value },
    }));
  }
}

customElements.define('wf-text-input', WfTextInput);
declare global { interface HTMLElementTagNameMap { 'wf-text-input': WfTextInput; } }
```

- [ ] **Step 4: Add text-input styles to components.css**

```css
/* TextInput */
.wf-text-input { display: block; }
.wf-text-input__input {
  width: 100%;
  padding: var(--wf-space-sm, 0.5rem) var(--wf-space-md, 0.75rem);
  font-family: var(--wf-font-sans, system-ui);
  font-size: var(--wf-font-size-base, 0.875rem);
  color: var(--wf-text, #111827);
  background: var(--wf-bg, #ffffff);
  border: 1px solid var(--wf-border, #e5e7eb);
  border-radius: var(--wf-radius-md, 0.375rem);
  outline: none;
  box-sizing: border-box;
}
.wf-text-input__input:focus {
  border-color: var(--wf-accent, #6366f1);
}
.wf-text-input[disabled] .wf-text-input__input {
  opacity: 0.5;
  cursor: not-allowed;
}
```

- [ ] **Step 5: Run test, verify it passes**

```bash
cd web/ui && npx vitest run tests/components/text-input.test.ts
```

- [ ] **Step 6: Commit**

```bash
git add web/ui/src/components/text-input.ts web/ui/tests/components/text-input.test.ts web/ui/src/styles/components.css
git commit -m "feat(ui): add WfTextInput component with tests"
```

---

### Task 8: WfList and WfListItem components

**Files:**
- Create: `web/ui/src/components/list.ts`, `list-item.ts`
- Create: `web/ui/tests/components/list.test.ts`

**Ref:** Spec line 119 — `data-wf` attribute convention for named content areas.

- [ ] **Step 1: Write the failing test**

```ts
// tests/components/list.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/list.js';
import '../../src/components/list-item.js';
import type { WfListItem } from '../../src/components/list-item.js';

describe('WfList', () => {
  afterEach(cleanup);

  it('renders with wf-list class and list role', async () => {
    const el = await fixture('wf-list');
    expect(el.classList.contains('wf-list')).toBe(true);
    expect(el.getAttribute('role')).toBe('list');
  });
});

describe('WfListItem', () => {
  afterEach(cleanup);

  it('renders with wf-list-item class and listitem role', async () => {
    const el = await fixture<WfListItem>('wf-list-item');
    expect(el.classList.contains('wf-list-item')).toBe(true);
    expect(el.getAttribute('role')).toBe('listitem');
  });

  it('applies active class', async () => {
    const el = await fixture<WfListItem>('wf-list-item', { active: true });
    expect(el.classList.contains('wf-list-item--active')).toBe(true);
  });

  it('wraps trailing content in trailing container', async () => {
    const el = await fixture<WfListItem>('wf-list-item');
    const trailing = document.createElement('span');
    trailing.setAttribute('data-wf', 'trailing');
    trailing.textContent = '3';
    el.appendChild(trailing);
    await el.updateComplete;
    const container = el.querySelector('.wf-list-item__trailing');
    expect(container).not.toBeNull();
    expect(container!.contains(trailing)).toBe(true);
  });

  it('dispatches wf-select on click', async () => {
    const el = await fixture<WfListItem>('wf-list-item');
    const handler = vi.fn();
    el.addEventListener('wf-select', handler);
    el.click();
    expect(handler).toHaveBeenCalledOnce();
  });
});
```

- [ ] **Step 2: Run test, verify it fails**

```bash
cd web/ui && npx vitest run tests/components/list.test.ts
```

- [ ] **Step 3: Implement WfList and WfListItem**

```ts
// src/components/list.ts
import { WfElement } from '../base.js';

export class WfList extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-list');
    this.setAttribute('role', 'list');
  }
}

customElements.define('wf-list', WfList);
declare global { interface HTMLElementTagNameMap { 'wf-list': WfList; } }
```

```ts
// src/components/list-item.ts
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfListItem extends WfElement {
  @property({ type: Boolean, reflect: true }) active = false;

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-list-item');
    this.setAttribute('role', 'listitem');
    this.addEventListener('click', this._handleClick);
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    this.removeEventListener('click', this._handleClick);
  }

  updated(changed: Map<string, unknown>): void {
    if (changed.has('active')) {
      this.classList.toggle('wf-list-item--active', this.active);
    }
    this._wrapTrailingContent();
  }

  /** Finds [data-wf="trailing"] children and wraps in .wf-list-item__trailing. */
  private _wrapTrailingContent(): void {
    const trailing = this.querySelectorAll('[data-wf="trailing"]');
    if (trailing.length === 0) return;
    let container = this.querySelector('.wf-list-item__trailing');
    if (!container) {
      container = document.createElement('span');
      container.className = 'wf-list-item__trailing';
      this.appendChild(container);
    }
    trailing.forEach((el) => {
      if (el.parentElement !== container) container!.appendChild(el);
    });
  }

  private _handleClick = (): void => {
    this.dispatchEvent(new CustomEvent('wf-select', { bubbles: true, composed: true }));
  };
}

customElements.define('wf-list-item', WfListItem);
declare global { interface HTMLElementTagNameMap { 'wf-list-item': WfListItem; } }
```

- [ ] **Step 4: Add list styles to components.css**

```css
/* List */
.wf-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 0;
  margin: 0;
  list-style: none;
}

/* ListItem */
.wf-list-item {
  display: flex;
  align-items: center;
  padding: var(--wf-space-sm, 0.5rem) var(--wf-space-md, 0.75rem);
  cursor: pointer;
  color: var(--wf-text, #111827);
  border-radius: var(--wf-radius-sm, 0.25rem);
}
.wf-list-item:hover { background: var(--wf-bg-secondary, #f3f4f6); }
.wf-list-item--active { background: var(--wf-bg-secondary, #f3f4f6); }
.wf-list-item__trailing {
  margin-left: auto;
  display: flex;
  align-items: center;
}
```

- [ ] **Step 5: Run test, verify it passes**

```bash
cd web/ui && npx vitest run tests/components/list.test.ts
```

- [ ] **Step 6: Commit**

```bash
git add web/ui/src/components/list.ts web/ui/src/components/list-item.ts web/ui/tests/components/list.test.ts web/ui/src/styles/components.css
git commit -m "feat(ui): add WfList and WfListItem components with tests"
```

---

### Task 9: WfScrollArea, WfErrorFallback, and core index

**Files:**
- Create: `web/ui/src/components/scroll-area.ts`, `error-fallback.ts`
- Create: `web/ui/tests/components/scroll-area.test.ts`, `error-fallback.test.ts`
- Create: `web/ui/src/index.ts`
- Create: `web/ui/tests/components/registration.test.ts`

- [ ] **Step 1: Write tests for ScrollArea and ErrorFallback**

```ts
// tests/components/scroll-area.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/scroll-area.js';

describe('WfScrollArea', () => {
  afterEach(cleanup);
  it('renders with wf-scroll-area class', async () => {
    const el = await fixture('wf-scroll-area');
    expect(el.classList.contains('wf-scroll-area')).toBe(true);
  });
});
```

```ts
// tests/components/error-fallback.test.ts
import { describe, it, expect, afterEach } from 'vitest';
import { fixture, cleanup } from '../helpers.js';
import '../../src/components/error-fallback.js';
import type { WfErrorFallback } from '../../src/components/error-fallback.js';

describe('WfErrorFallback', () => {
  afterEach(cleanup);

  it('renders title and message', async () => {
    const el = await fixture<WfErrorFallback>('wf-error-fallback', {
      title: 'Oops', message: 'Something went wrong',
    });
    expect(el.querySelector('.wf-error-fallback__title')!.textContent).toBe('Oops');
    expect(el.querySelector('.wf-error-fallback__message')!.textContent).toBe('Something went wrong');
  });

  it('has alert role', async () => {
    const el = await fixture('wf-error-fallback');
    expect(el.getAttribute('role')).toBe('alert');
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd web/ui && npx vitest run tests/components/scroll-area.test.ts tests/components/error-fallback.test.ts
```

- [ ] **Step 3: Implement ScrollArea and ErrorFallback**

```ts
// src/components/scroll-area.ts
import { WfElement } from '../base.js';

export class WfScrollArea extends WfElement {
  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-scroll-area');
  }
}

customElements.define('wf-scroll-area', WfScrollArea);
declare global { interface HTMLElementTagNameMap { 'wf-scroll-area': WfScrollArea; } }
```

```ts
// src/components/error-fallback.ts
import { html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { WfElement } from '../base.js';

export class WfErrorFallback extends WfElement {
  @property({ type: String }) title = '';
  @property({ type: String }) message = '';

  connectedCallback(): void {
    super.connectedCallback();
    this.classList.add('wf-error-fallback');
    this.setAttribute('role', 'alert');
  }

  render() {
    return html`
      ${this.title ? html`<div class="wf-error-fallback__title">${this.title}</div>` : nothing}
      ${this.message ? html`<div class="wf-error-fallback__message">${this.message}</div>` : nothing}
    `;
  }
}

customElements.define('wf-error-fallback', WfErrorFallback);
declare global { interface HTMLElementTagNameMap { 'wf-error-fallback': WfErrorFallback; } }
```

- [ ] **Step 4: Add scroll-area and error-fallback styles to components.css**

```css
/* ScrollArea */
.wf-scroll-area {
  display: block;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--wf-border, #e5e7eb) transparent;
}
.wf-scroll-area::-webkit-scrollbar { width: 6px; }
.wf-scroll-area::-webkit-scrollbar-track { background: transparent; }
.wf-scroll-area::-webkit-scrollbar-thumb {
  background: var(--wf-border, #e5e7eb);
  border-radius: 3px;
}

/* ErrorFallback */
.wf-error-fallback {
  display: block;
  padding: var(--wf-space-md, 0.75rem);
  color: var(--wf-text, #111827);
}
.wf-error-fallback__title {
  font-weight: 600;
  margin-bottom: var(--wf-space-xs, 0.25rem);
}
.wf-error-fallback__message {
  color: var(--wf-text-secondary, #4b5563);
  font-size: var(--wf-font-size-sm, 0.8125rem);
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd web/ui && npx vitest run tests/components/scroll-area.test.ts tests/components/error-fallback.test.ts
```

- [ ] **Step 6: Create core index.ts**

```ts
// src/index.ts
import './styles/tokens.css';
import './styles/components.css';

import './components/panel.js';
import './components/button.js';
import './components/badge.js';
import './components/status-dot.js';
import './components/skeleton.js';
import './components/divider.js';
import './components/text-input.js';
import './components/list.js';
import './components/list-item.js';
import './components/scroll-area.js';
import './components/error-fallback.js';

export { WfPanel } from './components/panel.js';
export { WfButton } from './components/button.js';
export { WfBadge } from './components/badge.js';
export { WfStatusDot } from './components/status-dot.js';
export { WfSkeleton } from './components/skeleton.js';
export { WfDivider } from './components/divider.js';
export { WfTextInput } from './components/text-input.js';
export { WfList } from './components/list.js';
export { WfListItem } from './components/list-item.js';
export { WfScrollArea } from './components/scroll-area.js';
export { WfErrorFallback } from './components/error-fallback.js';
export { WfElement } from './base.js';
```

- [ ] **Step 7: Write registration test**

```ts
// tests/components/registration.test.ts
import { describe, it, expect } from 'vitest';
import '../../src/index.js';

const EXPECTED = [
  'wf-panel', 'wf-button', 'wf-badge', 'wf-status-dot', 'wf-skeleton',
  'wf-divider', 'wf-text-input', 'wf-list', 'wf-list-item',
  'wf-scroll-area', 'wf-error-fallback',
];

describe('@workfort/ui registration', () => {
  it('registers all custom elements', () => {
    for (const tag of EXPECTED) {
      expect(customElements.get(tag), `${tag} should be registered`).toBeDefined();
    }
  });
});
```

- [ ] **Step 8: Run ALL component tests**

```bash
cd web/ui && npx vitest run tests/components/
```

Expected: All tests pass.

- [ ] **Step 9: Commit**

```bash
git add web/ui/src/components/scroll-area.ts web/ui/src/components/error-fallback.ts web/ui/src/index.ts web/ui/src/styles/components.css web/ui/tests/components/
git commit -m "feat(ui): add ScrollArea, ErrorFallback, and core index with all registrations"
```

---

## Chunk 2: Auth Package

### Task 10: Auth types and AuthInitError

**Files:**
- Create: `web/ui/src/auth/types.ts`

**Ref:** Spec lines 206-247

- [ ] **Step 1: Create types**

```ts
// src/auth/types.ts
export interface User {
  id: string;
  username: string;
  name: string;
  displayName: string;
  type: 'user' | 'agent' | 'service';
}

export interface Session {
  id: string;
  expiresAt: string;
  refreshedAt: string;
}

export type AuthEventMap = {
  change: User | null;
  logout: void;
};

export class AuthInitError extends Error {
  constructor(message: string, options?: { cause?: unknown }) {
    super(message, options);
    this.name = 'AuthInitError';
  }
}
```

- [ ] **Step 2: Verify types compile**

```bash
cd web/ui && npx tsc --noEmit src/auth/types.ts
```

- [ ] **Step 3: Commit**

```bash
git add web/ui/src/auth/types.ts
git commit -m "feat(ui/auth): add User, Session, AuthInitError types"
```

---

### Task 11: AuthClient implementation (TDD)

**Files:**
- Create: `web/ui/src/auth/client.ts`
- Create: `web/ui/tests/auth/client.test.ts`

**Ref:** Spec lines 175-258 (API, init error handling, refresh strategy)

- [ ] **Step 1: Write the failing tests**

```ts
// tests/auth/client.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { AuthClient } from '../../src/auth/client.js';
import { AuthInitError } from '../../src/auth/types.js';

const MOCK_USER = {
  id: '1', username: 'kazw', name: 'Kaz Walker',
  displayName: 'Kaz', type: 'user' as const,
};
const MOCK_SESSION = {
  id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z',
};

function mockSessionResponse(status = 200) {
  return new Response(
    status === 200 ? JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }) : '',
    { status, headers: status === 200 ? { 'Content-Type': 'application/json' } : {} },
  );
}

describe('AuthClient', () => {
  let client: AuthClient;

  beforeEach(() => {
    client = new AuthClient();
    vi.restoreAllMocks();
  });

  afterEach(() => { client.destroy(); });

  describe('init()', () => {
    it('fetches session and stores user', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse());
      await client.init();
      expect(client.getUser()).toEqual(MOCK_USER);
      expect(client.getSession()).toEqual(MOCK_SESSION);
      expect(client.isAuthenticated).toBe(true);
    });

    it('sets null on 401 without throwing', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse(401));
      await client.init();
      expect(client.getUser()).toBeNull();
      expect(client.isAuthenticated).toBe(false);
    });

    it('throws AuthInitError on network error', async () => {
      vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new TypeError('Failed to fetch'));
      await expect(client.init()).rejects.toThrow(AuthInitError);
    });

    it('throws AuthInitError on 500', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse(500));
      await expect(client.init()).rejects.toThrow(AuthInitError);
    });

    it('emits change event with user', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(mockSessionResponse());
      const handler = vi.fn();
      client.on('change', handler);
      await client.init();
      expect(handler).toHaveBeenCalledWith(MOCK_USER);
    });
  });

  describe('refresh()', () => {
    it('re-fetches session', async () => {
      const spy = vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(mockSessionResponse())
        .mockResolvedValueOnce(new Response(
          JSON.stringify({ user: { ...MOCK_USER, displayName: 'K' }, session: MOCK_SESSION }),
          { status: 200, headers: { 'Content-Type': 'application/json' } },
        ));
      await client.init();
      await client.refresh();
      expect(client.getUser()!.displayName).toBe('K');
      expect(spy).toHaveBeenCalledTimes(2);
    });

    it('emits logout on 401 during refresh', async () => {
      vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(mockSessionResponse())
        .mockResolvedValueOnce(mockSessionResponse(401));
      const logoutHandler = vi.fn();
      client.on('logout', logoutHandler);
      await client.init();
      await client.refresh();
      expect(client.isAuthenticated).toBe(false);
      expect(logoutHandler).toHaveBeenCalledOnce();
    });
  });

  describe('logout()', () => {
    it('clears state and emits logout', async () => {
      vi.spyOn(globalThis, 'fetch')
        .mockResolvedValueOnce(mockSessionResponse())
        .mockResolvedValueOnce(new Response('', { status: 200 }));
      await client.init();
      const logoutHandler = vi.fn();
      client.on('logout', logoutHandler);
      await client.logout();
      expect(client.getUser()).toBeNull();
      expect(client.isAuthenticated).toBe(false);
      expect(logoutHandler).toHaveBeenCalledOnce();
    });
  });

  describe('events', () => {
    it('off() removes listener', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      const handler = vi.fn();
      client.on('change', handler);
      client.off('change', handler);
      await client.init();
      expect(handler).not.toHaveBeenCalled();
    });
  });

  describe('visibility change', () => {
    it('calls refresh when visible after >5 min hidden', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      await client.init();
      // Simulate going hidden
      Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      // Advance time past threshold
      vi.useFakeTimers();
      vi.advanceTimersByTime(6 * 60 * 1000);
      vi.useRealTimers();
      // Simulate becoming visible
      Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true });
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      document.dispatchEvent(new Event('visibilitychange'));
      // refresh() should have been called (2nd fetch: init + refresh)
      await vi.waitFor(() => expect(globalThis.fetch).toHaveBeenCalled());
    });

    it('does not call refresh when hidden <5 min', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      await client.init();
      const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(mockSessionResponse());
      // Simulate going hidden then immediately visible
      Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
      expect(fetchSpy).not.toHaveBeenCalled();
    });
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd web/ui && npx vitest run tests/auth/client.test.ts
```

- [ ] **Step 3: Implement AuthClient**

```ts
// src/auth/client.ts
import type { User, Session, AuthEventMap } from './types.js';
import { AuthInitError } from './types.js';

type Listener<T> = (data: T) => void;

const SESSION_ENDPOINT = '/api/auth/v1/session';
const SIGNOUT_ENDPOINT = '/api/auth/v1/sign-out';
const VISIBILITY_THRESHOLD_MS = 5 * 60 * 1000;

export class AuthClient {
  private _user: User | null = null;
  private _session: Session | null = null;
  private _listeners = new Map<string, Set<Listener<any>>>();
  private _lastVisible = Date.now();
  private _visHandler: (() => void) | null = null;

  getUser(): User | null { return this._user; }
  getSession(): Session | null { return this._session; }
  get isAuthenticated(): boolean { return this._user !== null; }

  async init(): Promise<void> {
    await this._fetchSession();
    this._setupVisibilityListener();
  }

  async refresh(): Promise<void> {
    const wasAuth = this.isAuthenticated;
    await this._fetchSession();
    if (wasAuth && !this.isAuthenticated) {
      this._emit('logout', undefined as void);
    }
  }

  /** Clears session and emits events. Redirect to login is the shell's responsibility
   *  (it listens for the 'logout' event and navigates accordingly). */
  async logout(): Promise<void> {
    try {
      await fetch(SIGNOUT_ENDPOINT, { method: 'POST', credentials: 'include' });
    } catch { /* best-effort */ }
    this._user = null;
    this._session = null;
    this._emit('logout', undefined as void);
    this._emit('change', null);
  }

  on<K extends keyof AuthEventMap>(event: K, listener: Listener<AuthEventMap[K]>): void {
    if (!this._listeners.has(event)) this._listeners.set(event, new Set());
    this._listeners.get(event)!.add(listener);
  }

  off<K extends keyof AuthEventMap>(event: K, listener: Listener<AuthEventMap[K]>): void {
    this._listeners.get(event)?.delete(listener);
  }

  destroy(): void {
    if (this._visHandler) {
      document.removeEventListener('visibilitychange', this._visHandler);
      this._visHandler = null;
    }
  }

  private async _fetchSession(): Promise<void> {
    let res: Response;
    try {
      res = await fetch(SESSION_ENDPOINT, { credentials: 'include' });
    } catch (err) {
      throw new AuthInitError('Failed to reach auth service', { cause: err });
    }

    if (res.status === 401) {
      this._user = null;
      this._session = null;
      this._emit('change', null);
      return;
    }

    if (!res.ok) {
      throw new AuthInitError(`Auth service returned ${res.status}`);
    }

    let data: { user: User; session: Session };
    try {
      data = await res.json();
    } catch (err) {
      throw new AuthInitError('Invalid JSON from auth service', { cause: err });
    }

    this._user = data.user;
    this._session = data.session;
    this._emit('change', this._user);
  }

  private _setupVisibilityListener(): void {
    if (typeof document === 'undefined') return;
    this._visHandler = () => {
      if (document.visibilityState === 'visible') {
        if (Date.now() - this._lastVisible > VISIBILITY_THRESHOLD_MS) {
          this.refresh().catch(() => { /* best-effort; AuthInitError is non-fatal on visibility refresh */ });
        }
      } else {
        this._lastVisible = Date.now();
      }
    };
    document.addEventListener('visibilitychange', this._visHandler);
  }

  private _emit<K extends keyof AuthEventMap>(event: K, data: AuthEventMap[K]): void {
    this._listeners.get(event)?.forEach((fn) => fn(data));
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd web/ui && npx vitest run tests/auth/client.test.ts
```

- [ ] **Step 5: Commit**

```bash
git add web/ui/src/auth/client.ts web/ui/tests/auth/client.test.ts
git commit -m "feat(ui/auth): add AuthClient with init, refresh, logout, events"
```

---

### Task 12: Auth singleton and index

**Files:**
- Create: `web/ui/src/auth/index.ts`
- Create: `web/ui/tests/auth/singleton.test.ts`

**Ref:** Spec lines 260-278 (singleton pattern, single-package guarantee)

- [ ] **Step 1: Write the failing test**

```ts
// tests/auth/singleton.test.ts
import { describe, it, expect, beforeEach } from 'vitest';
import { getAuthClient, _resetAuthClient } from '../../src/auth/index.js';

describe('getAuthClient', () => {
  beforeEach(() => { _resetAuthClient(); });

  it('returns the same instance on repeated calls', () => {
    const a = getAuthClient();
    const b = getAuthClient();
    expect(a).toBe(b);
  });

  it('returns an AuthClient instance', () => {
    const client = getAuthClient();
    expect(typeof client.init).toBe('function');
    expect(typeof client.getUser).toBe('function');
  });
});
```

- [ ] **Step 2: Create auth index with singleton**

```ts
// src/auth/index.ts
import { AuthClient } from './client.js';

export { AuthClient } from './client.js';
export { AuthInitError } from './types.js';
export type { User, Session, AuthEventMap } from './types.js';

let instance: AuthClient | null = null;

/** Returns the singleton AuthClient. All adapters use this internally. */
export function getAuthClient(): AuthClient {
  if (!instance) instance = new AuthClient();
  return instance;
}

/** @internal Reset singleton for testing only. */
export function _resetAuthClient(): void {
  if (instance) instance.destroy();
  instance = null;
}
```

- [ ] **Step 3: Run all auth tests**

```bash
cd web/ui && npx vitest run tests/auth/
```

- [ ] **Step 4: Commit**

```bash
git add web/ui/src/auth/index.ts web/ui/tests/auth/singleton.test.ts
git commit -m "feat(ui/auth): add singleton pattern and auth index"
```

---

## Chunk 3: Framework Adapters + Build

### Task 13: React adapter

React needs wrapper components because React's Custom Element support is poor (event handling, boolean attributes). This is the only adapter with wrapper components.

**Files:**
- Create: `web/ui/src/react/components.tsx`, `use-auth.ts`, `use-theme.ts`, `index.tsx`
- Create: `web/ui/tests/react/components.test.tsx`, `use-auth.test.tsx`

**Ref:** Spec lines 294-315

- [ ] **Step 1: Write wrapper component test**

```tsx
// tests/react/components.test.tsx
import { describe, it, expect, afterEach, vi } from 'vitest';
import React from 'react';
import { render, cleanup } from '@testing-library/react';
import '../../src/index.js';
import { Panel, Button, Badge } from '../../src/react/index.js';

describe('React component wrappers', () => {
  afterEach(cleanup);

  it('Panel renders wf-panel with label', () => {
    const { container } = render(<Panel label="Test">Content</Panel>);
    const panel = container.querySelector('wf-panel');
    expect(panel).not.toBeNull();
    expect(panel!.getAttribute('label')).toBe('Test');
  });

  it('Button renders wf-button with variant', () => {
    const { container } = render(<Button variant="filled">Click</Button>);
    const button = container.querySelector('wf-button');
    expect(button!.getAttribute('variant')).toBe('filled');
  });

  it('Badge renders wf-badge with count', () => {
    const { container } = render(<Badge count={5} />);
    const badge = container.querySelector('wf-badge');
    expect(badge!.getAttribute('count')).toBe('5');
  });

  it('forwards onWfClick event to addEventListener on the Custom Element', () => {
    const handler = vi.fn();
    const { container } = render(<Button onWfClick={handler}>Click me</Button>);
    const button = container.querySelector('wf-button')!;
    button.dispatchEvent(new CustomEvent('wf-click', { bubbles: true }));
    expect(handler).toHaveBeenCalledOnce();
  });
});
```

- [ ] **Step 2: Write useAuth hook test**

```tsx
// tests/react/use-auth.test.tsx
import { describe, it, expect, afterEach, vi } from 'vitest';
import React from 'react';
import { render, cleanup, act } from '@testing-library/react';
import { useAuth } from '../../src/react/use-auth.js';
import { _resetAuthClient, getAuthClient } from '../../src/auth/index.js';

const MOCK_USER = {
  id: '1', username: 'kazw', name: 'Kaz Walker',
  displayName: 'Kaz', type: 'user' as const,
};
const MOCK_SESSION = {
  id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z',
};

function TestComponent() {
  const { user, isAuthenticated } = useAuth();
  return (
    <div>
      <span data-testid="auth">{isAuthenticated ? 'yes' : 'no'}</span>
      <span data-testid="user">{user?.username ?? 'none'}</span>
    </div>
  );
}

describe('useAuth (React)', () => {
  afterEach(() => { cleanup(); _resetAuthClient(); vi.restoreAllMocks(); });

  it('reflects auth state after init', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }), {
        status: 200, headers: { 'Content-Type': 'application/json' },
      }),
    );
    const { getByTestId } = render(<TestComponent />);
    expect(getByTestId('auth').textContent).toBe('no');
    await act(async () => { await getAuthClient().init(); });
    expect(getByTestId('auth').textContent).toBe('yes');
    expect(getByTestId('user').textContent).toBe('kazw');
  });
});
```

- [ ] **Step 3: Implement React wrapper components**

```tsx
// src/react/components.tsx
import React, { forwardRef, useRef, useEffect, useCallback } from 'react';
import '../index.js';

import type { WfPanel } from '../components/panel.js';
import type { WfButton } from '../components/button.js';
import type { WfBadge } from '../components/badge.js';
import type { WfStatusDot } from '../components/status-dot.js';
import type { WfSkeleton } from '../components/skeleton.js';
import type { WfTextInput } from '../components/text-input.js';
import type { WfList } from '../components/list.js';
import type { WfListItem } from '../components/list-item.js';
import type { WfScrollArea } from '../components/scroll-area.js';
import type { WfErrorFallback } from '../components/error-fallback.js';

type WfProps<E, P = {}> = P & React.HTMLAttributes<E> & { children?: React.ReactNode };

/**
 * Separates event props (onX) from attribute props, and attaches event listeners
 * via addEventListener on the Custom Element ref. This is needed because React 18
 * does not forward onX props to Custom Element addEventListener calls.
 * React 19+ handles this natively, but we support React 18.
 */
function useWcEvents<E extends HTMLElement>(
  forwardedRef: React.ForwardedRef<E>,
  props: Record<string, unknown>,
): { ref: React.RefCallback<E>; cleanProps: Record<string, unknown> } {
  const innerRef = useRef<E | null>(null);
  const listenersRef = useRef<Map<string, EventListener>>(new Map());

  const cleanProps: Record<string, unknown> = {};
  const eventProps: Record<string, EventListener> = {};

  for (const [key, val] of Object.entries(props)) {
    if (key.startsWith('on') && key.length > 2 && typeof val === 'function') {
      // Convert onWfClick → wf-click (camelCase to kebab-case)
      const raw = key[2].toLowerCase() + key.slice(3);
      const eventName = raw.replace(/([A-Z])/g, '-$1').toLowerCase();
      eventProps[eventName] = val as EventListener;
    } else {
      cleanProps[key] = val;
    }
  }

  const refCallback = useCallback((node: E | null) => {
    // Clean up old listeners
    if (innerRef.current) {
      listenersRef.current.forEach((fn, name) => innerRef.current!.removeEventListener(name, fn));
      listenersRef.current.clear();
    }
    innerRef.current = node;
    // Attach new listeners
    if (node) {
      for (const [name, fn] of Object.entries(eventProps)) {
        node.addEventListener(name, fn);
        listenersRef.current.set(name, fn);
      }
    }
    // Forward ref
    if (typeof forwardedRef === 'function') forwardedRef(node);
    else if (forwardedRef) forwardedRef.current = node;
  }, [forwardedRef, ...Object.keys(eventProps)]);

  return { ref: refCallback, cleanProps };
}

function wrapWc<E extends HTMLElement, P extends Record<string, unknown>>(
  tag: string,
  displayName: string,
) {
  const Comp = forwardRef<E, WfProps<E, P>>(({ children, ...rest }, ref) => {
    const { ref: wcRef, cleanProps } = useWcEvents<E>(ref, rest as Record<string, unknown>);
    return React.createElement(tag, { ref: wcRef, ...cleanProps }, children);
  });
  Comp.displayName = displayName;
  return Comp;
}

export const Panel = wrapWc<WfPanel, { label?: string }>('wf-panel', 'Panel');
export const Button = wrapWc<WfButton, { variant?: 'text' | 'filled'; disabled?: boolean }>('wf-button', 'Button');
export const Badge = wrapWc<WfBadge, { count?: number }>('wf-badge', 'Badge');
export const StatusDot = wrapWc<WfStatusDot, { status?: string }>('wf-status-dot', 'StatusDot');
export const Skeleton = wrapWc<WfSkeleton, { width?: string; height?: string }>('wf-skeleton', 'Skeleton');
export const Divider = wrapWc<HTMLElement, {}>('wf-divider', 'Divider');
export const TextInput = wrapWc<WfTextInput, { placeholder?: string; value?: string; disabled?: boolean }>('wf-text-input', 'TextInput');
export const List = wrapWc<WfList, {}>('wf-list', 'List');
export const ListItem = wrapWc<WfListItem, { active?: boolean }>('wf-list-item', 'ListItem');
export const ScrollArea = wrapWc<WfScrollArea, {}>('wf-scroll-area', 'ScrollArea');
export const ErrorFallback = wrapWc<WfErrorFallback, { title?: string; message?: string }>('wf-error-fallback', 'ErrorFallback');
```

- [ ] **Step 4: Implement useAuth and useTheme hooks**

```ts
// src/react/use-auth.ts
import { useSyncExternalStore, useCallback } from 'react';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

export function useAuth(): { user: User | null; isAuthenticated: boolean } {
  const client = getAuthClient();
  const subscribe = useCallback((cb: () => void) => {
    client.on('change', cb);
    client.on('logout', cb);
    return () => { client.off('change', cb); client.off('logout', cb); };
  }, [client]);
  const getSnapshot = useCallback(() => client.getUser(), [client]);
  const user = useSyncExternalStore(subscribe, getSnapshot);
  return { user, isAuthenticated: user !== null };
}
```

```ts
// src/react/use-theme.ts
import { useSyncExternalStore, useCallback } from 'react';

type Theme = 'dark' | 'light';

function getTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
}

export function useTheme(): Theme {
  const subscribe = useCallback((cb: () => void) => {
    if (typeof document === 'undefined') return () => {};
    const obs = new MutationObserver(() => cb());
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    return () => obs.disconnect();
  }, []);
  return useSyncExternalStore(subscribe, getTheme, () => 'dark' as Theme);
}
```

- [ ] **Step 5: Create React index**

```tsx
// src/react/index.tsx
export {
  Panel, Button, Badge, StatusDot, Skeleton, Divider,
  TextInput, List, ListItem, ScrollArea, ErrorFallback,
} from './components.js';
export { useAuth } from './use-auth.js';
export { useTheme } from './use-theme.js';
```

- [ ] **Step 6: Run all React tests**

```bash
cd web/ui && npx vitest run tests/react/
```

- [ ] **Step 7: Commit**

```bash
git add web/ui/src/react/ web/ui/tests/react/
git commit -m "feat(ui/react): add React adapter with wrapper components and hooks"
```

---

### Task 14: Vue adapter

Vue 3 handles Custom Elements natively. Adapter provides `useAuth()` and `useTheme()` composables.

**Files:**
- Create: `web/ui/src/vue/use-auth.ts`, `use-theme.ts`, `index.ts`
- Create: `web/ui/tests/vue/use-auth.test.ts`

- [ ] **Step 1: Implement Vue composables**

```ts
// src/vue/use-auth.ts
import { ref, readonly, onUnmounted } from 'vue';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

export function useAuth() {
  const client = getAuthClient();
  const user = ref<User | null>(client.getUser());
  const isAuthenticated = ref(client.isAuthenticated);

  const onChange = (u: User | null) => { user.value = u; isAuthenticated.value = u !== null; };
  const onLogout = () => { user.value = null; isAuthenticated.value = false; };

  client.on('change', onChange);
  client.on('logout', onLogout);

  try { onUnmounted(() => { client.off('change', onChange); client.off('logout', onLogout); }); }
  catch { /* not in component setup */ }

  return { user: readonly(user), isAuthenticated: readonly(isAuthenticated) };
}
```

```ts
// src/vue/use-theme.ts
import { ref, readonly, onUnmounted } from 'vue';

type Theme = 'dark' | 'light';

export function useTheme() {
  const theme = ref<Theme>(
    (typeof document !== 'undefined'
      ? (document.documentElement.getAttribute('data-theme') as Theme) : null) ?? 'dark',
  );
  let observer: MutationObserver | null = null;
  if (typeof document !== 'undefined') {
    observer = new MutationObserver(() => {
      theme.value = (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
  }
  try { onUnmounted(() => observer?.disconnect()); } catch { /* not in component setup */ }
  return readonly(theme);
}
```

```ts
// src/vue/index.ts
export { useAuth } from './use-auth.js';
export { useTheme } from './use-theme.js';
// Vue handles <wf-*> Custom Elements natively.
// Add compilerOptions.isCustomElement: (tag) => tag.startsWith('wf-') to Vue config.
```

- [ ] **Step 2: Write test**

```ts
// tests/vue/use-auth.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { nextTick } from 'vue';
import { _resetAuthClient, getAuthClient } from '../../src/auth/index.js';
import { useAuth } from '../../src/vue/use-auth.js';

const MOCK_USER = { id: '1', username: 'kazw', name: 'Kaz Walker', displayName: 'Kaz', type: 'user' as const };
const MOCK_SESSION = { id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z' };

describe('useAuth (Vue)', () => {
  afterEach(() => { _resetAuthClient(); vi.restoreAllMocks(); });

  it('returns reactive user after init', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    const { user, isAuthenticated } = useAuth();
    expect(user.value).toBeNull();
    await getAuthClient().init();
    await nextTick();
    expect(user.value).toEqual(MOCK_USER);
    expect(isAuthenticated.value).toBe(true);
  });
});
```

- [ ] **Step 3: Run Vue tests**

```bash
cd web/ui && npx vitest run tests/vue/
```

- [ ] **Step 4: Commit**

```bash
git add web/ui/src/vue/ web/ui/tests/vue/
git commit -m "feat(ui/vue): add Vue adapter with useAuth/useTheme composables"
```

---

### Task 15: Svelte adapter

Svelte handles Custom Elements natively. Adapter provides Svelte stores.

**Files:**
- Create: `web/ui/src/svelte/auth.ts`, `theme.ts`, `index.ts`
- Create: `web/ui/tests/svelte/auth.test.ts`

- [ ] **Step 1: Implement Svelte stores**

```ts
// src/svelte/auth.ts
import { readable, derived } from 'svelte/store';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

// getAuthClient() is called lazily inside the readable's start function,
// not at module scope. This avoids side-effecting the singleton on import
// and stays consistent with how React/Vue/Solid adapters call it inside
// their hook function bodies.
const user = readable<User | null>(null, (set) => {
  const client = getAuthClient();
  set(client.getUser());
  const onChange = (u: User | null) => set(u);
  const onLogout = () => set(null);
  client.on('change', onChange);
  client.on('logout', onLogout);
  return () => { client.off('change', onChange); client.off('logout', onLogout); };
});

const isAuthenticated = derived(user, ($user) => $user !== null);

export const auth = { user, isAuthenticated };
```

```ts
// src/svelte/theme.ts
import { readable } from 'svelte/store';

type Theme = 'dark' | 'light';

function getCurrentTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
}

export const theme = readable<Theme>(getCurrentTheme(), (set) => {
  if (typeof document === 'undefined') return;
  const obs = new MutationObserver(() => set(getCurrentTheme()));
  obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
  return () => obs.disconnect();
});
```

```ts
// src/svelte/index.ts
export { auth } from './auth.js';
export { theme } from './theme.js';
// Svelte handles <wf-*> Custom Elements natively in templates.
```

- [ ] **Step 2: Write test**

```ts
// tests/svelte/auth.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { _resetAuthClient, getAuthClient } from '../../src/auth/index.js';
import { auth } from '../../src/svelte/auth.js';

const MOCK_USER = { id: '1', username: 'kazw', name: 'Kaz Walker', displayName: 'Kaz', type: 'user' as const };
const MOCK_SESSION = { id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z' };

describe('auth store (Svelte)', () => {
  afterEach(() => { _resetAuthClient(); vi.restoreAllMocks(); });

  it('provides reactive user via store', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    expect(get(auth.user)).toBeNull();
    await getAuthClient().init();
    expect(get(auth.user)).toEqual(MOCK_USER);
    expect(get(auth.isAuthenticated)).toBe(true);
  });
});
```

- [ ] **Step 3: Run Svelte tests**

```bash
cd web/ui && npx vitest run tests/svelte/
```

- [ ] **Step 4: Commit**

```bash
git add web/ui/src/svelte/ web/ui/tests/svelte/
git commit -m "feat(ui/svelte): add Svelte adapter with auth/theme stores"
```

---

### Task 16: SolidJS adapter

SolidJS has excellent Custom Element interop. Adapter provides signal-based primitives.

**Files:**
- Create: `web/ui/src/solid/use-auth.ts`, `use-theme.ts`, `index.ts`
- Create: `web/ui/tests/solid/use-auth.test.ts`

- [ ] **Step 1: Implement SolidJS primitives**

```ts
// src/solid/use-auth.ts
import { createSignal, onCleanup } from 'solid-js';
import { getAuthClient } from '../auth/index.js';
import type { User } from '../auth/types.js';

export function useAuth() {
  const client = getAuthClient();
  const [user, setUser] = createSignal<User | null>(client.getUser());
  const isAuthenticated = () => user() !== null;

  const onChange = (u: User | null) => setUser(u);
  const onLogout = () => setUser(null);
  client.on('change', onChange);
  client.on('logout', onLogout);
  onCleanup(() => { client.off('change', onChange); client.off('logout', onLogout); });

  return { user, isAuthenticated };
}
```

```ts
// src/solid/use-theme.ts
import { createSignal, onCleanup } from 'solid-js';

type Theme = 'dark' | 'light';

function getCurrentTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return (document.documentElement.getAttribute('data-theme') as Theme) ?? 'dark';
}

export function useTheme() {
  const [theme, setTheme] = createSignal<Theme>(getCurrentTheme());
  if (typeof document !== 'undefined') {
    const obs = new MutationObserver(() => setTheme(getCurrentTheme()));
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
    onCleanup(() => obs.disconnect());
  }
  return theme;
}
```

```ts
// src/solid/index.ts
export { useAuth } from './use-auth.js';
export { useTheme } from './use-theme.js';
// SolidJS handles <wf-*> Custom Elements natively in JSX.
```

- [ ] **Step 2: Write test**

```ts
// tests/solid/use-auth.test.ts
import { describe, it, expect, afterEach, vi } from 'vitest';
import { createRoot } from 'solid-js';
import { _resetAuthClient, getAuthClient } from '../../src/auth/index.js';
import { useAuth } from '../../src/solid/use-auth.js';

const MOCK_USER = { id: '1', username: 'kazw', name: 'Kaz Walker', displayName: 'Kaz', type: 'user' as const };
const MOCK_SESSION = { id: 'sess-1', expiresAt: '2026-12-31T00:00:00Z', refreshedAt: '2026-03-12T00:00:00Z' };

describe('useAuth (Solid)', () => {
  afterEach(() => { _resetAuthClient(); vi.restoreAllMocks(); });

  it('provides reactive signals', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ user: MOCK_USER, session: MOCK_SESSION }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }),
    );
    await createRoot(async (dispose) => {
      const { user, isAuthenticated } = useAuth();
      expect(user()).toBeNull();
      await getAuthClient().init();
      expect(user()).toEqual(MOCK_USER);
      expect(isAuthenticated()).toBe(true);
      dispose();
    });
  });
});
```

- [ ] **Step 3: Run SolidJS tests**

```bash
cd web/ui && npx vitest run tests/solid/
```

- [ ] **Step 4: Commit**

```bash
git add web/ui/src/solid/ web/ui/tests/solid/
git commit -m "feat(ui/solid): add SolidJS adapter with useAuth/useTheme primitives"
```

---

### Task 17: Build verification and final integration

- [ ] **Step 1: Run the full build**

```bash
cd web/ui && npx vite build
```

Expected in `dist/`: `index.js`, `auth/index.js`, `react/index.js`, `vue/index.js`, `svelte/index.js`, `solid/index.js`, `style.css`, plus `.d.ts` files for each.

- [ ] **Step 2: Verify all entry points resolve**

```bash
cd web/ui && ls dist/index.js dist/auth/index.js dist/react/index.js dist/vue/index.js dist/svelte/index.js dist/solid/index.js dist/style.css
```

All files should exist.

- [ ] **Step 3: Run ALL tests**

```bash
cd web/ui && npx vitest run
```

Expected: All test suites pass (core, auth, react, vue, svelte, solid).

- [ ] **Step 4: Type check**

```bash
cd web/ui && npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 5: Commit**

```bash
git add web/ui/
git commit -m "feat(ui): verify full build and all tests pass"
```

---

## Spec Deviations

| Deviation | Spec says | Plan does | Rationale |
|-----------|-----------|-----------|-----------|
| Vue/Svelte/Solid wrappers | Typed component wrappers | Type augmentation docs only | These frameworks handle WC natively; wrappers add complexity with no DX benefit |
| Test runner for WCs | `@web/test-runner` | Vitest + happy-dom | Single runner is simpler; switch if happy-dom can't handle Lit light DOM |
| Storybook | Stories for each component | Not in scope | Add after core is working |
| `pnpm` | pnpm workspaces | npm | Single package doesn't need workspace tooling |
