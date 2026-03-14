# Phase 0: Documentation Site & Storybook — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the `Work-Fort/Documentation` repo with Docusaurus (general developer docs) and Storybook (interactive component explorer), deployed to GitHub Pages.

**Architecture:** pnpm monorepo with two sites — Docusaurus at root serving general docs, and Storybook at `/design/` showing component demos. Storybook uses composition to present three framework renderers (Lit, SolidJS, React) in a unified interface. GitHub Actions builds both sites and deploys to GitHub Pages on push to master.

**Tech Stack:** Docusaurus 3, Storybook (latest stable, `@storybook/web-components-vite` for Lit, `storybook-solidjs-vite` for Solid, `@storybook/react-vite` for React), pnpm workspaces, GitHub Actions, GitHub Pages.

**Working directory:** This plan creates a NEW repo (`Work-Fort/Documentation`). All file paths below are relative to the Documentation repo root unless prefixed with `scope/lead/` (this repo).

**Prerequisites:** `gh auth status` must pass with permissions to create repos in the `Work-Fort` GitHub org.

---

## Component API Reference

All 14 existing `@workfort/ui` components and their properties/events:

| Component | Tag | Properties | Events |
|-----------|-----|-----------|--------|
| Panel | `wf-panel` | `label: string` | — |
| Button | `wf-button` | `variant: 'text'\|'filled'`, `disabled: boolean` | `wf-click` |
| Badge | `wf-badge` | `count: number` | — |
| StatusDot | `wf-status-dot` | `status: 'online'\|'offline'\|'away'` | — |
| Skeleton | `wf-skeleton` | `width: string`, `height: string` | — |
| Divider | `wf-divider` | — | — |
| TextInput | `wf-text-input` | `placeholder: string`, `value: string`, `disabled: boolean` | `wf-input`, `wf-change` |
| List | `wf-list` | — | — |
| ListItem | `wf-list-item` | `active: boolean` | `wf-select` |
| ScrollArea | `wf-scroll-area` | — | — |
| ErrorFallback | `wf-error-fallback` | `title: string`, `message: string` | — |
| Banner | `wf-banner` | `variant: 'error'\|'warning'\|'info'`, `dismissible: boolean`, `headline: string`, `details: string` | `wf-dismiss` |
| Toast | `wf-toast` | `variant: 'error'\|'warning'\|'info'\|'success'`, `sticky: boolean`, `duration: number` | `wf-dismiss` |
| ToastContainer | `wf-toast-container` | `position: 'top-right'\|'top-left'\|'bottom-right'\|'bottom-left'` | — |

## File Structure

```
Work-Fort/Documentation/
├── pnpm-workspace.yaml
├── package.json                         # Root — scripts for build-all, dev
├── .github/
│   └── workflows/
│       └── deploy.yml                   # Build Docusaurus + Storybook → GitHub Pages
├── .gitignore
│
├── docs/                                # Docusaurus content
│   ├── architecture.md
│   ├── service-contract.md
│   ├── shared-packages.md
│   ├── auth.md
│   ├── dev-workflow.md
│   └── getting-started/
│       ├── solidjs.md
│       ├── react.md
│       ├── vue.md
│       ├── svelte.md
│       └── web-components.md
├── docusaurus.config.ts
├── sidebars.ts
├── src/
│   ├── css/
│   │   └── custom.css                   # Docusaurus theme overrides
│   └── pages/
│       └── index.tsx                    # Landing page
├── static/
│   └── .nojekyll                        # Required for GitHub Pages
│
├── storybook/
│   ├── lit/                             # Lit/web-components Storybook (host)
│   │   ├── package.json
│   │   ├── .storybook/
│   │   │   ├── main.ts
│   │   │   └── preview.ts
│   │   └── stories/
│   │       ├── Button.stories.ts
│   │       ├── Panel.stories.ts
│   │       ├── Badge.stories.ts
│   │       ├── StatusDot.stories.ts
│   │       ├── Skeleton.stories.ts
│   │       ├── Divider.stories.ts
│   │       ├── TextInput.stories.ts
│   │       ├── List.stories.ts
│   │       ├── ErrorFallback.stories.ts
│   │       ├── Banner.stories.ts
│   │       ├── Toast.stories.ts
│   │       ├── ToastContainer.stories.ts
│   │       └── ScrollArea.stories.ts
│   │
│   ├── solid/                           # SolidJS Storybook (child)
│   │   ├── package.json
│   │   ├── .storybook/
│   │   │   ├── main.ts
│   │   │   └── preview.ts
│   │   └── stories/
│   │       ├── Button.stories.tsx
│   │       ├── Panel.stories.tsx
│   │       └── Auth.stories.tsx
│   │
│   └── react/                           # React Storybook (child)
│       ├── package.json
│       ├── .storybook/
│       │   ├── main.ts
│       │   └── preview.ts
│       └── stories/
│           ├── Button.stories.tsx
│           ├── Panel.stories.tsx
│           └── Auth.stories.tsx
│
└── tsconfig.json
```

---

## Chunk 1: Repository & Docusaurus

### Task 1: Create GitHub repo and initialize workspace

**Files:**
- Create: `pnpm-workspace.yaml`
- Create: `package.json`
- Create: `.gitignore`
- Create: `tsconfig.json`

- [ ] **Step 1: Create the GitHub repository**

```bash
# Create empty repo on GitHub (no initial commit, no default branch yet)
gh repo create Work-Fort/Documentation --public --description "WorkFort developer documentation and component explorer"

# Initialize locally with master branch
mkdir /home/kazw/Work/WorkFort/documentation && cd /home/kazw/Work/WorkFort/documentation
git init -b master
```

The repo is created empty on GitHub — no `main` branch ever exists. The first push (Task 1, Step 7) establishes `master` as the default branch.

- [ ] **Step 2: Create pnpm-workspace.yaml**

```yaml
packages:
  - 'storybook/*'
```

- [ ] **Step 3: Create root package.json**

```json
{
  "name": "workfort-documentation",
  "private": true,
  "type": "module",
  "scripts": {
    "docs:dev": "docusaurus start",
    "docs:build": "docusaurus build",
    "storybook:lit:dev": "pnpm --filter @workfort-docs/storybook-lit storybook dev -p 6006",
    "storybook:solid:dev": "pnpm --filter @workfort-docs/storybook-solid storybook dev -p 6007",
    "storybook:react:dev": "pnpm --filter @workfort-docs/storybook-react storybook dev -p 6008",
    "build": "pnpm docs:build && pnpm storybook:build",
    "storybook:build": "pnpm storybook:build:solid && pnpm storybook:build:react && pnpm storybook:build:lit && pnpm storybook:merge",
    "storybook:build:lit": "pnpm --filter @workfort-docs/storybook-lit storybook build -o ../../build/design",
    "storybook:build:solid": "pnpm --filter @workfort-docs/storybook-solid storybook build -o ../../build/design/solid",
    "storybook:build:react": "pnpm --filter @workfort-docs/storybook-react storybook build -o ../../build/design/react",
    "storybook:merge": "echo 'Storybook outputs merged into build/design/'"
  },
  "devDependencies": {
    "@docusaurus/core": "^3.7.0",
    "@docusaurus/preset-classic": "^3.7.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "typescript": "^5.7.0"
  }
}
```

Note: `storybook:build` builds children FIRST (solid, react), then the host (lit). This ensures child output directories exist when the host references them.

- [ ] **Step 4: Create .gitignore**

```
node_modules/
build/
.docusaurus/
storybook-static/
*.tsbuildinfo
.DS_Store
```

- [ ] **Step 5: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
```

- [ ] **Step 6: Install dependencies and verify**

```bash
pnpm install
```

Expected: lockfile created, no errors.

- [ ] **Step 7: Commit and push to establish master on GitHub**

```bash
git remote add origin git@github.com:Work-Fort/Documentation.git
git add -A
git commit -m "chore: initialize Documentation repo with pnpm workspace"
git push -u origin master
```

---

### Task 2: Set up Docusaurus

**Files:**
- Create: `docusaurus.config.ts`
- Create: `sidebars.ts`
- Create: `src/pages/index.tsx`
- Create: `src/css/custom.css`
- Create: `static/.nojekyll`

- [ ] **Step 1: Create docusaurus.config.ts**

```ts
import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'WorkFort Docs',
  tagline: 'Build services on the WorkFort platform',
  url: 'https://work-fort.github.io',
  baseUrl: '/Documentation/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',
  organizationName: 'Work-Fort',
  projectName: 'Documentation',
  trailingSlash: false,

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: '/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    navbar: {
      title: 'WorkFort',
      items: [
        { type: 'docSidebar', sidebarId: 'docs', position: 'left', label: 'Docs' },
        { href: '/Documentation/design/', label: 'Components', position: 'left' },
        { href: 'https://github.com/Work-Fort', label: 'GitHub', position: 'right' },
      ],
    },
    footer: {
      style: 'dark',
      copyright: `Copyright © ${new Date().getFullYear()} WorkFort.`,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
```

Note: `routeBasePath: '/'` makes docs the root. `baseUrl: '/Documentation/'` is the GitHub Pages project site path. The "Components" nav link points to Storybook at `/Documentation/design/`.

- [ ] **Step 2: Create sidebars.ts**

```ts
import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'architecture',
    'service-contract',
    'shared-packages',
    'auth',
    'dev-workflow',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'getting-started/solidjs',
        'getting-started/react',
        'getting-started/vue',
        'getting-started/svelte',
        'getting-started/web-components',
      ],
    },
  ],
};

export default sidebars;
```

- [ ] **Step 3: Create landing page**

```tsx
// src/pages/index.tsx
import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';

export default function Home(): React.JSX.Element {
  return (
    <Layout title="Home" description="WorkFort developer documentation">
      <main style={{ padding: '2rem', maxWidth: '800px', margin: '0 auto' }}>
        <h1>WorkFort Documentation</h1>
        <p>Build services on the WorkFort platform.</p>
        <div style={{ display: 'flex', gap: '1rem', marginTop: '1.5rem' }}>
          <Link className="button button--primary button--lg" to="/architecture">
            Read the Docs
          </Link>
          <Link className="button button--secondary button--lg" href="/Documentation/design/">
            Browse Components
          </Link>
        </div>
      </main>
    </Layout>
  );
}
```

- [ ] **Step 4: Create custom.css**

```css
/* src/css/custom.css */
:root {
  --ifm-color-primary: #1c1917;
  --ifm-color-primary-dark: #0c0a09;
  --ifm-color-primary-darker: #000000;
  --ifm-color-primary-darkest: #000000;
  --ifm-color-primary-light: #292524;
  --ifm-color-primary-lighter: #44403c;
  --ifm-color-primary-lightest: #57534e;
  --ifm-font-family-base: ui-sans-serif, system-ui, -apple-system, sans-serif;
  --ifm-font-family-monospace: ui-monospace, 'SF Mono', 'Fira Code', monospace;
}

[data-theme='dark'] {
  --ifm-color-primary: #e7e5e4;
  --ifm-color-primary-dark: #d6d3d1;
  --ifm-color-primary-darker: #a8a29e;
  --ifm-color-primary-darkest: #78716c;
  --ifm-color-primary-light: #fafaf9;
  --ifm-color-primary-lighter: #ffffff;
  --ifm-color-primary-lightest: #ffffff;
}
```

- [ ] **Step 5: Create .nojekyll**

```bash
touch static/.nojekyll
```

This prevents GitHub Pages from processing the site with Jekyll, which would skip files starting with `_`.

- [ ] **Step 6: Verify Docusaurus builds**

```bash
pnpm docs:build
```

Expected: Build succeeds, outputs to `build/`. May warn about missing docs (added in Task 3).

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: set up Docusaurus with landing page and sidebar config"
```

---

### Task 3: Transfer frontend docs

**Files:**
- Create: `docs/architecture.md` (copy from `scope/lead/docs/frontend/architecture.md`)
- Create: `docs/service-contract.md`
- Create: `docs/shared-packages.md`
- Create: `docs/auth.md`
- Create: `docs/dev-workflow.md`
- Create: `docs/getting-started/solidjs.md`
- Create: `docs/getting-started/react.md`
- Create: `docs/getting-started/vue.md`
- Create: `docs/getting-started/svelte.md`
- Create: `docs/getting-started/web-components.md`

- [ ] **Step 1: Copy docs from scope/lead repo**

```bash
# From the Documentation repo root:
cp -r /home/kazw/Work/WorkFort/scope/lead/docs/frontend/architecture.md docs/
cp -r /home/kazw/Work/WorkFort/scope/lead/docs/frontend/service-contract.md docs/
cp -r /home/kazw/Work/WorkFort/scope/lead/docs/frontend/shared-packages.md docs/
cp -r /home/kazw/Work/WorkFort/scope/lead/docs/frontend/auth.md docs/
cp -r /home/kazw/Work/WorkFort/scope/lead/docs/frontend/dev-workflow.md docs/
mkdir -p docs/getting-started
cp -r /home/kazw/Work/WorkFort/scope/lead/docs/frontend/getting-started/*.md docs/getting-started/
```

Note: Do NOT copy `docs/frontend/README.md` — its role is replaced by the Docusaurus sidebar navigation.

- [ ] **Step 2: Add Docusaurus frontmatter to each doc**

Each markdown file needs a frontmatter block at the top for Docusaurus to process it. Add to the top of each file:

For `docs/architecture.md`:
```yaml
---
sidebar_position: 1
title: Architecture
---
```

For `docs/service-contract.md`:
```yaml
---
sidebar_position: 2
title: Service Frontend Contract
---
```

For `docs/shared-packages.md`:
```yaml
---
sidebar_position: 3
title: Shared Packages
---
```

For `docs/auth.md`:
```yaml
---
sidebar_position: 4
title: Authentication
---
```

For `docs/dev-workflow.md`:
```yaml
---
sidebar_position: 5
title: Development Workflow
---
```

For `docs/getting-started/solidjs.md`:
```yaml
---
sidebar_position: 1
title: "Getting Started: SolidJS"
---
```

For `docs/getting-started/react.md`:
```yaml
---
sidebar_position: 2
title: "Getting Started: React"
---
```

For `docs/getting-started/vue.md`:
```yaml
---
sidebar_position: 3
title: "Getting Started: Vue"
---
```

For `docs/getting-started/svelte.md`:
```yaml
---
sidebar_position: 4
title: "Getting Started: Svelte"
---
```

For `docs/getting-started/web-components.md`:
```yaml
---
sidebar_position: 5
title: "Getting Started: Web Components"
---
```

- [ ] **Step 3: Fix internal doc links**

The transferred docs use relative markdown links (e.g., `[Service Frontend Contract](../service-contract.md)`). Docusaurus resolves these differently. Update all internal links to use Docusaurus path format:

- `../service-contract.md` → `/service-contract`
- `../shared-packages.md` → `/shared-packages`
- `../auth.md` → `/auth`
- `../architecture.md` → `/architecture`
- `./solidjs.md` → `/getting-started/solidjs`
- `./react.md` → `/getting-started/react`
- etc.

Search each file for `](` patterns and update the paths.

- [ ] **Step 4: Verify Docusaurus builds with docs**

```bash
pnpm docs:build
```

Expected: Build succeeds with all 10 docs rendered. No broken link warnings.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "docs: transfer frontend documentation from scope/lead"
```

---

## Chunk 2: Lit Storybook

### Task 4: Set up Lit Storybook (host instance)

**Files:**
- Create: `storybook/lit/package.json`
- Create: `storybook/lit/.storybook/main.ts`
- Create: `storybook/lit/.storybook/preview.ts`

- [ ] **Step 1: Create storybook/lit/package.json**

```json
{
  "name": "@workfort-docs/storybook-lit",
  "private": true,
  "type": "module",
  "scripts": {
    "storybook": "storybook dev -p 6006",
    "build-storybook": "storybook build"
  },
  "dependencies": {
    "@workfort/ui": "latest",
    "lit": "^3.2.0"
  },
  "devDependencies": {
    "@storybook/web-components-vite": "latest",
    "storybook": "latest",
    "vite": "^6.0.0"
  }
}
```

- [ ] **Step 2: Create storybook/lit/.storybook/main.ts**

```ts
import type { StorybookConfig } from '@storybook/web-components-vite';

const config: StorybookConfig = {
  stories: ['../stories/**/*.stories.ts'],
  framework: '@storybook/web-components-vite',
  addons: ['@storybook/addon-essentials'],
  refs: {
    solid: {
      title: 'SolidJS',
      url: './solid/',
    },
    react: {
      title: 'React',
      url: './react/',
    },
  },
  viteFinal: async (config) => {
    return config;
  },
};

export default config;
```

The `refs` enable Storybook Composition — the Lit instance is the host and embeds the Solid and React instances in its sidebar. During local dev, update `url` to `http://localhost:6007` and `http://localhost:6008` respectively (or remove `refs` for standalone dev).

- [ ] **Step 3: Create storybook/lit/.storybook/preview.ts**

```ts
import '@workfort/ui/style.css';
import '@workfort/ui';

import type { Preview } from '@storybook/web-components';

const preview: Preview = {
  parameters: {
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#0c0a09' },
        { name: 'light', value: '#fafaf9' },
      ],
    },
  },
  decorators: [
    (story) => {
      // Ensure wf-* tokens are available
      document.documentElement.removeAttribute('data-theme');
      return story();
    },
  ],
};

export default preview;
```

- [ ] **Step 4: Install dependencies and verify Storybook starts**

```bash
pnpm install
pnpm --filter @workfort-docs/storybook-lit storybook dev -p 6006 --no-open
```

Expected: Storybook dev server starts on port 6006. Will show "No stories found" — that's expected before Task 5.

Press Ctrl+C to stop.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: set up Lit Storybook with composition refs"
```

---

### Task 5: Write Lit stories for all 14 components

**Files:**
- Create: `storybook/lit/stories/Button.stories.ts`
- Create: `storybook/lit/stories/Panel.stories.ts`
- Create: `storybook/lit/stories/Badge.stories.ts`
- Create: `storybook/lit/stories/StatusDot.stories.ts`
- Create: `storybook/lit/stories/Skeleton.stories.ts`
- Create: `storybook/lit/stories/Divider.stories.ts`
- Create: `storybook/lit/stories/TextInput.stories.ts`
- Create: `storybook/lit/stories/List.stories.ts`
- Create: `storybook/lit/stories/ErrorFallback.stories.ts`
- Create: `storybook/lit/stories/Banner.stories.ts`
- Create: `storybook/lit/stories/Toast.stories.ts`
- Create: `storybook/lit/stories/ToastContainer.stories.ts`
- Create: `storybook/lit/stories/ScrollArea.stories.ts`

Note: ListItem is covered within `List.stories.ts` since it is always used as a child of List. ScrollArea gets its own story file.

- [ ] **Step 1: Create Button.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Button',
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['text', 'filled'],
      description: 'Visual variant',
    },
    disabled: {
      control: 'boolean',
      description: 'Disables the button',
    },
    label: {
      control: 'text',
      description: 'Button text content',
    },
  },
  args: {
    variant: 'text',
    disabled: false,
    label: 'Click me',
  },
  render: (args) => html`
    <wf-button
      variant=${args.variant}
      ?disabled=${args.disabled}
      @wf-click=${() => console.log('wf-click fired')}
    >${args.label}</wf-button>
  `,
};

export default meta;
type Story = StoryObj;

export const Text: Story = {
  args: { variant: 'text', label: 'Text Button' },
};

export const Filled: Story = {
  args: { variant: 'filled', label: 'Filled Button' },
};

export const Disabled: Story = {
  args: { variant: 'filled', disabled: true, label: 'Disabled' },
};
```

- [ ] **Step 2: Create Panel.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Panel',
  tags: ['autodocs'],
  argTypes: {
    label: { control: 'text', description: 'Panel header label' },
  },
  args: { label: 'Panel Title' },
  render: (args) => html`
    <wf-panel label=${args.label}>
      <div style="padding: 1rem;">Panel content goes here.</div>
    </wf-panel>
  `,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};

export const NoLabel: Story = {
  args: { label: '' },
};
```

- [ ] **Step 3: Create Badge.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Badge',
  tags: ['autodocs'],
  argTypes: {
    count: { control: 'number', description: 'Badge count (hidden when 0)' },
  },
  args: { count: 5 },
  render: (args) => html`<wf-badge count=${args.count}></wf-badge>`,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
export const Zero: Story = { args: { count: 0 } };
export const Large: Story = { args: { count: 99 } };
```

- [ ] **Step 4: Create StatusDot.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/StatusDot',
  tags: ['autodocs'],
  argTypes: {
    status: { control: 'select', options: ['online', 'offline', 'away'] },
  },
  args: { status: 'online' },
  render: (args) => html`<wf-status-dot status=${args.status}></wf-status-dot>`,
};

export default meta;
type Story = StoryObj;

export const Online: Story = { args: { status: 'online' } };
export const Offline: Story = { args: { status: 'offline' } };
export const Away: Story = { args: { status: 'away' } };
```

- [ ] **Step 5: Create Skeleton.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Skeleton',
  tags: ['autodocs'],
  argTypes: {
    width: { control: 'text' },
    height: { control: 'text' },
  },
  args: { width: '200px', height: '1em' },
  render: (args) => html`<wf-skeleton width=${args.width} height=${args.height}></wf-skeleton>`,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
export const Block: Story = { args: { width: '100%', height: '80px' } };
export const Circle: Story = { args: { width: '40px', height: '40px' } };
```

- [ ] **Step 6: Create Divider.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Divider',
  tags: ['autodocs'],
  render: () => html`
    <div>
      <p>Content above</p>
      <wf-divider></wf-divider>
      <p>Content below</p>
    </div>
  `,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
```

- [ ] **Step 7: Create TextInput.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/TextInput',
  tags: ['autodocs'],
  argTypes: {
    placeholder: { control: 'text' },
    value: { control: 'text' },
    disabled: { control: 'boolean' },
  },
  args: { placeholder: 'Type something...', value: '', disabled: false },
  render: (args) => html`
    <wf-text-input
      placeholder=${args.placeholder}
      value=${args.value}
      ?disabled=${args.disabled}
      @wf-input=${(e: CustomEvent) => console.log('wf-input:', e.detail.value)}
      @wf-change=${(e: CustomEvent) => console.log('wf-change:', e.detail.value)}
    ></wf-text-input>
  `,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
export const WithValue: Story = { args: { value: 'Hello' } };
export const Disabled: Story = { args: { disabled: true, value: 'Cannot edit' } };
```

- [ ] **Step 8: Create List.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/List',
  tags: ['autodocs'],
  render: () => html`
    <wf-list>
      <wf-list-item>First item</wf-list-item>
      <wf-list-item active>Active item</wf-list-item>
      <wf-list-item>Third item</wf-list-item>
    </wf-list>
  `,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};

export const WithTrailing: Story = {
  render: () => html`
    <wf-list>
      <wf-list-item>
        Item with badge
        <wf-badge data-wf="trailing" count=${3}></wf-badge>
      </wf-list-item>
      <wf-list-item>
        Item with status
        <wf-status-dot data-wf="trailing" status="online"></wf-status-dot>
      </wf-list-item>
    </wf-list>
  `,
};
```

- [ ] **Step 9: Create ErrorFallback.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/ErrorFallback',
  tags: ['autodocs'],
  argTypes: {
    title: { control: 'text' },
    message: { control: 'text' },
  },
  args: { title: 'Something went wrong', message: 'Please try again later.' },
  render: (args) => html`
    <wf-error-fallback title=${args.title} message=${args.message}></wf-error-fallback>
  `,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
export const TitleOnly: Story = { args: { message: '' } };
```

- [ ] **Step 10: Create Banner.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Banner',
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['error', 'warning', 'info'] },
    headline: { control: 'text' },
    details: { control: 'text' },
    dismissible: { control: 'boolean' },
  },
  args: {
    variant: 'info',
    headline: 'Heads up',
    details: 'This is additional detail text that can be expanded.',
    dismissible: false,
  },
  render: (args) => html`
    <wf-banner
      variant=${args.variant}
      headline=${args.headline}
      details=${args.details}
      ?dismissible=${args.dismissible}
      @wf-dismiss=${() => console.log('wf-dismiss fired')}
    ></wf-banner>
  `,
};

export default meta;
type Story = StoryObj;

export const Info: Story = {};
export const Warning: Story = { args: { variant: 'warning', headline: 'Warning' } };
export const Error: Story = { args: { variant: 'error', headline: 'Error occurred' } };
export const Dismissible: Story = { args: { dismissible: true } };
```

- [ ] **Step 11: Create Toast.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/Toast',
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['info', 'success', 'warning', 'error'] },
    sticky: { control: 'boolean' },
    duration: { control: 'number' },
    label: { control: 'text' },
  },
  args: { variant: 'info', sticky: true, duration: 5000, label: 'Toast message' },
  render: (args) => html`
    <wf-toast
      variant=${args.variant}
      ?sticky=${args.sticky}
      duration=${args.duration}
    >${args.label}</wf-toast>
  `,
};

export default meta;
type Story = StoryObj;

export const Info: Story = {};
export const Success: Story = { args: { variant: 'success', label: 'Saved!' } };
export const Warning: Story = { args: { variant: 'warning', label: 'Check your input' } };
export const Error: Story = { args: { variant: 'error', label: 'Something failed' } };
```

Note: Stories use `sticky: true` to prevent toasts from auto-dismissing during Storybook interaction.

- [ ] **Step 12: Create ScrollArea.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/ScrollArea',
  tags: ['autodocs'],
  render: () => html`
    <wf-scroll-area style="height: 150px; border: 1px solid var(--wf-border);">
      <div style="padding: 1rem;">
        ${Array.from({ length: 20 }, (_, i) => html`<p>Scrollable line ${i + 1}</p>`)}
      </div>
    </wf-scroll-area>
  `,
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
```

- [ ] **Step 13: Create ToastContainer.stories.ts**

```ts
import { html } from 'lit';
import type { Meta, StoryObj } from '@storybook/web-components';
import '@workfort/ui';

const meta: Meta = {
  title: 'Components/ToastContainer',
  tags: ['autodocs'],
  argTypes: {
    position: {
      control: 'select',
      options: ['top-right', 'top-left', 'bottom-right', 'bottom-left'],
    },
  },
  args: { position: 'top-right' },
  render: (args) => html`
    <div style="position: relative; height: 200px; border: 1px dashed var(--wf-border);">
      <wf-toast-container position=${args.position}>
        <wf-toast variant="info" sticky>Notification</wf-toast>
      </wf-toast-container>
    </div>
  `,
};

export default meta;
type Story = StoryObj;

export const TopRight: Story = {};
export const BottomLeft: Story = { args: { position: 'bottom-left' } };
```

- [ ] **Step 14: Verify Storybook builds with all stories**

```bash
pnpm --filter @workfort-docs/storybook-lit storybook build -o ../../build/design --quiet
```

Expected: Build succeeds. All 13 story files (covering 14 components) produce rendered stories.

- [ ] **Step 15: Commit**

```bash
git add -A
git commit -m "feat: add Lit stories for all 14 components"
```

---

## Chunk 3: Framework Storybooks

### Task 6: Set up SolidJS Storybook + stories

**Files:**
- Create: `storybook/solid/package.json`
- Create: `storybook/solid/.storybook/main.ts`
- Create: `storybook/solid/.storybook/preview.ts`
- Create: `storybook/solid/stories/Button.stories.tsx`
- Create: `storybook/solid/stories/Panel.stories.tsx`
- Create: `storybook/solid/stories/Auth.stories.tsx`

SolidJS uses `wf-*` custom elements natively — no wrappers needed. These stories show how Sharkfin (SolidJS) developers use the components.

- [ ] **Step 1: Create storybook/solid/package.json**

```json
{
  "name": "@workfort-docs/storybook-solid",
  "private": true,
  "type": "module",
  "scripts": {
    "storybook": "storybook dev -p 6007",
    "build-storybook": "storybook build"
  },
  "dependencies": {
    "@workfort/ui": "latest",
    "@workfort/ui-solid": "latest",
    "solid-js": "^1.9.0"
  },
  "devDependencies": {
    "storybook": "latest",
    "storybook-solidjs-vite": "latest",
    "vite": "^6.0.0",
    "vite-plugin-solid": "^2.11.0"
  }
}
```

- [ ] **Step 2: Create storybook/solid/.storybook/main.ts**

```ts
import type { StorybookConfig } from 'storybook-solidjs-vite';

const config: StorybookConfig = {
  stories: ['../stories/**/*.stories.tsx'],
  framework: 'storybook-solidjs-vite',
  addons: ['@storybook/addon-essentials'],
};

export default config;
```

- [ ] **Step 3: Create storybook/solid/.storybook/preview.ts**

```ts
import '@workfort/ui/style.css';
import '@workfort/ui';

import type { Preview } from 'storybook-solidjs';

const preview: Preview = {
  parameters: {
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#0c0a09' },
        { name: 'light', value: '#fafaf9' },
      ],
    },
  },
};

export default preview;
```

- [ ] **Step 4: Create SolidJS Button story**

```tsx
// storybook/solid/stories/Button.stories.tsx
import type { Meta, StoryObj } from 'storybook-solidjs';
import '@workfort/ui';

const meta: Meta = {
  title: 'SolidJS/Button',
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['text', 'filled'] },
    disabled: { control: 'boolean' },
    label: { control: 'text' },
  },
  args: { variant: 'text', disabled: false, label: 'Click me' },
  render: (props) => (
    <wf-button
      variant={props.variant}
      disabled={props.disabled || undefined}
      on:wf-click={() => console.log('wf-click')}
    >
      {props.label}
    </wf-button>
  ),
};

export default meta;
type Story = StoryObj;

export const Text: Story = { args: { variant: 'text' } };
export const Filled: Story = { args: { variant: 'filled' } };
export const Disabled: Story = { args: { disabled: true } };
```

Note: SolidJS uses `on:wf-click` syntax for custom element events, and passes `undefined` (not `false`) to remove boolean attributes.

- [ ] **Step 5: Create SolidJS Panel story**

```tsx
// storybook/solid/stories/Panel.stories.tsx
import type { Meta, StoryObj } from 'storybook-solidjs';
import '@workfort/ui';

const meta: Meta = {
  title: 'SolidJS/Panel',
  tags: ['autodocs'],
  argTypes: {
    label: { control: 'text' },
  },
  args: { label: 'Panel Title' },
  render: (props) => (
    <wf-panel label={props.label}>
      <div style={{ padding: '1rem' }}>SolidJS content inside a panel.</div>
    </wf-panel>
  ),
};

export default meta;
type Story = StoryObj;

export const Default: Story = {};
```

- [ ] **Step 6: Create SolidJS Auth story**

```tsx
// storybook/solid/stories/Auth.stories.tsx
import { createSignal } from 'solid-js';
import type { Meta, StoryObj } from 'storybook-solidjs';

const meta: Meta = {
  title: 'SolidJS/Auth Hook',
  tags: ['autodocs'],
  render: () => {
    // Demonstrates the useAuth pattern from @workfort/ui-solid.
    // In a real app: import { useAuth } from '@workfort/ui-solid';
    // Here we mock it for the story.
    const [user] = createSignal({ displayName: 'Demo User', username: 'demo@example.com' });
    const [isAuthenticated] = createSignal(true);

    return (
      <wf-panel label="Auth State">
        <div style={{ padding: '1rem' }}>
          <p>Authenticated: <strong>{String(isAuthenticated())}</strong></p>
          <p>User: <strong>{user()?.displayName}</strong></p>
          <p>Email: {user()?.username}</p>
          <hr />
          <p style={{ color: 'var(--wf-text-muted)', 'font-size': '0.8rem' }}>
            In your app, use <code>useAuth()</code> from <code>@workfort/ui-solid</code> for reactive auth state.
          </p>
        </div>
      </wf-panel>
    );
  },
};

export default meta;
type Story = StoryObj;

export const AuthenticatedUser: Story = {};
```

- [ ] **Step 7: Install and verify SolidJS Storybook**

```bash
pnpm install
pnpm --filter @workfort-docs/storybook-solid storybook build -o ../../build/design/solid --quiet
```

Expected: Build succeeds. Three stories rendered.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add SolidJS Storybook with component and auth stories"
```

---

### Task 7: Set up React Storybook + stories

**Files:**
- Create: `storybook/react/package.json`
- Create: `storybook/react/.storybook/main.ts`
- Create: `storybook/react/.storybook/preview.ts`
- Create: `storybook/react/stories/Button.stories.tsx`
- Create: `storybook/react/stories/Panel.stories.tsx`
- Create: `storybook/react/stories/Auth.stories.tsx`

React stories use the wrapper components from `@workfort/ui-react`.

- [ ] **Step 1: Create storybook/react/package.json**

```json
{
  "name": "@workfort-docs/storybook-react",
  "private": true,
  "type": "module",
  "scripts": {
    "storybook": "storybook dev -p 6008",
    "build-storybook": "storybook build"
  },
  "dependencies": {
    "@workfort/ui": "latest",
    "@workfort/ui-react": "latest",
    "react": "^19.0.0",
    "react-dom": "^19.0.0"
  },
  "devDependencies": {
    "@storybook/react-vite": "latest",
    "storybook": "latest",
    "vite": "^6.0.0"
  }
}
```

- [ ] **Step 2: Create storybook/react/.storybook/main.ts**

```ts
import type { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: ['../stories/**/*.stories.tsx'],
  framework: '@storybook/react-vite',
  addons: ['@storybook/addon-essentials'],
};

export default config;
```

- [ ] **Step 3: Create storybook/react/.storybook/preview.ts**

```ts
import '@workfort/ui/style.css';
import '@workfort/ui';

import type { Preview } from '@storybook/react';

const preview: Preview = {
  parameters: {
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#0c0a09' },
        { name: 'light', value: '#fafaf9' },
      ],
    },
  },
};

export default preview;
```

- [ ] **Step 4: Create React Button story**

```tsx
// storybook/react/stories/Button.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from '@workfort/ui-react';

const meta: Meta<typeof Button> = {
  title: 'React/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['text', 'filled'] },
    disabled: { control: 'boolean' },
  },
  args: { variant: 'text', disabled: false },
  render: (args) => (
    <Button
      variant={args.variant}
      disabled={args.disabled}
      onWfClick={() => console.log('wf-click')}
    >
      {args.children ?? 'Click me'}
    </Button>
  ),
};

export default meta;
type Story = StoryObj<typeof Button>;

export const Text: Story = { args: { variant: 'text', children: 'Text Button' } };
export const Filled: Story = { args: { variant: 'filled', children: 'Filled Button' } };
export const Disabled: Story = { args: { disabled: true, children: 'Disabled' } };
```

Note: React wrappers convert `onWfClick` to `addEventListener('wf-click', ...)` via the `useWcEvents` hook in `@workfort/ui-react`.

- [ ] **Step 5: Create React Panel story**

```tsx
// storybook/react/stories/Panel.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Panel } from '@workfort/ui-react';

const meta: Meta<typeof Panel> = {
  title: 'React/Panel',
  component: Panel,
  tags: ['autodocs'],
  argTypes: {
    label: { control: 'text' },
  },
  args: { label: 'Panel Title' },
  render: (args) => (
    <Panel label={args.label}>
      <div style={{ padding: '1rem' }}>React content inside a panel.</div>
    </Panel>
  ),
};

export default meta;
type Story = StoryObj<typeof Panel>;

export const Default: Story = {};
```

- [ ] **Step 6: Create React Auth story**

```tsx
// storybook/react/stories/Auth.stories.tsx
import React, { useState } from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { Panel } from '@workfort/ui-react';

const meta: Meta = {
  title: 'React/Auth Hook',
  tags: ['autodocs'],
  render: () => {
    // Demonstrates the useAuth pattern from @workfort/ui-react.
    // In a real app: import { useAuth } from '@workfort/ui-react';
    // Here we mock it for the story.
    const [user] = useState({ displayName: 'Demo User', username: 'demo@example.com' });
    const [isAuthenticated] = useState(true);

    return (
      <Panel label="Auth State">
        <div style={{ padding: '1rem' }}>
          <p>Authenticated: <strong>{String(isAuthenticated)}</strong></p>
          <p>User: <strong>{user?.displayName}</strong></p>
          <p>Email: {user?.username}</p>
          <hr />
          <p style={{ color: 'var(--wf-text-muted)', fontSize: '0.8rem' }}>
            In your app, use <code>useAuth()</code> from <code>@workfort/ui-react</code> for reactive auth state.
          </p>
        </div>
      </Panel>
    );
  },
};

export default meta;
type Story = StoryObj;

export const AuthenticatedUser: Story = {};
```

- [ ] **Step 7: Install and verify React Storybook**

```bash
pnpm install
pnpm --filter @workfort-docs/storybook-react storybook build -o ../../build/design/react --quiet
```

Expected: Build succeeds. Three stories rendered.

- [ ] **Step 8: Commit**

```bash
git add -A
git commit -m "feat: add React Storybook with wrapper component and auth stories"
```

---

## Chunk 4: Deployment

### Task 8: Configure Storybook composition and verify full build

**Files:**
- Modify: `storybook/lit/.storybook/main.ts` (already has refs — verify they work)

- [ ] **Step 1: Build the full site (Docusaurus first, then Storybook)**

Build order: Docusaurus first (creates `build/`), then Storybook (writes into `build/design/`). This ensures Storybook output is not overwritten by Docusaurus.

```bash
rm -rf build/
pnpm build
```

This runs the root `build` script: `pnpm docs:build && pnpm storybook:build`.

Expected output structure:
```bash
ls build/           # Docusaurus files + design/
ls build/design/    # Lit Storybook files + solid/ + react/
```

- [ ] **Step 4: Test locally with a static server**

```bash
npx serve build/ -p 3000
```

Open `http://localhost:3000/Documentation/` — should show Docusaurus landing page.
Open `http://localhost:3000/Documentation/design/` — should show Lit Storybook with SolidJS and React tabs in sidebar.

Press Ctrl+C to stop.

- [ ] **Step 5: Commit if any build script adjustments were needed**

```bash
git add -A
git commit -m "fix: adjust build order for Docusaurus + Storybook coexistence"
```

---

### Task 9: GitHub Actions deployment workflow

**Files:**
- Create: `.github/workflows/deploy.yml`

- [ ] **Step 1: Create the deployment workflow**

```yaml
# .github/workflows/deploy.yml
name: Deploy to GitHub Pages

on:
  push:
    branches: [master]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: pnpm/action-setup@v4
        with:
          version: 9

      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm

      - run: pnpm install --frozen-lockfile

      # Build Docusaurus first
      - run: pnpm docs:build

      # Build Storybook (children first, then host)
      - run: pnpm storybook:build

      - uses: actions/upload-pages-artifact@v3
        with:
          path: build

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    needs: build
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: Enable GitHub Pages in repo settings**

This must be done manually (or via `gh` CLI):

```bash
gh api repos/Work-Fort/Documentation/pages \
  --method POST \
  --field build_type=workflow \
  2>/dev/null || echo "Pages may already be configured"
```

- [ ] **Step 3: Commit and push**

```bash
git add -A
git commit -m "ci: add GitHub Pages deployment workflow"
git push origin master
```

- [ ] **Step 4: Verify deployment**

```bash
gh run watch --repo Work-Fort/Documentation
```

Expected: Workflow runs successfully. Site is accessible at `https://work-fort.github.io/Documentation/`.

- [ ] **Step 5: Verify all URLs work**

| URL | Expected |
|-----|----------|
| `https://work-fort.github.io/Documentation/` | Docusaurus landing page |
| `https://work-fort.github.io/Documentation/architecture` | Architecture doc |
| `https://work-fort.github.io/Documentation/getting-started/solidjs` | SolidJS guide |
| `https://work-fort.github.io/Documentation/design/` | Lit Storybook |
| `https://work-fort.github.io/Documentation/design/?path=/docs/components-button--docs` | Button docs |

---

### Task 10: Final verification and cleanup

- [ ] **Step 1: Verify deployment succeeded**

```bash
gh run list --repo Work-Fort/Documentation --limit 1 --json status,conclusion
```

Expected: `{"status": "completed", "conclusion": "success"}`.

- [ ] **Step 2: Verify key pages return 200**

```bash
BASE="https://work-fort.github.io/Documentation"
for path in "/" "/architecture" "/getting-started/solidjs" "/design/" "/design/solid/" "/design/react/"; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${BASE}${path}")
  echo "${path} → ${STATUS}"
done
```

Expected: All paths return `200`.

- [ ] **Step 3: Verify Storybook stories are present in build output**

```bash
# Check that story entries exist in the Lit Storybook build
grep -l "Components/Button" build/design/*.js build/design/sb-preview/*.js 2>/dev/null | head -1
```

Expected: At least one JS file contains story references. If this fails, the stories didn't compile correctly.

- [ ] **Step 4: Manual smoke test (if running interactively)**

Open these URLs and spot-check:
- `https://work-fort.github.io/Documentation/` — Docusaurus landing
- `https://work-fort.github.io/Documentation/design/` — Storybook with Components sidebar
- Click "Components" in the Docusaurus navbar — should navigate to Storybook

- [ ] **Step 5: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "fix: final adjustments from deployment verification"
```
