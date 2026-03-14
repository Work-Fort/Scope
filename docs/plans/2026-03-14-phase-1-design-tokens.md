# Phase 1: Design Tokens — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restructure `@workfort/ui` styling into a three-tier design token system (primitives → semantic → component), rename token variables to follow the spec naming convention (`--wf-color-*`, `--wf-text-*`), and update all consumers (component CSS, shell global styles, UnoCSS bridge).

**Architecture:** CSS custom properties as the single source of truth, organized into three files:
- `primitives.css` (Tier 1) — raw scales: stone palette (50–950), status color palettes (red, amber, green, blue), numeric spacing, typography sizes/weights/line-heights, font stacks, radii, shadows, z-index, motion.
- `tokens.css` (Tier 2) — semantic aliases referencing primitives via `var()`. Dark mode on `:root`, light mode on `[data-theme="light"]`. Named spacing aliases (`xs`–`xl`) reference numeric primitives.
- `components.css` (Tier 3) — per-component tokens and structural styles. Component tokens reference semantic tokens.

**Tech Stack:** CSS custom properties, Vite (library mode CSS bundling), Vitest, UnoCSS

**Specs:**
- `docs/ui-component-library-design.md` — Design Tokens section, Token Categories, Phasing
- `docs/design-token-research.md` — Gap analysis (Section 6), Token Organization Strategy (Section 5.3)

---

## Token Rename Map

This phase renames variables to follow the spec's namespacing convention. All renames happen atomically in Chunk 3 to avoid broken intermediate states.

### Color tokens (add `color-` prefix)

| Old | New |
|-----|-----|
| `--wf-bg` | `--wf-color-bg` |
| `--wf-bg-secondary` | `--wf-color-bg-secondary` |
| `--wf-text` (color context) | `--wf-color-text` |
| `--wf-text-secondary` | `--wf-color-text-secondary` |
| `--wf-text-muted` | `--wf-color-text-muted` |
| `--wf-border` | `--wf-color-border` |
| `--wf-accent` | `--wf-color-accent` |
| `--wf-error` | `--wf-color-error` |
| `--wf-error-subtle` | `--wf-color-error-subtle` |
| `--wf-warning` | `--wf-color-warning` |
| `--wf-warning-subtle` | `--wf-color-warning-subtle` |
| `--wf-success` | `--wf-color-success` |
| `--wf-success-subtle` | `--wf-color-success-subtle` |

### Typography size tokens (shorten prefix)

| Old | New |
|-----|-----|
| `--wf-font-size-xs` | `--wf-text-xs` |
| `--wf-font-size-sm` | `--wf-text-sm` |
| `--wf-font-size-base` | `--wf-text-base` |
| `--wf-font-size-lg` | `--wf-text-lg` |

### Unchanged tokens

These keep their current names (already match spec):
- `--wf-space-xs`, `--wf-space-sm`, `--wf-space-md`, `--wf-space-lg`, `--wf-space-xl`
- `--wf-font-sans`, `--wf-font-mono`
- `--wf-radius-sm`, `--wf-radius-md`, `--wf-radius-lg`

### New semantic tokens added

| Token | Dark value | Light value |
|-------|-----------|-------------|
| `--wf-color-bg-elevated` | `var(--wf-stone-800)` | `var(--wf-stone-50)` |
| `--wf-color-bg-overlay` | `rgba(0, 0, 0, 0.5)` | `rgba(0, 0, 0, 0.3)` |
| `--wf-color-text-disabled` | `var(--wf-stone-700)` | `var(--wf-stone-300)` |
| `--wf-color-text-on-accent` | `var(--wf-stone-950)` | `var(--wf-stone-50)` |
| `--wf-color-border-focus` | `var(--wf-stone-400)` | `var(--wf-stone-600)` |
| `--wf-color-border-strong` | `var(--wf-stone-600)` | `var(--wf-stone-400)` |
| `--wf-color-info` | `var(--wf-blue-500)` | `var(--wf-blue-600)` |
| `--wf-color-info-subtle` | `color-mix(...)` | `color-mix(...)` |

### Hardcoded values to tokenize

| File | Hardcoded | New token reference |
|------|-----------|-------------------|
| `toast.css` | `box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15)` | `box-shadow: var(--wf-shadow-md)` |
| `toast.css` | `z-index: 9999` | `z-index: var(--wf-z-toast)` |
| `toast.css` | `200ms ease-out` (animation) | `var(--wf-duration-normal) var(--wf-ease-out)` |
| `banner.css` | `font-weight: 600` | `font-weight: var(--wf-weight-semibold)` |

---

## Chunk 1: Primitive Tokens (Tier 1)

### Task 1: Write token definition tests

**Files:**
- Create: `web/packages/ui/tests/tokens/definitions.test.ts`

- [ ] **Step 1: Create test directory**

```bash
mkdir -p web/packages/ui/tests/tokens
```

- [ ] **Step 2: Write the test file**

```typescript
// tests/tokens/definitions.test.ts
import { readFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { describe, it, expect } from 'vitest';

const __dirname = dirname(fileURLToPath(import.meta.url));
const stylesDir = resolve(__dirname, '../../src/styles');

function readCSS(filename: string): string {
  return readFileSync(resolve(stylesDir, filename), 'utf8');
}

describe('primitives.css', () => {
  const css = readCSS('primitives.css');

  it('defines stone palette (50-950)', () => {
    for (const stop of [50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950]) {
      expect(css, `missing --wf-stone-${stop}`).toContain(`--wf-stone-${stop}`);
    }
  });

  it('defines status color palettes', () => {
    for (const hue of ['red', 'amber', 'green', 'blue']) {
      for (const stop of [50, 500, 950]) {
        expect(css, `missing --wf-${hue}-${stop}`).toContain(`--wf-${hue}-${stop}`);
      }
    }
  });

  it('defines numeric spacing scale', () => {
    for (const n of [1, 2, 3, 4, 5, 6, 8, 10, 12, 16, 20, 24]) {
      expect(css, `missing --wf-space-${n}`).toContain(`--wf-space-${n}:`);
    }
  });

  it('defines typography size primitives', () => {
    for (const size of ['xs', 'sm', 'base', 'md', 'lg', 'xl', '2xl', '3xl', '4xl']) {
      expect(css, `missing --wf-text-${size}`).toContain(`--wf-text-${size}`);
    }
  });

  it('defines font weight primitives', () => {
    for (const w of ['normal', 'medium', 'semibold', 'bold']) {
      expect(css, `missing --wf-weight-${w}`).toContain(`--wf-weight-${w}`);
    }
  });

  it('defines line-height primitives', () => {
    for (const lh of ['tight', 'normal', 'relaxed']) {
      expect(css, `missing --wf-leading-${lh}`).toContain(`--wf-leading-${lh}`);
    }
  });

  it('defines font stack primitives', () => {
    expect(css).toContain('--wf-font-sans');
    expect(css).toContain('--wf-font-mono');
  });

  it('defines radius primitives', () => {
    for (const r of ['none', 'xs', 'sm', 'md', 'lg', 'xl', 'full']) {
      expect(css, `missing --wf-radius-${r}`).toContain(`--wf-radius-${r}`);
    }
  });

  it('defines shadow primitives', () => {
    for (const s of ['sm', 'md', 'lg', 'xl']) {
      expect(css, `missing --wf-shadow-${s}`).toContain(`--wf-shadow-${s}`);
    }
  });

  it('defines z-index scale', () => {
    for (const z of ['dropdown', 'sticky', 'modal', 'toast', 'tooltip']) {
      expect(css, `missing --wf-z-${z}`).toContain(`--wf-z-${z}`);
    }
  });

  it('defines motion primitives', () => {
    for (const d of ['fast', 'normal', 'slow']) {
      expect(css, `missing --wf-duration-${d}`).toContain(`--wf-duration-${d}`);
    }
    for (const e of ['in', 'out', 'in-out']) {
      expect(css, `missing --wf-ease-${e}`).toContain(`--wf-ease-${e}`);
    }
  });
});

describe('tokens.css', () => {
  const css = readCSS('tokens.css');

  it('defines semantic color tokens', () => {
    const expected = [
      '--wf-color-bg', '--wf-color-bg-secondary', '--wf-color-bg-elevated',
      '--wf-color-text', '--wf-color-text-secondary', '--wf-color-text-muted',
      '--wf-color-border', '--wf-color-accent',
      '--wf-color-error', '--wf-color-warning', '--wf-color-success', '--wf-color-info',
    ];
    for (const token of expected) {
      expect(css, `missing ${token}`).toContain(token);
    }
  });

  it('references primitives via var() — no hardcoded hex', () => {
    expect(css).toContain('var(--wf-stone-');
    // Semantic color definitions must not contain raw hex
    // (rgba() in subtle variants is acceptable — it wraps primitive refs via color-mix)
    const lines = css.split('\n').filter(l =>
      l.includes('--wf-color-') && l.includes(':') && !l.trim().startsWith('/*')
    );
    const hexInDefinition = lines.filter(l => /#[0-9a-f]{6}\b/i.test(l));
    expect(hexInDefinition, 'found hardcoded hex in semantic token definitions').toHaveLength(0);
  });

  it('has dark and light themes', () => {
    expect(css).toContain(':root');
    expect(css).toContain('[data-theme="light"]');
  });

  it('defines named spacing aliases referencing numeric primitives', () => {
    for (const name of ['xs', 'sm', 'md', 'lg', 'xl']) {
      expect(css, `missing --wf-space-${name}`).toContain(`--wf-space-${name}`);
    }
    // Aliases should reference var(--wf-space-N)
    expect(css).toContain('var(--wf-space-');
  });
});

describe('no old token names in component styles', () => {
  const files = ['components.css', 'banner.css', 'toast.css'];

  for (const file of files) {
    describe(file, () => {
      const css = readCSS(file);

      it('uses --wf-color-* (not old --wf-bg/--wf-text color names)', () => {
        const oldColorRefs = [
          'var(--wf-bg)',
          'var(--wf-bg-secondary)',
          'var(--wf-text)',
          'var(--wf-text-secondary)',
          'var(--wf-text-muted)',
          'var(--wf-border)',
          'var(--wf-accent)',
          'var(--wf-error)',
          'var(--wf-error-subtle)',
          'var(--wf-warning)',
          'var(--wf-warning-subtle)',
          'var(--wf-success)',
          'var(--wf-success-subtle)',
        ];
        for (const ref of oldColorRefs) {
          expect(css, `${file} still uses old token: ${ref}`).not.toContain(ref);
        }
      });

      it('uses --wf-text-* (not old --wf-font-size-* names)', () => {
        expect(css, `${file} still uses old --wf-font-size-*`).not.toContain('var(--wf-font-size-');
      });
    });
  }
});
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd web && pnpm --filter @workfort/ui test -- tests/tokens/definitions.test.ts
```

Expected: FAIL — `primitives.css` does not exist.

- [ ] **Step 4: Commit test file**

```bash
git add web/packages/ui/tests/tokens/definitions.test.ts
git commit -m "test: add design token definition tests (red — Phase 1)"
```

### Task 2: Create primitives.css

**Files:**
- Create: `web/packages/ui/src/styles/primitives.css`
- Modify: `web/packages/ui/src/index.ts:1`

- [ ] **Step 1: Create primitives.css**

Color hex values are sourced from the Tailwind stone, red, amber, green, and blue palettes — the same values already in use throughout `tokens.css` (currently hardcoded).

```css
/* primitives.css — Tier 1: raw scales. No semantic meaning. */

:root {
  /* ── Stone palette ── */
  --wf-stone-50: #fafaf9;
  --wf-stone-100: #f5f5f4;
  --wf-stone-200: #e7e5e4;
  --wf-stone-300: #d6d3d1;
  --wf-stone-400: #a8a29e;
  --wf-stone-500: #78716c;
  --wf-stone-600: #57534e;
  --wf-stone-700: #44403c;
  --wf-stone-800: #292524;
  --wf-stone-900: #1c1917;
  --wf-stone-950: #0c0a09;

  /* ── Red palette (status: error) ── */
  --wf-red-50: #fef2f2;
  --wf-red-100: #fee2e2;
  --wf-red-200: #fecaca;
  --wf-red-300: #fca5a5;
  --wf-red-400: #f87171;
  --wf-red-500: #ef4444;
  --wf-red-600: #dc2626;
  --wf-red-700: #b91c1c;
  --wf-red-800: #991b1b;
  --wf-red-900: #7f1d1d;
  --wf-red-950: #450a0a;

  /* ── Amber palette (status: warning) ── */
  --wf-amber-50: #fffbeb;
  --wf-amber-100: #fef3c7;
  --wf-amber-200: #fde68a;
  --wf-amber-300: #fcd34d;
  --wf-amber-400: #fbbf24;
  --wf-amber-500: #f59e0b;
  --wf-amber-600: #d97706;
  --wf-amber-700: #b45309;
  --wf-amber-800: #92400e;
  --wf-amber-900: #78350f;
  --wf-amber-950: #451a03;

  /* ── Green palette (status: success) ── */
  --wf-green-50: #f0fdf4;
  --wf-green-100: #dcfce7;
  --wf-green-200: #bbf7d0;
  --wf-green-300: #86efac;
  --wf-green-400: #4ade80;
  --wf-green-500: #22c55e;
  --wf-green-600: #16a34a;
  --wf-green-700: #15803d;
  --wf-green-800: #166534;
  --wf-green-900: #14532d;
  --wf-green-950: #052e16;

  /* ── Blue palette (status: info) ── */
  --wf-blue-50: #eff6ff;
  --wf-blue-100: #dbeafe;
  --wf-blue-200: #bfdbfe;
  --wf-blue-300: #93c5fd;
  --wf-blue-400: #60a5fa;
  --wf-blue-500: #3b82f6;
  --wf-blue-600: #2563eb;
  --wf-blue-700: #1d4ed8;
  --wf-blue-800: #1e40af;
  --wf-blue-900: #1e3a8a;
  --wf-blue-950: #172554;

  /* ── Spacing (numeric, rem-based) ── */
  --wf-space-1: 0.25rem;
  --wf-space-2: 0.5rem;
  --wf-space-3: 0.75rem;
  --wf-space-4: 1rem;
  --wf-space-5: 1.25rem;
  --wf-space-6: 1.5rem;
  --wf-space-8: 2rem;
  --wf-space-10: 2.5rem;
  --wf-space-12: 3rem;
  --wf-space-16: 4rem;
  --wf-space-20: 5rem;
  --wf-space-24: 6rem;

  /* ── Typography — sizes ── */
  --wf-text-xs: 0.75rem;
  --wf-text-sm: 0.8125rem;
  --wf-text-base: 0.875rem;
  --wf-text-md: 1rem;
  --wf-text-lg: 1.125rem;
  --wf-text-xl: 1.25rem;
  --wf-text-2xl: 1.5rem;
  --wf-text-3xl: 1.875rem;
  --wf-text-4xl: 2.25rem;

  /* ── Typography — weights ── */
  --wf-weight-normal: 400;
  --wf-weight-medium: 500;
  --wf-weight-semibold: 600;
  --wf-weight-bold: 700;

  /* ── Typography — line heights ── */
  --wf-leading-tight: 1.25;
  --wf-leading-normal: 1.5;
  --wf-leading-relaxed: 1.75;

  /* ── Typography — font stacks ── */
  --wf-font-sans: ui-sans-serif, system-ui, -apple-system, sans-serif;
  --wf-font-mono: ui-monospace, 'SF Mono', 'Fira Code', monospace;

  /* ── Border radius ── */
  --wf-radius-none: 0;
  --wf-radius-xs: 4px;
  --wf-radius-sm: 6px;
  --wf-radius-md: 8px;
  --wf-radius-lg: 12px;
  --wf-radius-xl: 16px;
  --wf-radius-full: 9999px;

  /* ── Shadows ── */
  --wf-shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.08);
  --wf-shadow-md: 0 4px 12px rgba(0, 0, 0, 0.12);
  --wf-shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.16);
  --wf-shadow-xl: 0 16px 48px rgba(0, 0, 0, 0.2);

  /* ── Z-index layers ── */
  --wf-z-dropdown: 100;
  --wf-z-sticky: 200;
  --wf-z-modal: 300;
  --wf-z-toast: 400;
  --wf-z-tooltip: 500;

  /* ── Motion — durations ── */
  --wf-duration-fast: 100ms;
  --wf-duration-normal: 200ms;
  --wf-duration-slow: 300ms;

  /* ── Motion — easings ── */
  --wf-ease-in: cubic-bezier(0.4, 0, 1, 0.2);
  --wf-ease-out: cubic-bezier(0, 0, 0.2, 1);
  --wf-ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
}
```

- [ ] **Step 2: Add primitives.css import in index.ts**

In `web/packages/ui/src/index.ts`, add the import before `tokens.css`:

```typescript
import './styles/primitives.css';
import './styles/tokens.css';
import './styles/components.css';
// ... rest of imports unchanged
```

- [ ] **Step 3: Run primitives tests**

```bash
cd web && pnpm --filter @workfort/ui test -- tests/tokens/definitions.test.ts -t "primitives"
```

Expected: All `primitives.css` tests PASS. The `tokens.css` and `no old token names` suites still fail (that's expected — they're fixed in Chunks 2–3).

- [ ] **Step 4: Run build to verify CSS bundling**

```bash
cd web && pnpm --filter @workfort/ui build
```

Expected: Build succeeds. `web/packages/ui/dist/style.css` contains primitive variable definitions.

- [ ] **Step 5: Commit**

```bash
git add web/packages/ui/src/styles/primitives.css web/packages/ui/src/index.ts
git commit -m "feat(ui): add primitives.css with full design token scales"
```

---

## Chunk 2: Semantic Token Refactor (Tier 2)

### Task 3: Rewrite tokens.css with new naming

This task replaces the entire `tokens.css` file. The new version:
1. Renames all variables to follow the spec convention (`--wf-color-*`, `--wf-text-*`)
2. References primitives via `var()` instead of hardcoding hex values
3. Adds missing semantic tokens (info, elevated, overlay, disabled, focus, strong)
4. Adds named spacing aliases that reference numeric primitives
5. Uses `color-mix()` for subtle status variants (references semantic color tokens, so theme changes cascade automatically)

**Files:**
- Modify: `web/packages/ui/src/styles/tokens.css`

- [ ] **Step 1: Replace tokens.css**

```css
/* tokens.css — Tier 2: semantic aliases. Reference primitives via var(). */

@import './primitives.css';

/* ── Dark theme (default) ── */
:root {
  /* Surfaces */
  --wf-color-bg: var(--wf-stone-950);
  --wf-color-bg-secondary: var(--wf-stone-900);
  --wf-color-bg-elevated: var(--wf-stone-800);
  --wf-color-bg-overlay: rgba(0, 0, 0, 0.5);

  /* Text */
  --wf-color-text: var(--wf-stone-200);
  --wf-color-text-secondary: var(--wf-stone-400);
  --wf-color-text-muted: var(--wf-stone-600);
  --wf-color-text-disabled: var(--wf-stone-700);
  --wf-color-text-on-accent: var(--wf-stone-950);

  /* Borders */
  --wf-color-border: var(--wf-stone-800);
  --wf-color-border-focus: var(--wf-stone-400);
  --wf-color-border-strong: var(--wf-stone-600);

  /* Accent */
  --wf-color-accent: var(--wf-stone-200);

  /* Status — primary */
  --wf-color-error: var(--wf-red-500);
  --wf-color-warning: var(--wf-amber-500);
  --wf-color-success: var(--wf-green-500);
  --wf-color-info: var(--wf-blue-500);

  /* Status — subtle backgrounds */
  --wf-color-error-subtle: color-mix(in srgb, var(--wf-color-error) 12%, transparent);
  --wf-color-warning-subtle: color-mix(in srgb, var(--wf-color-warning) 12%, transparent);
  --wf-color-success-subtle: color-mix(in srgb, var(--wf-color-success) 12%, transparent);
  --wf-color-info-subtle: color-mix(in srgb, var(--wf-color-info) 12%, transparent);

  /* Named spacing aliases (reference numeric primitives) */
  --wf-space-xs: var(--wf-space-1);
  --wf-space-sm: var(--wf-space-2);
  --wf-space-md: var(--wf-space-3);
  --wf-space-lg: var(--wf-space-4);
  --wf-space-xl: var(--wf-space-6);
}

/* ── Light theme ── */
[data-theme="light"] {
  /* Surfaces */
  --wf-color-bg: var(--wf-stone-50);
  --wf-color-bg-secondary: var(--wf-stone-100);
  --wf-color-bg-elevated: var(--wf-stone-50);
  --wf-color-bg-overlay: rgba(0, 0, 0, 0.3);

  /* Text */
  --wf-color-text: var(--wf-stone-900);
  --wf-color-text-secondary: var(--wf-stone-600);
  --wf-color-text-muted: var(--wf-stone-400);
  --wf-color-text-disabled: var(--wf-stone-300);
  --wf-color-text-on-accent: var(--wf-stone-50);

  /* Borders */
  --wf-color-border: var(--wf-stone-200);
  --wf-color-border-focus: var(--wf-stone-600);
  --wf-color-border-strong: var(--wf-stone-400);

  /* Accent */
  --wf-color-accent: var(--wf-stone-900);

  /* Status — primary */
  --wf-color-error: var(--wf-red-600);
  --wf-color-warning: var(--wf-amber-600);
  --wf-color-success: var(--wf-green-600);
  --wf-color-info: var(--wf-blue-600);

  /* Status — subtle backgrounds (lower opacity on light backgrounds) */
  --wf-color-error-subtle: color-mix(in srgb, var(--wf-color-error) 8%, transparent);
  --wf-color-warning-subtle: color-mix(in srgb, var(--wf-color-warning) 8%, transparent);
  --wf-color-success-subtle: color-mix(in srgb, var(--wf-color-success) 8%, transparent);
  --wf-color-info-subtle: color-mix(in srgb, var(--wf-color-info) 8%, transparent);
}
```

- [ ] **Step 2: Run tokens.css tests**

```bash
cd web && pnpm --filter @workfort/ui test -- tests/tokens/definitions.test.ts -t "tokens.css"
```

Expected: All `tokens.css` tests PASS.

- [ ] **Step 3: Commit**

```bash
git add web/packages/ui/src/styles/tokens.css
git commit -m "refactor(ui): rewrite tokens.css to reference primitives, add new semantic tokens"
```

---

## Chunk 3: Component Style Migration

### Task 4: Update components.css

Rename all token references and add per-component tokens. The component token definitions go at the top of the file as `:root` variables so consumers can override them.

**Files:**
- Modify: `web/packages/ui/src/styles/components.css`

- [ ] **Step 1: Replace components.css**

```css
/* components.css — Tier 3: component tokens + structural styles */
@import './banner.css';
@import './toast.css';

/* ── Component-level tokens (Tier 3) ── */
:root {
  --wf-panel-padding: var(--wf-space-md);
  --wf-panel-radius: var(--wf-radius-md);
  --wf-button-padding-x: var(--wf-space-md);
  --wf-button-padding-y: var(--wf-space-sm);
  --wf-button-radius: var(--wf-radius-md);
  --wf-input-height: 2rem;
  --wf-input-padding: var(--wf-space-sm);
  --wf-input-radius: var(--wf-radius-md);
}

/* Panel */
.wf-panel {
  display: block;
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-panel-radius);
  padding: var(--wf-panel-padding);
  font-family: var(--wf-font-sans);
  color: var(--wf-color-text);
}
.wf-panel__label {
  font-size: var(--wf-text-sm);
  color: var(--wf-color-text-muted);
  margin-bottom: var(--wf-space-sm);
  font-weight: var(--wf-weight-medium);
}

/* Button */
.wf-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--wf-space-sm);
  padding: var(--wf-button-padding-y) var(--wf-button-padding-x);
  background: none;
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-button-radius);
  color: var(--wf-color-text);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-sm);
  cursor: pointer;
}
.wf-button[disabled] { opacity: 0.5; cursor: not-allowed; pointer-events: none; }
.wf-button--filled { background: var(--wf-color-accent); border-color: var(--wf-color-accent); color: var(--wf-color-bg); }

/* Badge */
.wf-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.25em;
  height: 1.25em;
  padding: 0 var(--wf-space-xs);
  border-radius: var(--wf-radius-lg);
  background: var(--wf-color-accent);
  color: var(--wf-color-bg);
  font-size: var(--wf-text-xs);
  font-family: var(--wf-font-sans);
  line-height: 1;
}
.wf-badge:empty { display: none; }

/* StatusDot */
.wf-status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: var(--wf-radius-full);
  background: var(--wf-color-text-muted);
}
.wf-status-dot--online { background: var(--wf-color-success); }
.wf-status-dot--away { background: var(--wf-color-warning); }
.wf-status-dot--offline { background: var(--wf-color-error); }

/* Skeleton */
.wf-skeleton {
  display: block;
  background: var(--wf-color-bg-secondary);
  border-radius: var(--wf-radius-sm);
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
  background: var(--wf-color-border);
  margin: var(--wf-space-sm) 0;
}

/* TextInput */
.wf-text-input { display: block; }
.wf-text-input__input {
  width: 100%;
  height: var(--wf-input-height);
  padding: 0 var(--wf-input-padding);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-base);
  color: var(--wf-color-text);
  background: var(--wf-color-bg);
  border: 1px solid var(--wf-color-border);
  border-radius: var(--wf-input-radius);
  outline: none;
  box-sizing: border-box;
}
.wf-text-input__input:focus {
  border-color: var(--wf-color-accent);
}
.wf-text-input[disabled] .wf-text-input__input {
  opacity: 0.5;
  cursor: not-allowed;
}

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
  padding: var(--wf-space-sm) var(--wf-space-md);
  cursor: pointer;
  color: var(--wf-color-text);
  border-radius: var(--wf-radius-sm);
}
.wf-list-item:hover { background: var(--wf-color-bg-secondary); }
.wf-list-item--active { background: var(--wf-color-bg-secondary); }
.wf-list-item__trailing {
  margin-left: auto;
  display: flex;
  align-items: center;
}

/* ScrollArea */
.wf-scroll-area {
  display: block;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--wf-color-border) transparent;
}
.wf-scroll-area::-webkit-scrollbar { width: 6px; }
.wf-scroll-area::-webkit-scrollbar-track { background: transparent; }
.wf-scroll-area::-webkit-scrollbar-thumb {
  background: var(--wf-color-border);
  border-radius: var(--wf-radius-full);
}

/* ErrorFallback */
.wf-error-fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--wf-space-sm);
  padding: var(--wf-space-xl);
  color: var(--wf-color-text-muted);
  font-family: var(--wf-font-sans);
}
.wf-error-fallback__title {
  font-size: var(--wf-text-lg);
  font-weight: var(--wf-weight-semibold);
  color: var(--wf-color-text);
}
.wf-error-fallback__message {
  font-size: var(--wf-text-sm);
}
```

- [ ] **Step 2: Verify components.css compiles**

```bash
cd web && pnpm --filter @workfort/ui build
```

Expected: Build succeeds. No errors about undefined variables (primitives provide all referenced values).

### Task 5: Update banner.css

**Files:**
- Modify: `web/packages/ui/src/styles/banner.css`

- [ ] **Step 1: Replace banner.css**

```css
/* banner.css — Banner component styles */

.wf-banner {
  display: block;
  padding: var(--wf-space-sm) var(--wf-space-lg);
  border-left: 4px solid var(--wf-color-border);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-sm);
}

.wf-banner--error {
  background: var(--wf-color-error-subtle);
  border-left-color: var(--wf-color-error);
}
.wf-banner--warning {
  background: var(--wf-color-warning-subtle);
  border-left-color: var(--wf-color-warning);
}
.wf-banner--info {
  background: var(--wf-color-bg-secondary);
  border-left-color: var(--wf-color-accent);
}

.wf-banner__content {
  display: flex;
  align-items: center;
  gap: var(--wf-space-sm);
}

.wf-banner__icon {
  flex-shrink: 0;
  font-size: var(--wf-text-xs);
}
.wf-banner--error .wf-banner__icon { color: var(--wf-color-error); }
.wf-banner--warning .wf-banner__icon { color: var(--wf-color-warning); }
.wf-banner--info .wf-banner__icon { color: var(--wf-color-accent); }

.wf-banner__headline {
  flex: 1;
  font-weight: var(--wf-weight-semibold);
  color: var(--wf-color-text);
}

.wf-banner__actions {
  display: flex;
  align-items: center;
  gap: var(--wf-space-xs);
}

.wf-banner__toggle,
.wf-banner__close {
  background: none;
  border: none;
  color: var(--wf-color-text-secondary);
  cursor: pointer;
  padding: var(--wf-space-xs);
  font-size: var(--wf-text-sm);
  line-height: 1;
}
.wf-banner__toggle:hover,
.wf-banner__close:hover {
  color: var(--wf-color-text);
}

.wf-banner__details {
  margin-top: var(--wf-space-sm);
  padding-left: calc(var(--wf-space-sm) + 0.5rem + var(--wf-space-sm));
  font-family: var(--wf-font-mono);
  font-size: var(--wf-text-xs);
  color: var(--wf-color-text-secondary);
  white-space: pre-wrap;
}
```

### Task 6: Update toast.css

This file also tokenizes previously hardcoded values (shadow, z-index, animation timing).

**Files:**
- Modify: `web/packages/ui/src/styles/toast.css`

- [ ] **Step 1: Replace toast.css**

```css
/* toast.css — Toast and ToastContainer styles */

.wf-toast {
  display: flex;
  align-items: center;
  gap: var(--wf-space-sm);
  padding: var(--wf-space-sm) var(--wf-space-md);
  border-left: 4px solid var(--wf-color-border);
  border-radius: var(--wf-radius-md);
  background: var(--wf-color-bg);
  color: var(--wf-color-text);
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-sm);
  box-shadow: var(--wf-shadow-md);
  min-width: 240px;
  max-width: 400px;
  animation: wf-toast-in var(--wf-duration-normal) var(--wf-ease-out);
}

@keyframes wf-toast-in {
  from { opacity: 0; transform: translateX(16px); }
  to { opacity: 1; transform: translateX(0); }
}

.wf-toast--error { border-left-color: var(--wf-color-error); background: var(--wf-color-error-subtle); }
.wf-toast--warning { border-left-color: var(--wf-color-warning); background: var(--wf-color-warning-subtle); }
.wf-toast--info { border-left-color: var(--wf-color-accent); }
.wf-toast--success { border-left-color: var(--wf-color-success); background: var(--wf-color-success-subtle); }

.wf-toast__close {
  background: none;
  border: none;
  color: var(--wf-color-text-secondary);
  cursor: pointer;
  padding: var(--wf-space-xs);
  font-size: var(--wf-text-sm);
  line-height: 1;
  margin-left: auto;
}
.wf-toast__close:hover { color: var(--wf-color-text); }

.wf-toast-container {
  position: fixed;
  z-index: var(--wf-z-toast);
  display: flex;
  flex-direction: column;
  gap: var(--wf-space-sm);
  pointer-events: none;
}
.wf-toast-container > * { pointer-events: auto; }

.wf-toast-container--top-right { top: var(--wf-space-lg); right: var(--wf-space-lg); }
.wf-toast-container--top-left { top: var(--wf-space-lg); left: var(--wf-space-lg); }
.wf-toast-container--bottom-right { bottom: var(--wf-space-lg); right: var(--wf-space-lg); }
.wf-toast-container--bottom-left { bottom: var(--wf-space-lg); left: var(--wf-space-lg); }
```

### Task 7: Run all tests and verify build

- [ ] **Step 1: Run token definition tests**

```bash
cd web && pnpm --filter @workfort/ui test -- tests/tokens/definitions.test.ts
```

Expected: ALL tests PASS — primitives defined, tokens reference primitives, no old token names in component styles.

- [ ] **Step 2: Run all existing component tests**

```bash
cd web && pnpm --filter @workfort/ui test
```

Expected: All component behavioral tests pass. The CSS changes are style-only — no component TypeScript was modified. Tests use happy-dom which doesn't evaluate CSS, so styling changes cannot break them.

- [ ] **Step 3: Build and verify output**

```bash
cd web && pnpm --filter @workfort/ui build
```

Expected: Build succeeds. Verify `web/packages/ui/dist/style.css` contains the new token names:

```bash
grep -c '\-\-wf-color-' web/packages/ui/dist/style.css
grep -c '\-\-wf-stone-' web/packages/ui/dist/style.css
```

Expected: Both return non-zero counts. The old `--wf-bg:` (definition, not reference) should NOT appear:

```bash
grep '\-\-wf-bg:' web/packages/ui/dist/style.css
```

Expected: No output (old definition removed).

- [ ] **Step 4: Commit**

```bash
git add web/packages/ui/src/styles/
git commit -m "refactor(ui): migrate component styles to new token naming convention

Rename --wf-bg → --wf-color-bg, --wf-font-size-* → --wf-text-*, etc.
Add component-level tokens (panel, button, input).
Tokenize hardcoded shadow, z-index, and motion values in toast."
```

---

## Chunk 4: Shell & UnoCSS Bridge Migration

### Task 8: Update shell global.css

The shell's `global.css` references token variables for layout styling. All old token names must be updated.

**Files:**
- Modify: `web/shell/src/global.css`

- [ ] **Step 1: Replace global.css**

```css
*,
*::before,
*::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html, body, #app {
  height: 100%;
  font-family: var(--wf-font-sans);
  font-size: var(--wf-text-base);
  background: var(--wf-color-bg);
  color: var(--wf-color-text);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.shell-layout {
  display: grid;
  grid-template-rows: auto auto 1fr;
  grid-template-columns: 240px 1fr;
  grid-template-areas:
    "banners banners"
    "nav     nav"
    "sidebar content";
  height: 100%;
}

.shell-layout--no-sidebar {
  grid-template-columns: 1fr;
  grid-template-areas:
    "banners"
    "nav"
    "content";
}

.shell-banners {
  grid-area: banners;
}

.shell-nav {
  grid-area: nav;
  display: flex;
  align-items: center;
  gap: var(--wf-space-lg);
  padding: 0 var(--wf-space-lg);
  height: 48px;
  background: var(--wf-color-bg);
  border-bottom: 1px solid var(--wf-color-border);
}

.shell-nav__brand {
  font-weight: var(--wf-weight-semibold);
  font-size: var(--wf-text-sm);
  letter-spacing: -0.02em;
  color: var(--wf-color-text);
  white-space: nowrap;
  text-transform: uppercase;
}

.shell-nav__tabs {
  display: flex;
  align-items: center;
  gap: var(--wf-space-xs);
  flex: 1;
}

.shell-nav__spacer {
  flex: 1;
}

.shell-sidebar {
  grid-area: sidebar;
  border-right: 1px solid var(--wf-color-border);
  overflow-y: auto;
}

.shell-content {
  grid-area: content;
  overflow-y: auto;
  padding: var(--wf-space-lg);
}

.shell-unavailable {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.shell-nav__tab--disabled {
  opacity: 0.4;
  cursor: default;
}

.shell-nav__tabs wf-status-dot {
  margin-right: var(--wf-space-xs);
}
```

- [ ] **Step 2: Verify no old token names remain in global.css**

```bash
grep -n 'var(--wf-bg)' web/shell/src/global.css
grep -n 'var(--wf-text)' web/shell/src/global.css | grep -v 'color-text' | grep -v 'text-'
grep -n 'var(--wf-font-size-' web/shell/src/global.css
```

Expected: No output from any grep (old names fully replaced).

### Task 9: Update shell uno.config.ts

Expand the UnoCSS theme bridge to include the new semantic tokens. This lets the shell use utility classes like `bg-wf-bg-elevated` or `text-wf-error`.

**Files:**
- Modify: `web/shell/uno.config.ts`

- [ ] **Step 1: Replace uno.config.ts**

```typescript
import { defineConfig, presetWind } from 'unocss';

export default defineConfig({
  presets: [presetWind()],
  theme: {
    colors: {
      wf: {
        bg: 'var(--wf-color-bg)',
        'bg-secondary': 'var(--wf-color-bg-secondary)',
        'bg-elevated': 'var(--wf-color-bg-elevated)',
        text: 'var(--wf-color-text)',
        'text-secondary': 'var(--wf-color-text-secondary)',
        'text-muted': 'var(--wf-color-text-muted)',
        'text-disabled': 'var(--wf-color-text-disabled)',
        border: 'var(--wf-color-border)',
        'border-focus': 'var(--wf-color-border-focus)',
        accent: 'var(--wf-color-accent)',
        error: 'var(--wf-color-error)',
        warning: 'var(--wf-color-warning)',
        success: 'var(--wf-color-success)',
        info: 'var(--wf-color-info)',
      },
    },
  },
});
```

- [ ] **Step 2: Commit shell changes**

```bash
git add web/shell/src/global.css web/shell/uno.config.ts
git commit -m "refactor(shell): migrate to new token naming, expand UnoCSS bridge"
```

---

## Chunk 5: Verification & Cleanup

### Task 10: Full workspace verification

- [ ] **Step 1: Run all UI package tests**

```bash
cd web && pnpm --filter @workfort/ui test
```

Expected: All tests pass (token definitions + component behavior).

- [ ] **Step 2: Run full workspace build**

```bash
cd web && pnpm build
```

Expected: All packages build successfully.

- [ ] **Step 3: Verify built CSS contains no old token names**

```bash
# Check built style.css for any old token definitions
echo "=== Should find ZERO old definitions ==="
grep -cE '\-\-wf-(bg|text|border|accent|error|warning|success|font-size-)' web/packages/ui/dist/style.css | head -1 || echo "0"

echo "=== Should find new definitions ==="
grep -c '\-\-wf-color-' web/packages/ui/dist/style.css
grep -c '\-\-wf-stone-' web/packages/ui/dist/style.css
grep -c '\-\-wf-text-xs' web/packages/ui/dist/style.css
```

Expected: Zero old definitions, non-zero new definitions.

- [ ] **Step 4: Spot-check token cascade**

Open `web/packages/ui/dist/style.css` and verify:
1. `primitives.css` content appears first (`:root { --wf-stone-50: ...}`)
2. `tokens.css` content appears second (`:root { --wf-color-bg: var(--wf-stone-950); ...}`)
3. `components.css` content appears last (`.wf-panel { ... var(--wf-color-bg) ...}`)

This ordering ensures the cascade resolves correctly.

- [ ] **Step 5: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "fix(ui): final adjustments from Phase 1 verification"
```
