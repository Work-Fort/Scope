# UI Package Split — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split the monolithic `@workfort/ui` into 5 independent npm packages in a pnpm workspace.

**Spec:** `docs/2026-03-12-workfort-ui-wc-design.md` — see "Amendment: Package Split" section.

**Architecture:** pnpm workspace at `web/` with packages under `web/packages/{ui,react,vue,svelte,solid}/`. Core Lit Web Components in `@workfort/ui`. Framework adapters in `@workfort/ui-react`, `@workfort/ui-vue`, `@workfort/ui-svelte`, `@workfort/ui-solid`. Each framework package peer-depends on `@workfort/ui` and hard-depends on `@workfort/auth`. Follows passport monorepo conventions.

**Tech Stack:** pnpm workspaces, TypeScript, Vite (ui + react), tsc (vue/svelte/solid), vitest, happy-dom

---

## File Structure

### Files to modify (toolchain)

| File | Change |
|------|--------|
| `mise.toml` | Add `node = "lts"` and `"npm:pnpm" = "10"` to `[tools]` |

### New files to create

| File | Responsibility |
|------|---------------|
| `web/pnpm-workspace.yaml` | Declares `packages/*` as workspace members |
| `web/package.json` | Private workspace root with `build` and `test` scripts |
| `web/packages/ui/package.json` | `@workfort/ui` package metadata, deps: `lit` |
| `web/packages/ui/tsconfig.json` | TypeScript config for dev (includes tests) |
| `web/packages/ui/tsconfig.build.json` | TypeScript config for build (excludes tests) |
| `web/packages/ui/vite.config.ts` | Vite library build: single entry, CSS bundling, externalize `lit` |
| `web/packages/ui/vitest.config.ts` | Vitest config: happy-dom, test pattern |
| `web/packages/react/package.json` | `@workfort/ui-react` metadata, deps: `@workfort/auth`, peers: `@workfort/ui`, `react` |
| `web/packages/react/tsconfig.json` | TypeScript config with `jsx: react-jsx` |
| `web/packages/react/tsconfig.build.json` | Build config (excludes tests) |
| `web/packages/react/vite.config.ts` | Vite build with React plugin, externalize `@workfort/ui`, `@workfort/auth`, `react` |
| `web/packages/react/vitest.config.ts` | Vitest with React plugin |
| `web/packages/vue/package.json` | `@workfort/ui-vue` metadata, deps: `@workfort/auth`, peers: `@workfort/ui`, `vue` |
| `web/packages/vue/tsconfig.json` | TypeScript config with `outDir: dist`, `declaration: true` |
| `web/packages/vue/vitest.config.ts` | Vitest config |
| `web/packages/svelte/package.json` | `@workfort/ui-svelte` metadata, deps: `@workfort/auth`, peers: `@workfort/ui`, `svelte` |
| `web/packages/svelte/tsconfig.json` | TypeScript config with `outDir: dist`, `declaration: true` |
| `web/packages/svelte/vitest.config.ts` | Vitest config |
| `web/packages/solid/package.json` | `@workfort/ui-solid` metadata, deps: `@workfort/auth`, peers: `@workfort/ui`, `solid-js` |
| `web/packages/solid/tsconfig.json` | TypeScript config with `outDir: dist`, `declaration: true` |
| `web/packages/solid/vitest.config.ts` | Vitest config |
| `.github/workflows/release-react.yml` | npm publish workflow for `@workfort/ui-react`, tag prefix `ui-react-v` |
| `.github/workflows/release-vue.yml` | npm publish workflow for `@workfort/ui-vue`, tag prefix `ui-vue-v` |
| `.github/workflows/release-svelte.yml` | npm publish workflow for `@workfort/ui-svelte`, tag prefix `ui-svelte-v` |
| `.github/workflows/release-solid.yml` | npm publish workflow for `@workfort/ui-solid`, tag prefix `ui-solid-v` |

### Files to move (no modifications needed unless noted)

| From | To | Notes |
|------|----|-------|
| `web/ui/src/index.ts` | `web/packages/ui/src/index.ts` | Unchanged |
| `web/ui/src/base.ts` | `web/packages/ui/src/base.ts` | Unchanged |
| `web/ui/src/components/` | `web/packages/ui/src/components/` | All 11 files unchanged |
| `web/ui/src/styles/` | `web/packages/ui/src/styles/` | Both CSS files unchanged |
| `web/ui/tests/helpers.ts` | `web/packages/ui/tests/helpers.ts` | Unchanged |
| `web/ui/tests/components/` | `web/packages/ui/tests/components/` | All 8 test files unchanged |
| `web/ui/src/react/components.tsx` | `web/packages/react/src/components.tsx` | **Modify imports** (see Task 3) |
| `web/ui/src/react/index.tsx` | `web/packages/react/src/index.tsx` | **Modify imports** |
| `web/ui/src/react/use-auth.ts` | `web/packages/react/src/use-auth.ts` | Unchanged |
| `web/ui/src/react/use-theme.ts` | `web/packages/react/src/use-theme.ts` | Unchanged |
| `web/ui/tests/react/components.test.tsx` | `web/packages/react/tests/components.test.tsx` | **Modify imports** |
| `web/ui/tests/react/use-auth.test.tsx` | `web/packages/react/tests/use-auth.test.tsx` | **Modify imports** |
| `web/ui/src/vue/` | `web/packages/vue/src/` | All 3 files unchanged |
| `web/ui/tests/vue/` | `web/packages/vue/tests/` | 1 test file unchanged |
| `web/ui/src/svelte/` | `web/packages/svelte/src/` | All 3 files unchanged |
| `web/ui/tests/svelte/` | `web/packages/svelte/tests/` | 1 test file unchanged |
| `web/ui/src/solid/` | `web/packages/solid/src/` | All 3 files unchanged |
| `web/ui/tests/solid/` | `web/packages/solid/tests/` | 1 test file unchanged |

### Files to modify

| File | Change |
|------|--------|
| `web/packages/react/src/components.tsx` | Change `import '../index.js'` → `import '@workfort/ui'`; change type imports from relative to `@workfort/ui` |
| `.github/workflows/release-cli.yml:8` | Change `paths-ignore: 'web/ui/**'` → `paths-ignore: 'web/**'` |
| `.github/workflows/release-ui.yml` | Full rewrite: `checkout@v6`, `mise-action@v3`, `default_bump: false`, pnpm workspace paths |

### Files to delete

| File | Reason |
|------|--------|
| `web/ui/` (entire directory) | All files moved to `web/packages/`. Old config files (`package.json`, `vite.config.ts`, `vitest.workspace.ts`, `tsconfig.json`, `tsconfig.build.json`) replaced by per-package configs. |

---

## Chunk 1: Workspace Root + Core UI Package

### Task 1: Create workspace root

**Files:**
- Modify: `mise.toml`
- Create: `web/pnpm-workspace.yaml`
- Create: `web/package.json`

- [ ] **Step 1: Add node and pnpm to `mise.toml`**

Add `node` and `"npm:pnpm"` to the `[tools]` section (matching passport repo conventions):

```toml
[tools]
go = "latest"
node = "lts"
"npm:pnpm" = "10"
```

- [ ] **Step 2: Create `web/pnpm-workspace.yaml`**

```yaml
packages:
  - "packages/*"
```

- [ ] **Step 3: Create `web/package.json`**

```json
{
  "private": true,
  "type": "module",
  "packageManager": "pnpm@10.32.1",
  "scripts": {
    "build": "pnpm -r run build",
    "test": "pnpm -r run test"
  }
}
```

- [ ] **Step 4: Commit**

```bash
git add mise.toml web/pnpm-workspace.yaml web/package.json
git commit -m "chore: create pnpm workspace root for web packages"
```

---

### Task 2: Create `@workfort/ui` core package

**Files:**
- Create: `web/packages/ui/package.json`
- Create: `web/packages/ui/tsconfig.json`
- Create: `web/packages/ui/tsconfig.build.json`
- Create: `web/packages/ui/vite.config.ts`
- Create: `web/packages/ui/vitest.config.ts`
- Move: `web/ui/src/{index.ts,base.ts}` → `web/packages/ui/src/`
- Move: `web/ui/src/components/` → `web/packages/ui/src/components/`
- Move: `web/ui/src/styles/` → `web/packages/ui/src/styles/`
- Move: `web/ui/tests/helpers.ts` → `web/packages/ui/tests/helpers.ts`
- Move: `web/ui/tests/components/` → `web/packages/ui/tests/components/`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p web/packages/ui/src web/packages/ui/tests
```

- [ ] **Step 2: Create `web/packages/ui/package.json`**

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
    "./style.css": "./dist/style.css"
  },
  "files": ["dist"],
  "scripts": {
    "build": "vite build",
    "test": "vitest run"
  },
  "dependencies": {
    "lit": "^3.2.0"
  },
  "devDependencies": {
    "typescript": "^5.6.0",
    "vite": "^6.0.0",
    "vite-plugin-dts": "^4.0.0",
    "vitest": "^2.1.0",
    "happy-dom": "^15.0.0"
  }
}
```

- [ ] **Step 3: Create `web/packages/ui/tsconfig.json`**

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
    "useDefineForClassFields": false
  },
  "include": ["src", "tests"]
}
```

- [ ] **Step 4: Create `web/packages/ui/tsconfig.build.json`**

```json
{
  "extends": "./tsconfig.json",
  "exclude": ["tests"]
}
```

- [ ] **Step 5: Create `web/packages/ui/vite.config.ts`**

```typescript
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';

export default defineConfig({
  plugins: [
    dts({ tsconfigPath: './tsconfig.build.json' }),
  ],
  build: {
    lib: {
      entry: 'src/index.ts',
      formats: ['es'],
      cssFileName: 'style',
    },
    rollupOptions: {
      external: ['lit', /^lit\//],
    },
    cssCodeSplit: false,
  },
});
```

- [ ] **Step 6: Create `web/packages/ui/vitest.config.ts`**

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['tests/**/*.test.ts'],
    environment: 'happy-dom',
  },
});
```

- [ ] **Step 7: Move source files**

```bash
git mv web/ui/src/index.ts web/packages/ui/src/index.ts
git mv web/ui/src/base.ts web/packages/ui/src/base.ts
git mv web/ui/src/components web/packages/ui/src/components
git mv web/ui/src/styles web/packages/ui/src/styles
```

No modifications needed — these files have no cross-package imports. Using `git mv` preserves file history. The remaining `web/ui/` contents (framework dirs, config files) are deleted in Task 9 (Chunk 3).

- [ ] **Step 8: Move test files**

```bash
git mv web/ui/tests/helpers.ts web/packages/ui/tests/helpers.ts
git mv web/ui/tests/components web/packages/ui/tests/components
```

No modifications needed — test helpers and component tests use relative imports only.

- [ ] **Step 9: Install dependencies**

```bash
cd web && pnpm install
```

Expected: lockfile generated, `lit`, `vite`, `vitest`, `happy-dom` installed.

- [ ] **Step 10: Run tests**

```bash
cd web && pnpm --filter @workfort/ui test
```

Expected: All 27 component tests pass (8 test files: button, error-fallback, leaf, list, panel, registration, scroll-area, text-input).

- [ ] **Step 11: Run build**

```bash
cd web && pnpm --filter @workfort/ui build
```

Expected: `web/packages/ui/dist/index.js` and `web/packages/ui/dist/style.css` exist.

- [ ] **Step 12: Commit**

```bash
git add web/packages/ui/ web/pnpm-lock.yaml
git commit -m "feat: create @workfort/ui core package in workspace"
```

---

## Chunk 2: Framework Packages

### Task 3: Create `@workfort/ui-react` package

**Files:**
- Create: `web/packages/react/package.json`
- Create: `web/packages/react/tsconfig.json`
- Create: `web/packages/react/tsconfig.build.json`
- Create: `web/packages/react/vite.config.ts`
- Create: `web/packages/react/vitest.config.ts`
- Move + modify: `web/ui/src/react/{index.tsx,components.tsx,use-auth.ts,use-theme.ts}` → `web/packages/react/src/`
- Move: `web/ui/tests/react/{components.test.tsx,use-auth.test.tsx}` → `web/packages/react/tests/`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p web/packages/react/src web/packages/react/tests
```

- [ ] **Step 2: Create `web/packages/react/package.json`**

```json
{
  "name": "@workfort/ui-react",
  "version": "0.0.1",
  "type": "module",
  "license": "Apache-2.0",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "default": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "vite build",
    "test": "vitest run"
  },
  "dependencies": {
    "@workfort/auth": "^0.0.1"
  },
  "peerDependencies": {
    "@workfort/ui": "^0.0.1",
    "react": "^18.0.0 || ^19.0.0"
  },
  "devDependencies": {
    "@workfort/ui": "workspace:*",
    "@testing-library/react": "^16.0.0",
    "@types/react": "^18.3.0",
    "@vitejs/plugin-react": "^4.3.0",
    "react": "^18.3.0",
    "react-dom": "^18.3.0",
    "typescript": "^5.6.0",
    "vite": "^6.0.0",
    "vite-plugin-dts": "^4.0.0",
    "vitest": "^2.1.0",
    "happy-dom": "^15.0.0"
  }
}
```

`@workfort/ui` appears in both `peerDependencies` (for published consumers) and `devDependencies` via `workspace:*` (for local dev/testing).

- [ ] **Step 3: Create `web/packages/react/tsconfig.json`**

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
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "jsx": "react-jsx"
  },
  "include": ["src", "tests"]
}
```

- [ ] **Step 4: Create `web/packages/react/tsconfig.build.json`**

```json
{
  "extends": "./tsconfig.json",
  "exclude": ["tests"]
}
```

- [ ] **Step 5: Create `web/packages/react/vite.config.ts`**

```typescript
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
      entry: 'src/index.tsx',
      formats: ['es'],
    },
    rollupOptions: {
      external: [
        '@workfort/ui',
        '@workfort/auth',
        'react',
        'react/jsx-runtime',
        'react-dom',
      ],
    },
  },
});
```

- [ ] **Step 6: Create `web/packages/react/vitest.config.ts`**

```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    include: ['tests/**/*.test.tsx'],
    environment: 'happy-dom',
  },
});
```

- [ ] **Step 7: Move source files**

```bash
git mv web/ui/src/react/use-auth.ts web/packages/react/src/use-auth.ts
git mv web/ui/src/react/use-theme.ts web/packages/react/src/use-theme.ts
git mv web/ui/src/react/index.tsx web/packages/react/src/index.tsx
git mv web/ui/src/react/components.tsx web/packages/react/src/components.tsx
```

- [ ] **Step 8: Update imports in `web/packages/react/src/components.tsx`**

Change the top of the file from:

```typescript
import React, { forwardRef, useRef, useCallback } from 'react';
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
```

To:

```typescript
import React, { forwardRef, useRef, useCallback } from 'react';
import '@workfort/ui';

import type {
  WfPanel, WfButton, WfBadge, WfStatusDot, WfSkeleton,
  WfTextInput, WfList, WfListItem, WfScrollArea, WfErrorFallback,
} from '@workfort/ui';
```

The rest of the file is unchanged — `WfProps`, `useWcEvents`, `wrapWc`, and all component exports stay exactly as they are.

- [ ] **Step 9: Move test files**

```bash
git mv web/ui/tests/react/components.test.tsx web/packages/react/tests/components.test.tsx
git mv web/ui/tests/react/use-auth.test.tsx web/packages/react/tests/use-auth.test.tsx
```

- [ ] **Step 10: Update test imports in `web/packages/react/tests/components.test.tsx`**

Change lines 4-5 from:

```typescript
import '../../src/index.js';
import { Panel, Button, Badge } from '../../src/react/index.js';
```

To:

```typescript
import '@workfort/ui';
import { Panel, Button, Badge } from '../src/index.js';
```

- [ ] **Step 11: Update test imports in `web/packages/react/tests/use-auth.test.tsx`**

Change line 4 from:

```typescript
import { useAuth } from '../../src/react/use-auth.js';
```

To:

```typescript
import { useAuth } from '../src/use-auth.js';
```

- [ ] **Step 12: Install dependencies**

```bash
cd web && pnpm install
```

Expected: `@workfort/ui` resolved via workspace link, React + testing library installed.

- [ ] **Step 13: Run tests**

```bash
cd web && pnpm --filter @workfort/ui-react test
```

Expected: All 5 React tests pass (4 component tests + 1 use-auth test).

- [ ] **Step 14: Run build**

```bash
cd web && pnpm --filter @workfort/ui-react build
```

Expected: `web/packages/react/dist/index.js` exists.

- [ ] **Step 15: Commit**

```bash
git add web/packages/react/ web/pnpm-lock.yaml
git commit -m "feat: create @workfort/ui-react package with component wrappers and hooks"
```

---

### Task 4: Create `@workfort/ui-vue` package

**Files:**
- Create: `web/packages/vue/package.json`
- Create: `web/packages/vue/tsconfig.json`
- Create: `web/packages/vue/vitest.config.ts`
- Move: `web/ui/src/vue/{index.ts,use-auth.ts,use-theme.ts}` → `web/packages/vue/src/`
- Move: `web/ui/tests/vue/use-auth.test.ts` → `web/packages/vue/tests/`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p web/packages/vue/src web/packages/vue/tests
```

- [ ] **Step 2: Create `web/packages/vue/package.json`**

```json
{
  "name": "@workfort/ui-vue",
  "version": "0.0.1",
  "type": "module",
  "license": "Apache-2.0",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsc",
    "test": "vitest run",
    "prepublishOnly": "tsc"
  },
  "dependencies": {
    "@workfort/auth": "^0.0.1"
  },
  "peerDependencies": {
    "@workfort/ui": "^0.0.1",
    "vue": "^3.3.0"
  },
  "devDependencies": {
    "@workfort/ui": "workspace:*",
    "typescript": "^5.6.0",
    "vue": "^3.5.0",
    "vitest": "^2.1.0",
    "happy-dom": "^15.0.0"
  }
}
```

`@workfort/ui` appears in both `peerDependencies` (for published consumers) and `devDependencies` via `workspace:*` (for local dev/testing).

- [ ] **Step 3: Create `web/packages/vue/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "outDir": "dist",
    "rootDir": "src",
    "declaration": true,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src"],
  "exclude": ["src/**/*.test.ts"]
}
```

- [ ] **Step 4: Create `web/packages/vue/vitest.config.ts`**

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['tests/**/*.test.ts'],
    environment: 'happy-dom',
  },
});
```

- [ ] **Step 5: Move source and test files**

```bash
git mv web/ui/src/vue/index.ts web/packages/vue/src/index.ts
git mv web/ui/src/vue/use-auth.ts web/packages/vue/src/use-auth.ts
git mv web/ui/src/vue/use-theme.ts web/packages/vue/src/use-theme.ts
git mv web/ui/tests/vue/use-auth.test.ts web/packages/vue/tests/use-auth.test.ts
```

No source import changes needed — `use-auth.ts` already imports from `@workfort/auth`, `use-theme.ts` uses only DOM APIs, `index.ts` re-exports from relative paths.

- [ ] **Step 6: Update test imports in `web/packages/vue/tests/use-auth.test.ts`**

Change line 4 from:

```typescript
import { useAuth } from '../../src/vue/use-auth.js';
```

To:

```typescript
import { useAuth } from '../src/use-auth.js';
```

- [ ] **Step 7: Install dependencies**

```bash
cd web && pnpm install
```

- [ ] **Step 8: Run tests**

```bash
cd web && pnpm --filter @workfort/ui-vue test
```

Expected: 1 Vue auth test passes.

- [ ] **Step 9: Run build**

```bash
cd web && pnpm --filter @workfort/ui-vue build
```

Expected: `web/packages/vue/dist/index.js` and `web/packages/vue/dist/index.d.ts` exist.

- [ ] **Step 10: Commit**

```bash
git add web/packages/vue/ web/pnpm-lock.yaml
git commit -m "feat: create @workfort/ui-vue package with auth and theme composables"
```

---

### Task 5: Create `@workfort/ui-svelte` package

**Files:**
- Create: `web/packages/svelte/package.json`
- Create: `web/packages/svelte/tsconfig.json`
- Create: `web/packages/svelte/vitest.config.ts`
- Move: `web/ui/src/svelte/{index.ts,auth.ts,theme.ts}` → `web/packages/svelte/src/`
- Move: `web/ui/tests/svelte/auth.test.ts` → `web/packages/svelte/tests/`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p web/packages/svelte/src web/packages/svelte/tests
```

- [ ] **Step 2: Create `web/packages/svelte/package.json`**

```json
{
  "name": "@workfort/ui-svelte",
  "version": "0.0.1",
  "type": "module",
  "license": "Apache-2.0",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsc",
    "test": "vitest run",
    "prepublishOnly": "tsc"
  },
  "dependencies": {
    "@workfort/auth": "^0.0.1"
  },
  "peerDependencies": {
    "@workfort/ui": "^0.0.1",
    "svelte": "^4.0.0 || ^5.0.0"
  },
  "devDependencies": {
    "@workfort/ui": "workspace:*",
    "svelte": "^5.0.0",
    "typescript": "^5.6.0",
    "vitest": "^2.1.0",
    "happy-dom": "^15.0.0"
  }
}
```

`@workfort/ui` appears in both `peerDependencies` (for published consumers) and `devDependencies` via `workspace:*` (for local dev/testing).

- [ ] **Step 3: Create `web/packages/svelte/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "outDir": "dist",
    "rootDir": "src",
    "declaration": true,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src"],
  "exclude": ["src/**/*.test.ts"]
}
```

- [ ] **Step 4: Create `web/packages/svelte/vitest.config.ts`**

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['tests/**/*.test.ts'],
    environment: 'happy-dom',
  },
});
```

- [ ] **Step 5: Move source and test files**

```bash
git mv web/ui/src/svelte/index.ts web/packages/svelte/src/index.ts
git mv web/ui/src/svelte/auth.ts web/packages/svelte/src/auth.ts
git mv web/ui/src/svelte/theme.ts web/packages/svelte/src/theme.ts
git mv web/ui/tests/svelte/auth.test.ts web/packages/svelte/tests/auth.test.ts
```

No source import changes needed — `auth.ts` already imports from `@workfort/auth`, `theme.ts` uses only DOM APIs.

- [ ] **Step 6: Update test imports in `web/packages/svelte/tests/auth.test.ts`**

Change line 4 from:

```typescript
import { auth } from '../../src/svelte/auth.js';
```

To:

```typescript
import { auth } from '../src/auth.js';
```

- [ ] **Step 7: Install dependencies**

```bash
cd web && pnpm install
```

- [ ] **Step 8: Run tests**

```bash
cd web && pnpm --filter @workfort/ui-svelte test
```

Expected: 1 Svelte auth test passes.

- [ ] **Step 9: Run build**

```bash
cd web && pnpm --filter @workfort/ui-svelte build
```

Expected: `web/packages/svelte/dist/index.js` and `web/packages/svelte/dist/index.d.ts` exist.

- [ ] **Step 10: Commit**

```bash
git add web/packages/svelte/ web/pnpm-lock.yaml
git commit -m "feat: create @workfort/ui-svelte package with auth and theme stores"
```

---

### Task 6: Create `@workfort/ui-solid` package

**Files:**
- Create: `web/packages/solid/package.json`
- Create: `web/packages/solid/tsconfig.json`
- Create: `web/packages/solid/vitest.config.ts`
- Move: `web/ui/src/solid/{index.ts,use-auth.ts,use-theme.ts}` → `web/packages/solid/src/`
- Move: `web/ui/tests/solid/use-auth.test.ts` → `web/packages/solid/tests/`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p web/packages/solid/src web/packages/solid/tests
```

- [ ] **Step 2: Create `web/packages/solid/package.json`**

```json
{
  "name": "@workfort/ui-solid",
  "version": "0.0.1",
  "type": "module",
  "license": "Apache-2.0",
  "main": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsc",
    "test": "vitest run",
    "prepublishOnly": "tsc"
  },
  "dependencies": {
    "@workfort/auth": "^0.0.1"
  },
  "peerDependencies": {
    "@workfort/ui": "^0.0.1",
    "solid-js": "^1.8.0"
  },
  "devDependencies": {
    "@workfort/ui": "workspace:*",
    "solid-js": "^1.9.0",
    "typescript": "^5.6.0",
    "vitest": "^2.1.0",
    "happy-dom": "^15.0.0"
  }
}
```

`@workfort/ui` appears in both `peerDependencies` (for published consumers) and `devDependencies` via `workspace:*` (for local dev/testing).

- [ ] **Step 3: Create `web/packages/solid/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "Node16",
    "moduleResolution": "Node16",
    "outDir": "dist",
    "rootDir": "src",
    "declaration": true,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src"],
  "exclude": ["src/**/*.test.ts"]
}
```

- [ ] **Step 4: Create `web/packages/solid/vitest.config.ts`**

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['tests/**/*.test.ts'],
    environment: 'happy-dom',
  },
});
```

- [ ] **Step 5: Move source and test files**

```bash
git mv web/ui/src/solid/index.ts web/packages/solid/src/index.ts
git mv web/ui/src/solid/use-auth.ts web/packages/solid/src/use-auth.ts
git mv web/ui/src/solid/use-theme.ts web/packages/solid/src/use-theme.ts
git mv web/ui/tests/solid/use-auth.test.ts web/packages/solid/tests/use-auth.test.ts
```

No source import changes needed — `use-auth.ts` already imports from `@workfort/auth`, `use-theme.ts` uses only DOM APIs.

- [ ] **Step 6: Update test imports in `web/packages/solid/tests/use-auth.test.ts`**

Change line 4 from:

```typescript
import { useAuth } from '../../src/solid/use-auth.js';
```

To:

```typescript
import { useAuth } from '../src/use-auth.js';
```

- [ ] **Step 7: Install dependencies**

```bash
cd web && pnpm install
```

- [ ] **Step 8: Run tests**

```bash
cd web && pnpm --filter @workfort/ui-solid test
```

Expected: 1 Solid auth test passes.

- [ ] **Step 9: Run build**

```bash
cd web && pnpm --filter @workfort/ui-solid build
```

Expected: `web/packages/solid/dist/index.js` and `web/packages/solid/dist/index.d.ts` exist.

- [ ] **Step 10: Commit**

```bash
git add web/packages/solid/ web/pnpm-lock.yaml
git commit -m "feat: create @workfort/ui-solid package with auth and theme hooks"
```

---

## Chunk 3: Workflows, Cleanup, Full Verification

### Task 7: Update existing release workflows

**Files:**
- Modify: `.github/workflows/release-cli.yml`
- Modify: `.github/workflows/release-ui.yml`

All workflows must follow passport repo conventions: `actions/checkout@v6`, `jdx/mise-action@v3` (mise manages node + pnpm), `default_bump: false` (require conventional commits).

- [ ] **Step 1: Update `release-cli.yml` paths-ignore**

Change line 8:

```yaml
# Before:
      - 'web/ui/**'
# After:
      - 'web/**'
```

- [ ] **Step 2: Rewrite `release-ui.yml`**

The workflow currently uses `npm ci`, `actions/setup-node`, and `actions/checkout@v4`. Rewrite the entire file to match passport conventions:

```yaml
# SPDX-License-Identifier: GPL-3.0-or-later
name: Release UI Package

on:
  push:
    branches: [master]
    paths:
      - 'web/packages/ui/**'

permissions:
  contents: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: jdx/mise-action@v3

      - name: Install and build
        run: cd web && pnpm install && pnpm --filter @workfort/ui build

      - name: Test
        run: cd web && pnpm --filter @workfort/ui test

  tag:
    name: Create SDK Tag
    runs-on: ubuntu-latest
    needs: build
    outputs:
      new_tag: ${{ steps.tag_version.outputs.new_tag }}
      changelog: ${{ steps.tag_version.outputs.changelog }}
      should_release: ${{ steps.check_tag.outputs.should_release }}

    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Bump version and create tag
        id: tag_version
        uses: Work-Fort/github-tag-action@v6.3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          default_bump: false
          release_branches: master
          tag_prefix: ui-v
          paths: web/packages/ui/**

      - name: Check if new tag was created
        id: check_tag
        run: |
          if [ -n "$TAG" ]; then
            echo "should_release=true" >> "$GITHUB_OUTPUT"
          else
            echo "should_release=false" >> "$GITHUB_OUTPUT"
          fi
        env:
          TAG: ${{ steps.tag_version.outputs.new_tag }}

  release:
    name: Publish to npm
    runs-on: ubuntu-latest
    needs: tag
    if: needs.tag.outputs.should_release == 'true'

    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ needs.tag.outputs.new_tag }}

      - uses: jdx/mise-action@v3

      - name: Build and publish
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-v||')
          cd web
          pnpm install
          cd packages/ui
          npm pkg set version="$VERSION"
          pnpm run build
          npm publish --provenance --access public
          npm pack
          mv workfort-ui-*.tgz ../../..

      - name: Create release notes
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
          CHANGELOG: ${{ needs.tag.outputs.changelog }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-v||')
          cat > release-notes.md << EOF
          # @workfort/ui v${VERSION}

          ## What's Changed

          ${CHANGELOG}

          ## Installation

          \`\`\`bash
          npm install @workfort/ui@${VERSION}
          \`\`\`

          ---

          Built automatically by Scope CI
          EOF

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ needs.tag.outputs.new_tag }}
          name: UI SDK ${{ needs.tag.outputs.new_tag }}
          body_path: release-notes.md
          draft: false
          prerelease: false
          files: workfort-ui-*.tgz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release-cli.yml .github/workflows/release-ui.yml
git commit -m "fix(ci): update workflows for workspace layout, migrate to mise"
```

---

### Task 8: Create framework release workflows

**Files:**
- Create: `.github/workflows/release-react.yml`
- Create: `.github/workflows/release-vue.yml`
- Create: `.github/workflows/release-svelte.yml`
- Create: `.github/workflows/release-solid.yml`

Each workflow follows the same structure as the updated `release-ui.yml` (passport conventions: `checkout@v6`, `mise-action@v3`, `default_bump: false`) with different path filter, tag prefix, pnpm filter, and package name.

- [ ] **Step 1: Create `.github/workflows/release-react.yml`**

```yaml
# SPDX-License-Identifier: GPL-3.0-or-later
name: Release React Package

on:
  push:
    branches: [master]
    paths:
      - 'web/packages/react/**'

permissions:
  contents: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: jdx/mise-action@v3

      - name: Install and build
        run: cd web && pnpm install && pnpm --filter @workfort/ui-react build

      - name: Test
        run: cd web && pnpm --filter @workfort/ui-react test

  tag:
    name: Create SDK Tag
    runs-on: ubuntu-latest
    needs: build
    outputs:
      new_tag: ${{ steps.tag_version.outputs.new_tag }}
      changelog: ${{ steps.tag_version.outputs.changelog }}
      should_release: ${{ steps.check_tag.outputs.should_release }}

    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Bump version and create tag
        id: tag_version
        uses: Work-Fort/github-tag-action@v6.3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          default_bump: false
          release_branches: master
          tag_prefix: ui-react-v
          paths: web/packages/react/**

      - name: Check if new tag was created
        id: check_tag
        run: |
          if [ -n "$TAG" ]; then
            echo "should_release=true" >> "$GITHUB_OUTPUT"
          else
            echo "should_release=false" >> "$GITHUB_OUTPUT"
          fi
        env:
          TAG: ${{ steps.tag_version.outputs.new_tag }}

  release:
    name: Publish to npm
    runs-on: ubuntu-latest
    needs: tag
    if: needs.tag.outputs.should_release == 'true'

    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ needs.tag.outputs.new_tag }}

      - uses: jdx/mise-action@v3

      - name: Build and publish
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-react-v||')
          cd web
          pnpm install
          cd packages/react
          npm pkg set version="$VERSION"
          pnpm run build
          npm publish --provenance --access public
          npm pack
          mv workfort-ui-react-*.tgz ../../..

      - name: Create release notes
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
          CHANGELOG: ${{ needs.tag.outputs.changelog }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-react-v||')
          cat > release-notes.md << EOF
          # @workfort/ui-react v${VERSION}

          ## What's Changed

          ${CHANGELOG}

          ## Installation

          \`\`\`bash
          npm install @workfort/ui-react@${VERSION}
          \`\`\`

          ---

          Built automatically by Scope CI
          EOF

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ needs.tag.outputs.new_tag }}
          name: React SDK ${{ needs.tag.outputs.new_tag }}
          body_path: release-notes.md
          draft: false
          prerelease: false
          files: workfort-ui-react-*.tgz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Create `.github/workflows/release-vue.yml`**

Same structure as `release-react.yml` with these substitutions:
- Name: `Release Vue Package`
- Path: `web/packages/vue/**`
- Tag prefix: `vue-v`
- Filter: `@workfort/ui-vue`
- Sed: `sed 's|^ui-vue-v||'`
- Pack prefix: `workfort-ui-vue-*.tgz`
- Release name: `Vue SDK`
- Package name in notes: `@workfort/ui-vue`

```yaml
# SPDX-License-Identifier: GPL-3.0-or-later
name: Release Vue Package

on:
  push:
    branches: [master]
    paths:
      - 'web/packages/vue/**'

permissions:
  contents: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: jdx/mise-action@v3

      - name: Install and build
        run: cd web && pnpm install && pnpm --filter @workfort/ui-vue build

      - name: Test
        run: cd web && pnpm --filter @workfort/ui-vue test

  tag:
    name: Create SDK Tag
    runs-on: ubuntu-latest
    needs: build
    outputs:
      new_tag: ${{ steps.tag_version.outputs.new_tag }}
      changelog: ${{ steps.tag_version.outputs.changelog }}
      should_release: ${{ steps.check_tag.outputs.should_release }}

    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Bump version and create tag
        id: tag_version
        uses: Work-Fort/github-tag-action@v6.3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          default_bump: false
          release_branches: master
          tag_prefix: ui-vue-v
          paths: web/packages/vue/**

      - name: Check if new tag was created
        id: check_tag
        run: |
          if [ -n "$TAG" ]; then
            echo "should_release=true" >> "$GITHUB_OUTPUT"
          else
            echo "should_release=false" >> "$GITHUB_OUTPUT"
          fi
        env:
          TAG: ${{ steps.tag_version.outputs.new_tag }}

  release:
    name: Publish to npm
    runs-on: ubuntu-latest
    needs: tag
    if: needs.tag.outputs.should_release == 'true'

    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ needs.tag.outputs.new_tag }}

      - uses: jdx/mise-action@v3

      - name: Build and publish
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-vue-v||')
          cd web
          pnpm install
          cd packages/vue
          npm pkg set version="$VERSION"
          pnpm run build
          npm publish --provenance --access public
          npm pack
          mv workfort-ui-vue-*.tgz ../../..

      - name: Create release notes
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
          CHANGELOG: ${{ needs.tag.outputs.changelog }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-vue-v||')
          cat > release-notes.md << EOF
          # @workfort/ui-vue v${VERSION}

          ## What's Changed

          ${CHANGELOG}

          ## Installation

          \`\`\`bash
          npm install @workfort/ui-vue@${VERSION}
          \`\`\`

          ---

          Built automatically by Scope CI
          EOF

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ needs.tag.outputs.new_tag }}
          name: Vue SDK ${{ needs.tag.outputs.new_tag }}
          body_path: release-notes.md
          draft: false
          prerelease: false
          files: workfort-ui-vue-*.tgz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Create `.github/workflows/release-svelte.yml`**

Same structure with substitutions:
- Name: `Release Svelte Package`
- Path: `web/packages/svelte/**`
- Tag prefix: `svelte-v`
- Filter: `@workfort/ui-svelte`
- Sed: `sed 's|^ui-svelte-v||'`
- Pack prefix: `workfort-ui-svelte-*.tgz`
- Release name: `Svelte SDK`
- Package name in notes: `@workfort/ui-svelte`

```yaml
# SPDX-License-Identifier: GPL-3.0-or-later
name: Release Svelte Package

on:
  push:
    branches: [master]
    paths:
      - 'web/packages/svelte/**'

permissions:
  contents: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: jdx/mise-action@v3

      - name: Install and build
        run: cd web && pnpm install && pnpm --filter @workfort/ui-svelte build

      - name: Test
        run: cd web && pnpm --filter @workfort/ui-svelte test

  tag:
    name: Create SDK Tag
    runs-on: ubuntu-latest
    needs: build
    outputs:
      new_tag: ${{ steps.tag_version.outputs.new_tag }}
      changelog: ${{ steps.tag_version.outputs.changelog }}
      should_release: ${{ steps.check_tag.outputs.should_release }}

    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Bump version and create tag
        id: tag_version
        uses: Work-Fort/github-tag-action@v6.3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          default_bump: false
          release_branches: master
          tag_prefix: ui-svelte-v
          paths: web/packages/svelte/**

      - name: Check if new tag was created
        id: check_tag
        run: |
          if [ -n "$TAG" ]; then
            echo "should_release=true" >> "$GITHUB_OUTPUT"
          else
            echo "should_release=false" >> "$GITHUB_OUTPUT"
          fi
        env:
          TAG: ${{ steps.tag_version.outputs.new_tag }}

  release:
    name: Publish to npm
    runs-on: ubuntu-latest
    needs: tag
    if: needs.tag.outputs.should_release == 'true'

    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ needs.tag.outputs.new_tag }}

      - uses: jdx/mise-action@v3

      - name: Build and publish
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-svelte-v||')
          cd web
          pnpm install
          cd packages/svelte
          npm pkg set version="$VERSION"
          pnpm run build
          npm publish --provenance --access public
          npm pack
          mv workfort-ui-svelte-*.tgz ../../..

      - name: Create release notes
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
          CHANGELOG: ${{ needs.tag.outputs.changelog }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-svelte-v||')
          cat > release-notes.md << EOF
          # @workfort/ui-svelte v${VERSION}

          ## What's Changed

          ${CHANGELOG}

          ## Installation

          \`\`\`bash
          npm install @workfort/ui-svelte@${VERSION}
          \`\`\`

          ---

          Built automatically by Scope CI
          EOF

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ needs.tag.outputs.new_tag }}
          name: Svelte SDK ${{ needs.tag.outputs.new_tag }}
          body_path: release-notes.md
          draft: false
          prerelease: false
          files: workfort-ui-svelte-*.tgz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 4: Create `.github/workflows/release-solid.yml`**

Same structure with substitutions:
- Name: `Release Solid Package`
- Path: `web/packages/solid/**`
- Tag prefix: `solid-v`
- Filter: `@workfort/ui-solid`
- Sed: `sed 's|^ui-solid-v||'`
- Pack prefix: `workfort-ui-solid-*.tgz`
- Release name: `Solid SDK`
- Package name in notes: `@workfort/ui-solid`

```yaml
# SPDX-License-Identifier: GPL-3.0-or-later
name: Release Solid Package

on:
  push:
    branches: [master]
    paths:
      - 'web/packages/solid/**'

permissions:
  contents: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: jdx/mise-action@v3

      - name: Install and build
        run: cd web && pnpm install && pnpm --filter @workfort/ui-solid build

      - name: Test
        run: cd web && pnpm --filter @workfort/ui-solid test

  tag:
    name: Create SDK Tag
    runs-on: ubuntu-latest
    needs: build
    outputs:
      new_tag: ${{ steps.tag_version.outputs.new_tag }}
      changelog: ${{ steps.tag_version.outputs.changelog }}
      should_release: ${{ steps.check_tag.outputs.should_release }}

    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Bump version and create tag
        id: tag_version
        uses: Work-Fort/github-tag-action@v6.3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          default_bump: false
          release_branches: master
          tag_prefix: ui-solid-v
          paths: web/packages/solid/**

      - name: Check if new tag was created
        id: check_tag
        run: |
          if [ -n "$TAG" ]; then
            echo "should_release=true" >> "$GITHUB_OUTPUT"
          else
            echo "should_release=false" >> "$GITHUB_OUTPUT"
          fi
        env:
          TAG: ${{ steps.tag_version.outputs.new_tag }}

  release:
    name: Publish to npm
    runs-on: ubuntu-latest
    needs: tag
    if: needs.tag.outputs.should_release == 'true'

    steps:
      - uses: actions/checkout@v6
        with:
          ref: ${{ needs.tag.outputs.new_tag }}

      - uses: jdx/mise-action@v3

      - name: Build and publish
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-solid-v||')
          cd web
          pnpm install
          cd packages/solid
          npm pkg set version="$VERSION"
          pnpm run build
          npm publish --provenance --access public
          npm pack
          mv workfort-ui-solid-*.tgz ../../..

      - name: Create release notes
        env:
          TAG: ${{ needs.tag.outputs.new_tag }}
          CHANGELOG: ${{ needs.tag.outputs.changelog }}
        run: |
          VERSION=$(echo "$TAG" | sed 's|^ui-solid-v||')
          cat > release-notes.md << EOF
          # @workfort/ui-solid v${VERSION}

          ## What's Changed

          ${CHANGELOG}

          ## Installation

          \`\`\`bash
          npm install @workfort/ui-solid@${VERSION}
          \`\`\`

          ---

          Built automatically by Scope CI
          EOF

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ needs.tag.outputs.new_tag }}
          name: Solid SDK ${{ needs.tag.outputs.new_tag }}
          body_path: release-notes.md
          draft: false
          prerelease: false
          files: workfort-ui-solid-*.tgz
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/release-react.yml .github/workflows/release-vue.yml .github/workflows/release-svelte.yml .github/workflows/release-solid.yml
git commit -m "feat(ci): add release workflows for React, Vue, Svelte, and Solid packages"
```

---

### Task 9: Delete old monolithic package

**Files:**
- Delete: `web/ui/` (entire directory)

- [ ] **Step 1: Verify all files have been moved**

Check that every source and test file from `web/ui/src/` and `web/ui/tests/` exists in its new location under `web/packages/`.

```bash
# Core
diff web/ui/src/index.ts web/packages/ui/src/index.ts
diff web/ui/src/base.ts web/packages/ui/src/base.ts
diff -r web/ui/src/components web/packages/ui/src/components
diff -r web/ui/src/styles web/packages/ui/src/styles
diff web/ui/tests/helpers.ts web/packages/ui/tests/helpers.ts
diff -r web/ui/tests/components web/packages/ui/tests/components

# React (components.tsx will differ due to import changes)
diff web/ui/src/react/use-auth.ts web/packages/react/src/use-auth.ts
diff web/ui/src/react/use-theme.ts web/packages/react/src/use-theme.ts

# Vue
diff -r web/ui/src/vue web/packages/vue/src

# Svelte
diff -r web/ui/src/svelte web/packages/svelte/src

# Solid
diff -r web/ui/src/solid web/packages/solid/src
```

Expected: All unchanged files are identical. `components.tsx` shows only the import path changes from Step 8 of Task 3.

- [ ] **Step 2: Delete `web/ui/`**

```bash
git rm -rf web/ui
```

- [ ] **Step 3: Run full workspace verification**

```bash
cd web && pnpm install && pnpm test && pnpm build
```

Expected: All 5 packages build and all tests pass.

- [ ] **Step 4: Commit**

```bash
git commit -m "chore: remove old monolithic web/ui package"
```

---

### Task 10: Commit documentation updates

The design spec amendment and plan superseded notices were written before execution began.

**Files already modified:**
- `docs/2026-03-12-workfort-ui-wc-design.md` — "Amendment: Package Split" section added
- `docs/plans/2026-03-12-workfort-ui-wc.md` — SUPERSEDED notice added

- [ ] **Step 1: Commit documentation**

```bash
git add docs/2026-03-12-workfort-ui-wc-design.md docs/plans/2026-03-12-workfort-ui-wc.md docs/plans/2026-03-12-ui-package-split.md
git commit -m "docs: add package split amendment to WC design spec, supersede old plan"
```

---

## Final Verification

After all tasks, from the repository root:

```bash
cd web && pnpm install && pnpm test && pnpm build
```

Expected output:
- `@workfort/ui`: 27 tests pass, `dist/index.js` + `dist/style.css` built
- `@workfort/ui-react`: 5 tests pass, `dist/index.js` built
- `@workfort/ui-vue`: 1 test passes, `dist/index.js` + `dist/index.d.ts` built
- `@workfort/ui-svelte`: 1 test passes, `dist/index.js` + `dist/index.d.ts` built
- `@workfort/ui-solid`: 1 test passes, `dist/index.js` + `dist/index.d.ts` built

Total: 35 tests pass, 5 packages build successfully.
