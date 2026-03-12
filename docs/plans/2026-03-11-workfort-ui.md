# `@workfort/ui` — Headless Component Library Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a headless SolidJS component library themed via CSS custom properties, published as `@workfort/ui` for type safety. At runtime, Module Federation resolves it from the shell as a singleton.

**Architecture:** Three sub-modules — `@workfort/ui` (structural components), `@workfort/ui/theme` (theme context and tokens), `@workfort/ui/auth` (auth context from better-auth). All components consume `--wf-*` CSS variables set by the shell. Components are headless: they define structure and behavior, not visual style.

**Tech Stack:** SolidJS, TypeScript, Vite (library mode), Vitest + `@solidjs/testing-library`, UnoCSS (Tailwind preset)

**Specs:**
- `docs/2026-03-11-web-ui-design.md` — Primary spec (component inventory, theme tokens, auth flow)
- `docs/2026-03-11-service-auth-design.md` — Auth design (session, identity model)

---

## File Structure

```
# Root workspace (new)
pnpm-workspace.yaml          # Workspace: packages/*, web
package.json                  # Root workspace package.json (private)

packages/ui/                  # @workfort/ui npm package
├── package.json              # name: @workfort/ui, exports map
├── tsconfig.json
├── vite.config.ts            # Library mode build
├── vitest.config.ts
├── src/
│   ├── index.ts              # Re-exports all components
│   │
│   ├── theme/                # @workfort/ui/theme sub-module
│   │   ├── index.ts          # Re-exports: ThemeProvider, useTheme, tokens
│   │   ├── tokens.ts         # Typed CSS variable name constants
│   │   ├── provider.tsx      # ThemeProvider component (dark/light, localStorage)
│   │   └── context.ts        # Theme context, useTheme() accessor
│   │
│   ├── auth/                 # @workfort/ui/auth sub-module
│   │   ├── index.ts          # Re-exports: AuthProvider, useAuth, RequireAuth
│   │   ├── provider.tsx      # AuthProvider (wraps better-auth client)
│   │   ├── context.ts        # Auth context, useAuth() accessor
│   │   └── require-auth.tsx  # RequireAuth guard component
│   │
│   ├── panel.tsx             # Panel — container with label, min-width
│   ├── list.tsx              # List + List.Item — scrollable selection list
│   ├── text-input.tsx        # TextInput — single-line input
│   ├── badge.tsx             # Badge — count indicator
│   ├── button.tsx            # Button — action trigger
│   ├── divider.tsx           # Divider — horizontal separator
│   ├── scroll-area.tsx       # ScrollArea — scrollable container
│   ├── skeleton.tsx          # Skeleton — loading placeholder
│   ├── status-dot.tsx        # StatusDot — online/offline indicator
│   └── error-boundary.tsx    # ErrorBoundary — catches render errors
│
└── tests/                    # Test files
    ├── theme.test.tsx
    ├── auth.test.tsx
    ├── panel.test.tsx
    ├── list.test.tsx
    ├── text-input.test.tsx
    ├── badge.test.tsx
    ├── button.test.tsx
    ├── divider.test.tsx
    ├── scroll-area.test.tsx
    ├── skeleton.test.tsx
    ├── status-dot.test.tsx
    └── error-boundary.test.tsx
```

**Sub-module exports via package.json `exports` field:**
```json
{
  "exports": {
    ".": "./src/index.ts",
    "./theme": "./src/theme/index.ts",
    "./auth": "./src/auth/index.ts"
  }
}
```

In production builds, these point to compiled output in `dist/`.

---

## Chunk 1: Project Scaffolding & Theme System

### Task 1: Workspace and Package Setup

**Files:**
- Create: `pnpm-workspace.yaml`
- Create: `package.json` (root)
- Create: `packages/ui/package.json`
- Create: `packages/ui/tsconfig.json`
- Create: `packages/ui/vite.config.ts`
- Create: `packages/ui/vitest.config.ts`

- [ ] **Step 1: Create root workspace files**

```yaml
# pnpm-workspace.yaml
packages:
  - "packages/*"
  - "web"
```

```json
// package.json (root — private workspace root)
{
  "private": true,
  "packageManager": "pnpm@10.11.0",
  "scripts": {
    "build:ui": "pnpm --filter @workfort/ui build",
    "test:ui": "pnpm --filter @workfort/ui test",
    "dev:ui": "pnpm --filter @workfort/ui dev"
  }
}
```

- [ ] **Step 2: Create the @workfort/ui package**

```json
// packages/ui/package.json
{
  "name": "@workfort/ui",
  "version": "0.1.0",
  "type": "module",
  "license": "Apache-2.0",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js",
      "solid": "./src/index.ts"
    },
    "./theme": {
      "types": "./dist/theme/index.d.ts",
      "import": "./dist/theme/index.js",
      "solid": "./src/theme/index.ts"
    },
    "./auth": {
      "types": "./dist/auth/index.d.ts",
      "import": "./dist/auth/index.js",
      "solid": "./src/auth/index.ts"
    }
  },
  "files": ["dist", "src"],
  "scripts": {
    "dev": "vite build --watch",
    "build": "vite build && tsc --emitDeclarationOnly",
    "test": "vitest run",
    "test:watch": "vitest"
  },
  "peerDependencies": {
    "solid-js": "^1.9.0"
  },
  "devDependencies": {
    "@solidjs/testing-library": "^0.8.0",
    "@testing-library/jest-dom": "^6.0.0",
    "jsdom": "^26.0.0",
    "solid-js": "^1.9.0",
    "typescript": "^5.7.0",
    "vite": "^6.0.0",
    "vite-plugin-solid": "^2.11.0",
    "vitest": "^3.0.0"
  }
}
```

```json
// packages/ui/tsconfig.json
{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "preserve",
    "jsxImportSource": "solid-js",
    "strict": true,
    "declaration": true,
    "declarationDir": "dist",
    "emitDeclarationOnly": true,
    "outDir": "dist",
    "rootDir": "src",
    "skipLibCheck": true,
    "types": ["vitest/globals"]
  },
  "include": ["src/**/*.ts", "src/**/*.tsx"]
}
```

```ts
// packages/ui/vite.config.ts
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

export default defineConfig({
  plugins: [solid()],
  build: {
    lib: {
      entry: {
        index: "src/index.ts",
        "theme/index": "src/theme/index.ts",
        "auth/index": "src/auth/index.ts",
      },
      formats: ["es"],
    },
    rollupOptions: {
      external: ["solid-js", "solid-js/web", "solid-js/store"],
    },
  },
});
```

```ts
// packages/ui/vitest.config.ts
import { defineConfig } from "vitest/config";
import solid from "vite-plugin-solid";

export default defineConfig({
  plugins: [solid()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: [],
    deps: {
      optimizer: {
        web: { enabled: true },
      },
    },
  },
  resolve: {
    conditions: ["development", "browser"],
  },
});
```

- [ ] **Step 3: Install dependencies**

Run: `cd packages/ui && pnpm install`
Expected: `node_modules` created, `pnpm-lock.yaml` updated

- [ ] **Step 4: Create empty entry points so tests can run**

```ts
// packages/ui/src/index.ts
export {};
```

```ts
// packages/ui/src/theme/index.ts
export {};
```

```ts
// packages/ui/src/auth/index.ts
export {};
```

- [ ] **Step 5: Verify setup**

Run: `cd packages/ui && pnpm test`
Expected: "No test files found" (no tests yet — this is fine, confirms Vitest works)

- [ ] **Step 6: Commit**

```bash
git add pnpm-workspace.yaml package.json packages/ui/package.json packages/ui/tsconfig.json packages/ui/vite.config.ts packages/ui/vitest.config.ts packages/ui/src/index.ts packages/ui/src/theme/index.ts packages/ui/src/auth/index.ts pnpm-lock.yaml
git commit -m "feat(ui): scaffold @workfort/ui package with Vite + Vitest"
```

---

### Task 2: Theme Tokens and Provider

**Files:**
- Create: `packages/ui/src/theme/tokens.ts`
- Create: `packages/ui/src/theme/context.ts`
- Create: `packages/ui/src/theme/provider.tsx`
- Modify: `packages/ui/src/theme/index.ts`
- Create: `packages/ui/tests/theme.test.tsx`

- [ ] **Step 1: Write the failing test for theme system**

```tsx
// packages/ui/tests/theme.test.tsx
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { ThemeProvider, useTheme } from "../src/theme";

function ThemeConsumer() {
  const theme = useTheme();
  return (
    <div>
      <span data-testid="mode">{theme.mode()}</span>
      <button data-testid="toggle" onClick={theme.toggle}>
        toggle
      </button>
    </div>
  );
}

describe("ThemeProvider", () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove("light");
  });

  it("defaults to dark mode", () => {
    render(() => (
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    ));
    expect(screen.getByTestId("mode").textContent).toBe("dark");
  });

  it("toggles to light mode", async () => {
    render(() => (
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    ));
    screen.getByTestId("toggle").click();
    expect(screen.getByTestId("mode").textContent).toBe("light");
  });

  it("persists theme to localStorage", () => {
    render(() => (
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    ));
    screen.getByTestId("toggle").click();
    expect(localStorage.getItem("wf-theme")).toBe("light");
  });

  it("restores theme from localStorage", () => {
    localStorage.setItem("wf-theme", "light");
    render(() => (
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    ));
    expect(screen.getByTestId("mode").textContent).toBe("light");
  });

  it("adds light class to documentElement in light mode", () => {
    render(() => (
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    ));
    screen.getByTestId("toggle").click();
    expect(document.documentElement.classList.contains("light")).toBe(true);
  });

  it("removes light class when toggling back to dark", () => {
    render(() => (
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>
    ));
    screen.getByTestId("toggle").click();
    screen.getByTestId("toggle").click();
    expect(document.documentElement.classList.contains("light")).toBe(false);
  });
});

describe("useTheme", () => {
  it("throws when used outside ThemeProvider", () => {
    expect(() => render(() => <ThemeConsumer />)).toThrow(
      "useTheme must be used within a ThemeProvider"
    );
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd packages/ui && pnpm test`
Expected: FAIL — `ThemeProvider`, `useTheme` not exported

- [ ] **Step 3: Write the theme tokens**

```ts
// packages/ui/src/theme/tokens.ts

/** CSS custom property names for the WorkFort theme system. */
export const tokens = {
  // Colors
  bg: "--wf-bg",
  bgSecondary: "--wf-bg-secondary",
  text: "--wf-text",
  textSecondary: "--wf-text-secondary",
  textMuted: "--wf-text-muted",
  border: "--wf-border",
  accent: "--wf-accent",

  // Spacing
  spaceXs: "--wf-space-xs",
  spaceSm: "--wf-space-sm",
  spaceMd: "--wf-space-md",
  spaceLg: "--wf-space-lg",
  spaceXl: "--wf-space-xl",

  // Typography
  fontSans: "--wf-font-sans",
  fontMono: "--wf-font-mono",
  fontSizeXs: "--wf-font-size-xs",
  fontSizeSm: "--wf-font-size-sm",
  fontSizeBase: "--wf-font-size-base",
  fontSizeLg: "--wf-font-size-lg",

  // Radii
  radiusSm: "--wf-radius-sm",
  radiusMd: "--wf-radius-md",
  radiusLg: "--wf-radius-lg",
} as const;

export type ThemeToken = (typeof tokens)[keyof typeof tokens];
```

- [ ] **Step 4: Write the theme context**

```ts
// packages/ui/src/theme/context.ts
import { createContext, useContext, type Accessor } from "solid-js";

export type ThemeMode = "dark" | "light";

export interface ThemeContextValue {
  mode: Accessor<ThemeMode>;
  toggle: () => void;
}

export const ThemeContext = createContext<ThemeContextValue>();

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return ctx;
}
```

- [ ] **Step 5: Write the theme provider**

```tsx
// packages/ui/src/theme/provider.tsx
import { createSignal, createEffect, type ParentComponent } from "solid-js";
import { ThemeContext, type ThemeMode } from "./context";

const STORAGE_KEY = "wf-theme";

export const ThemeProvider: ParentComponent = (props) => {
  const stored = localStorage.getItem(STORAGE_KEY) as ThemeMode | null;
  const [mode, setMode] = createSignal<ThemeMode>(stored ?? "dark");

  const toggle = () => {
    setMode((prev) => (prev === "dark" ? "light" : "dark"));
  };

  createEffect(() => {
    const current = mode();
    localStorage.setItem(STORAGE_KEY, current);
    if (current === "light") {
      document.documentElement.classList.add("light");
    } else {
      document.documentElement.classList.remove("light");
    }
  });

  return (
    <ThemeContext.Provider value={{ mode, toggle }}>
      {props.children}
    </ThemeContext.Provider>
  );
};
```

- [ ] **Step 6: Wire up the theme index**

```ts
// packages/ui/src/theme/index.ts
export { ThemeProvider } from "./provider";
export { useTheme, type ThemeMode, type ThemeContextValue } from "./context";
export { tokens, type ThemeToken } from "./tokens";
```

- [ ] **Step 7: Run test to verify it passes**

Run: `cd packages/ui && pnpm test`
Expected: PASS (6 theme tests)

- [ ] **Step 8: Commit**

```bash
git add packages/ui/src/theme/ packages/ui/tests/theme.test.tsx
git commit -m "feat(ui): add theme system with tokens, provider, and useTheme"
```

---

## Chunk 2: Core Components (Part 1)

### Task 3: Panel Component

**Files:**
- Create: `packages/ui/src/panel.tsx`
- Create: `packages/ui/tests/panel.test.tsx`
- Modify: `packages/ui/src/index.ts`

- [ ] **Step 1: Write the failing test for Panel**

```tsx
// packages/ui/tests/panel.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { Panel } from "../src";

describe("Panel", () => {
  it("renders children", () => {
    render(() => <Panel>Hello</Panel>);
    expect(screen.getByText("Hello")).toBeDefined();
  });

  it("renders label when provided", () => {
    render(() => <Panel label="Channels">Content</Panel>);
    expect(screen.getByText("Channels")).toBeDefined();
  });

  it("omits label element when not provided", () => {
    const { container } = render(() => <Panel>Content</Panel>);
    expect(container.querySelector("[data-panel-label]")).toBeNull();
  });

  it("sets min-width style when minWidth is provided", () => {
    render(() => (
      <Panel minWidth={300} data-testid="panel">
        Content
      </Panel>
    ));
    const panel = screen.getByTestId("panel");
    expect(panel.style.minWidth).toBe("300px");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd packages/ui && pnpm test`
Expected: FAIL — `Panel` not exported

- [ ] **Step 3: Write Panel component**

```tsx
// packages/ui/src/panel.tsx
import type { JSX, ParentComponent } from "solid-js";
import { Show, splitProps } from "solid-js";

export interface PanelProps
  extends Omit<JSX.HTMLAttributes<HTMLDivElement>, "style"> {
  label?: string;
  minWidth?: number;
  style?: JSX.CSSProperties;
}

export const Panel: ParentComponent<PanelProps> = (props) => {
  const [local, rest] = splitProps(props, ["label", "minWidth", "children", "style"]);

  const style = (): JSX.CSSProperties => ({
    ...(local.style as JSX.CSSProperties | undefined),
    "min-width": local.minWidth ? `${local.minWidth}px` : undefined,
    "font-family": "var(--wf-font-sans)",
    color: "var(--wf-text)",
  });

  return (
    <div {...rest} style={style()} data-panel>
      <Show when={local.label}>
        <div
          data-panel-label
          style={{
            "font-size": "var(--wf-font-size-xs)",
            "text-transform": "uppercase",
            "letter-spacing": "0.05em",
            color: "var(--wf-text-muted)",
            padding: `var(--wf-space-sm) var(--wf-space-md)`,
            "font-weight": "600",
          }}
        >
          {local.label}
        </div>
      </Show>
      {local.children}
    </div>
  );
};
```

- [ ] **Step 4: Export from index**

```ts
// packages/ui/src/index.ts
export { Panel, type PanelProps } from "./panel";
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd packages/ui && pnpm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/ui/src/panel.tsx packages/ui/src/index.ts packages/ui/tests/panel.test.tsx
git commit -m "feat(ui): add Panel component"
```

---

### Task 4: List and List.Item Components

**Files:**
- Create: `packages/ui/src/list.tsx`
- Create: `packages/ui/tests/list.test.tsx`
- Modify: `packages/ui/src/index.ts`

- [ ] **Step 1: Write the failing test for List**

```tsx
// packages/ui/tests/list.test.tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { List } from "../src";

describe("List", () => {
  it("renders items via renderItem", () => {
    const items = ["alpha", "bravo", "charlie"];
    render(() => (
      <List
        items={items}
        renderItem={(item) => <List.Item>{item}</List.Item>}
      />
    ));
    expect(screen.getByText("alpha")).toBeDefined();
    expect(screen.getByText("bravo")).toBeDefined();
    expect(screen.getByText("charlie")).toBeDefined();
  });

  it("renders empty list without errors", () => {
    const { container } = render(() => (
      <List items={[]} renderItem={(item) => <List.Item>{item}</List.Item>} />
    ));
    expect(container.querySelector("[data-list]")).not.toBeNull();
  });
});

describe("List.Item", () => {
  it("renders children", () => {
    render(() => <List.Item>Item text</List.Item>);
    expect(screen.getByText("Item text")).toBeDefined();
  });

  it("applies active styling via data attribute", () => {
    render(() => (
      <List.Item active data-testid="item">
        Active item
      </List.Item>
    ));
    const item = screen.getByTestId("item");
    expect(item.dataset.active).toBe("");
  });

  it("calls onClick when clicked", () => {
    const onClick = vi.fn();
    render(() => (
      <List.Item onClick={onClick} data-testid="item">
        Clickable
      </List.Item>
    ));
    screen.getByTestId("item").click();
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("renders leading and trailing slots", () => {
    render(() => (
      <List.Item
        leading={<span data-testid="leading">L</span>}
        trailing={<span data-testid="trailing">T</span>}
      >
        Content
      </List.Item>
    ));
    expect(screen.getByTestId("leading")).toBeDefined();
    expect(screen.getByTestId("trailing")).toBeDefined();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd packages/ui && pnpm test`
Expected: FAIL — `List` not exported

- [ ] **Step 3: Write List component**

```tsx
// packages/ui/src/list.tsx
import type { JSX, ParentComponent } from "solid-js";
import { For, Show, splitProps } from "solid-js";

// --- List ---

export interface ListProps<T> {
  items: T[];
  renderItem: (item: T, index: number) => JSX.Element;
}

function ListRoot<T>(props: ListProps<T>) {
  return (
    <div data-list role="listbox" style={{ overflow: "auto" }}>
      <For each={props.items}>
        {(item, index) => props.renderItem(item, index())}
      </For>
    </div>
  );
}

// --- List.Item ---

export interface ListItemProps
  extends Omit<JSX.HTMLAttributes<HTMLDivElement>, "style"> {
  active?: boolean;
  leading?: JSX.Element;
  trailing?: JSX.Element;
  style?: JSX.CSSProperties;
}

const ListItem: ParentComponent<ListItemProps> = (props) => {
  const [local, rest] = splitProps(props, [
    "active",
    "leading",
    "trailing",
    "children",
    "onClick",
    "style",
  ]);

  return (
    <div
      {...rest}
      onClick={local.onClick}
      role="option"
      aria-selected={local.active ?? false}
      data-list-item
      {...(local.active ? { "data-active": "" } : {})}
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        display: "flex",
        "align-items": "center",
        gap: "var(--wf-space-sm)",
        padding: `var(--wf-space-xs) var(--wf-space-md)`,
        cursor: local.onClick ? "pointer" : "default",
        "font-size": "var(--wf-font-size-base)",
        "border-radius": "var(--wf-radius-sm)",
        color: "var(--wf-text)",
      }}
    >
      <Show when={local.leading}>{local.leading}</Show>
      <div style={{ flex: "1", "min-width": "0" }}>{local.children}</div>
      <Show when={local.trailing}>{local.trailing}</Show>
    </div>
  );
};

// --- Compound component ---

export const List = Object.assign(ListRoot, { Item: ListItem });
```

- [ ] **Step 4: Export from index**

```ts
// packages/ui/src/index.ts — add to existing exports
export { List, type ListProps, type ListItemProps } from "./list";
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd packages/ui && pnpm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/ui/src/list.tsx packages/ui/src/index.ts packages/ui/tests/list.test.tsx
git commit -m "feat(ui): add List and List.Item components"
```

---

### Task 5: TextInput, Button, and Badge

**Files:**
- Create: `packages/ui/src/text-input.tsx`
- Create: `packages/ui/src/button.tsx`
- Create: `packages/ui/src/badge.tsx`
- Create: `packages/ui/tests/text-input.test.tsx`
- Create: `packages/ui/tests/button.test.tsx`
- Create: `packages/ui/tests/badge.test.tsx`
- Modify: `packages/ui/src/index.ts`

- [ ] **Step 1: Write the failing tests**

```tsx
// packages/ui/tests/text-input.test.tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { TextInput } from "../src";

describe("TextInput", () => {
  it("renders an input element", () => {
    render(() => <TextInput data-testid="input" />);
    expect(screen.getByTestId("input").tagName).toBe("INPUT");
  });

  it("passes placeholder through", () => {
    render(() => <TextInput placeholder="Search..." data-testid="input" />);
    expect(screen.getByTestId("input").getAttribute("placeholder")).toBe(
      "Search..."
    );
  });

  it("calls onInput when typed into", async () => {
    const onInput = vi.fn();
    render(() => <TextInput onInput={onInput} data-testid="input" />);
    await fireEvent.input(screen.getByTestId("input"), {
      target: { value: "hello" },
    });
    expect(onInput).toHaveBeenCalled();
  });
});
```

```tsx
// packages/ui/tests/button.test.tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { Button } from "../src";

describe("Button", () => {
  it("renders children", () => {
    render(() => <Button>Click me</Button>);
    expect(screen.getByText("Click me")).toBeDefined();
  });

  it("renders a button element", () => {
    render(() => <Button data-testid="btn">OK</Button>);
    expect(screen.getByTestId("btn").tagName).toBe("BUTTON");
  });

  it("calls onClick when clicked", () => {
    const onClick = vi.fn();
    render(() => (
      <Button onClick={onClick} data-testid="btn">
        Go
      </Button>
    ));
    screen.getByTestId("btn").click();
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("supports disabled state", () => {
    render(() => (
      <Button disabled data-testid="btn">
        Nope
      </Button>
    ));
    expect(
      (screen.getByTestId("btn") as HTMLButtonElement).disabled
    ).toBe(true);
  });
});
```

```tsx
// packages/ui/tests/badge.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { Badge } from "../src";

describe("Badge", () => {
  it("renders the count", () => {
    render(() => <Badge count={5} />);
    expect(screen.getByText("5")).toBeDefined();
  });

  it("renders 99+ for large counts", () => {
    render(() => <Badge count={150} />);
    expect(screen.getByText("99+")).toBeDefined();
  });

  it("renders nothing when count is 0", () => {
    const { container } = render(() => <Badge count={0} />);
    expect(container.innerHTML).toBe("");
  });
});
```

- [ ] **Step 2: Run test to verify they fail**

Run: `cd packages/ui && pnpm test`
Expected: FAIL — `TextInput`, `Button`, `Badge` not exported

- [ ] **Step 3: Write the components**

```tsx
// packages/ui/src/text-input.tsx
import type { JSX } from "solid-js";
import { splitProps } from "solid-js";

export interface TextInputProps
  extends Omit<JSX.InputHTMLAttributes<HTMLInputElement>, "style"> {
  style?: JSX.CSSProperties;
}

export function TextInput(props: TextInputProps) {
  const [local, rest] = splitProps(props, ["style"]);

  return (
    <input
      type="text"
      {...rest}
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        width: "100%",
        padding: `var(--wf-space-xs) var(--wf-space-sm)`,
        "font-size": "var(--wf-font-size-base)",
        "font-family": "var(--wf-font-sans)",
        color: "var(--wf-text)",
        "background-color": "var(--wf-bg)",
        border: "1px solid var(--wf-border)",
        "border-radius": "var(--wf-radius-sm)",
        outline: "none",
      }}
    />
  );
}
```

```tsx
// packages/ui/src/button.tsx
import type { JSX, ParentComponent } from "solid-js";
import { splitProps } from "solid-js";

export interface ButtonProps
  extends Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "style"> {
  style?: JSX.CSSProperties;
}

export const Button: ParentComponent<ButtonProps> = (props) => {
  const [local, rest] = splitProps(props, ["children", "style"]);

  return (
    <button
      {...rest}
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        padding: `var(--wf-space-xs) var(--wf-space-md)`,
        "font-size": "var(--wf-font-size-base)",
        "font-family": "var(--wf-font-sans)",
        color: "var(--wf-text)",
        "background-color": "var(--wf-bg-secondary)",
        border: "1px solid var(--wf-border)",
        "border-radius": "var(--wf-radius-md)",
        cursor: rest.disabled ? "not-allowed" : "pointer",
        opacity: rest.disabled ? "0.5" : "1",
      }}
    >
      {local.children}
    </button>
  );
};
```

```tsx
// packages/ui/src/badge.tsx
import { Show } from "solid-js";

export interface BadgeProps {
  count: number;
  max?: number;
}

export function Badge(props: BadgeProps) {
  const max = () => props.max ?? 99;
  const label = () => (props.count > max() ? `${max()}+` : `${props.count}`);

  return (
    <Show when={props.count > 0}>
      <span
        data-badge
        style={{
          display: "inline-flex",
          "align-items": "center",
          "justify-content": "center",
          "min-width": "18px",
          height: "18px",
          padding: "0 5px",
          "font-size": "var(--wf-font-size-xs)",
          "font-family": "var(--wf-font-sans)",
          "font-weight": "600",
          "line-height": "1",
          color: "var(--wf-bg)",
          "background-color": "var(--wf-text-muted)",
          "border-radius": "9px",
        }}
      >
        {label()}
      </span>
    </Show>
  );
}
```

- [ ] **Step 4: Export from index**

```ts
// packages/ui/src/index.ts — add to existing exports
export { TextInput, type TextInputProps } from "./text-input";
export { Button, type ButtonProps } from "./button";
export { Badge, type BadgeProps } from "./badge";
```

- [ ] **Step 5: Run test to verify they pass**

Run: `cd packages/ui && pnpm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/ui/src/text-input.tsx packages/ui/src/button.tsx packages/ui/src/badge.tsx packages/ui/tests/text-input.test.tsx packages/ui/tests/button.test.tsx packages/ui/tests/badge.test.tsx packages/ui/src/index.ts
git commit -m "feat(ui): add TextInput, Button, and Badge components"
```

---

## Chunk 3: Core Components (Part 2)

### Task 6: Divider, StatusDot, and Skeleton

**Files:**
- Create: `packages/ui/src/divider.tsx`
- Create: `packages/ui/src/status-dot.tsx`
- Create: `packages/ui/src/skeleton.tsx`
- Create: `packages/ui/tests/divider.test.tsx`
- Create: `packages/ui/tests/status-dot.test.tsx`
- Create: `packages/ui/tests/skeleton.test.tsx`
- Modify: `packages/ui/src/index.ts`

- [ ] **Step 1: Write the failing tests**

```tsx
// packages/ui/tests/divider.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { Divider } from "../src";

describe("Divider", () => {
  it("renders an hr element", () => {
    render(() => <Divider data-testid="divider" />);
    expect(screen.getByTestId("divider").tagName).toBe("HR");
  });

  it("has separator role", () => {
    render(() => <Divider data-testid="divider" />);
    expect(screen.getByTestId("divider").getAttribute("role")).toBe(
      "separator"
    );
  });
});
```

```tsx
// packages/ui/tests/status-dot.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { StatusDot } from "../src";

describe("StatusDot", () => {
  it("renders online indicator", () => {
    render(() => <StatusDot online data-testid="dot" />);
    const dot = screen.getByTestId("dot");
    expect(dot.dataset.online).toBe("");
  });

  it("renders offline indicator", () => {
    render(() => <StatusDot online={false} data-testid="dot" />);
    const dot = screen.getByTestId("dot");
    expect(dot.dataset.online).toBeUndefined();
  });
});
```

```tsx
// packages/ui/tests/skeleton.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { Skeleton } from "../src";

describe("Skeleton", () => {
  it("renders with default height", () => {
    render(() => <Skeleton data-testid="skeleton" />);
    expect(screen.getByTestId("skeleton")).toBeDefined();
  });

  it("applies custom height", () => {
    render(() => <Skeleton height={40} data-testid="skeleton" />);
    expect(screen.getByTestId("skeleton").style.height).toBe("40px");
  });

  it("applies custom width", () => {
    render(() => <Skeleton width={200} data-testid="skeleton" />);
    expect(screen.getByTestId("skeleton").style.width).toBe("200px");
  });
});
```

- [ ] **Step 2: Run test to verify they fail**

Run: `cd packages/ui && pnpm test`
Expected: FAIL — components not exported

- [ ] **Step 3: Write the components**

```tsx
// packages/ui/src/divider.tsx
import type { JSX } from "solid-js";
import { splitProps } from "solid-js";

export interface DividerProps
  extends Omit<JSX.HTMLAttributes<HTMLHRElement>, "style"> {
  style?: JSX.CSSProperties;
}

export function Divider(props: DividerProps) {
  const [local, rest] = splitProps(props, ["style"]);

  return (
    <hr
      {...rest}
      role="separator"
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        border: "none",
        "border-top": "1px solid var(--wf-border)",
        margin: `var(--wf-space-sm) 0`,
      }}
    />
  );
}
```

```tsx
// packages/ui/src/status-dot.tsx
import type { JSX } from "solid-js";
import { splitProps } from "solid-js";

export interface StatusDotProps
  extends Omit<JSX.HTMLAttributes<HTMLSpanElement>, "style"> {
  online: boolean;
  style?: JSX.CSSProperties;
}

export function StatusDot(props: StatusDotProps) {
  const [local, rest] = splitProps(props, ["online", "style"]);

  return (
    <span
      {...rest}
      {...(local.online ? { "data-online": "" } : {})}
      aria-label={local.online ? "Online" : "Offline"}
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        display: "inline-block",
        width: "8px",
        height: "8px",
        "border-radius": "50%",
        "background-color": local.online
          ? "var(--wf-accent)"
          : "var(--wf-text-muted)",
        "flex-shrink": "0",
      }}
    />
  );
}
```

```tsx
// packages/ui/src/skeleton.tsx
import type { JSX } from "solid-js";
import { splitProps } from "solid-js";

export interface SkeletonProps
  extends Omit<JSX.HTMLAttributes<HTMLDivElement>, "style"> {
  width?: number;
  height?: number;
  style?: JSX.CSSProperties;
}

// Note: The `wf-pulse` keyframes animation must be defined by the shell's
// global CSS. The library only references the animation name.
export function Skeleton(props: SkeletonProps) {
  const [local, rest] = splitProps(props, ["width", "height", "style"]);

  return (
    <div
      {...rest}
      data-skeleton
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        width: local.width ? `${local.width}px` : "100%",
        height: local.height ? `${local.height}px` : "16px",
        "background-color": "var(--wf-bg-secondary)",
        "border-radius": "var(--wf-radius-sm)",
        animation: "wf-pulse 1.5s ease-in-out infinite",
      }}
    />
  );
}
```

- [ ] **Step 4: Export from index**

```ts
// packages/ui/src/index.ts — add to existing exports
export { Divider, type DividerProps } from "./divider";
export { StatusDot, type StatusDotProps } from "./status-dot";
export { Skeleton, type SkeletonProps } from "./skeleton";
```

- [ ] **Step 5: Run test to verify they pass**

Run: `cd packages/ui && pnpm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/ui/src/divider.tsx packages/ui/src/status-dot.tsx packages/ui/src/skeleton.tsx packages/ui/tests/divider.test.tsx packages/ui/tests/status-dot.test.tsx packages/ui/tests/skeleton.test.tsx packages/ui/src/index.ts
git commit -m "feat(ui): add Divider, StatusDot, and Skeleton components"
```

---

### Task 7: ScrollArea and ErrorBoundary

**Files:**
- Create: `packages/ui/src/scroll-area.tsx`
- Create: `packages/ui/src/error-boundary.tsx`
- Create: `packages/ui/tests/scroll-area.test.tsx`
- Create: `packages/ui/tests/error-boundary.test.tsx`
- Modify: `packages/ui/src/index.ts`

- [ ] **Step 1: Write the failing tests**

```tsx
// packages/ui/tests/scroll-area.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { ScrollArea } from "../src";

describe("ScrollArea", () => {
  it("renders children", () => {
    render(() => <ScrollArea>Scrollable content</ScrollArea>);
    expect(screen.getByText("Scrollable content")).toBeDefined();
  });

  it("applies max-height when provided", () => {
    render(() => (
      <ScrollArea maxHeight={300} data-testid="scroll">
        Content
      </ScrollArea>
    ));
    expect(screen.getByTestId("scroll").style.maxHeight).toBe("300px");
  });
});
```

```tsx
// packages/ui/tests/error-boundary.test.tsx
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { WfErrorBoundary } from "../src";

function ThrowingComponent(): never {
  throw new Error("Component exploded");
}

describe("WfErrorBoundary", () => {
  it("renders children when no error", () => {
    render(() => (
      <WfErrorBoundary>
        <span>Safe content</span>
      </WfErrorBoundary>
    ));
    expect(screen.getByText("Safe content")).toBeDefined();
  });

  it("catches errors and shows fallback", () => {
    // Suppress console.error for expected error
    const spy = vi.spyOn(console, "error").mockImplementation(() => {});
    render(() => (
      <WfErrorBoundary>
        <ThrowingComponent />
      </WfErrorBoundary>
    ));
    expect(screen.getByText(/something went wrong/i)).toBeDefined();
    spy.mockRestore();
  });
});
```

- [ ] **Step 2: Run test to verify they fail**

Run: `cd packages/ui && pnpm test`
Expected: FAIL

- [ ] **Step 3: Write the components**

```tsx
// packages/ui/src/scroll-area.tsx
import type { JSX, ParentComponent } from "solid-js";
import { splitProps } from "solid-js";

export interface ScrollAreaProps
  extends Omit<JSX.HTMLAttributes<HTMLDivElement>, "style"> {
  maxHeight?: number;
  style?: JSX.CSSProperties;
}

export const ScrollArea: ParentComponent<ScrollAreaProps> = (props) => {
  const [local, rest] = splitProps(props, ["maxHeight", "children", "style"]);

  return (
    <div
      {...rest}
      data-scroll-area
      style={{
        ...(local.style as JSX.CSSProperties | undefined),
        overflow: "auto",
        "max-height": local.maxHeight ? `${local.maxHeight}px` : undefined,
      }}
    >
      {local.children}
    </div>
  );
};
```

```tsx
// packages/ui/src/error-boundary.tsx
import { ErrorBoundary, type ParentComponent } from "solid-js";

export interface WfErrorBoundaryProps {
  fallback?: (err: Error) => import("solid-js").JSX.Element;
}

export const WfErrorBoundary: ParentComponent<WfErrorBoundaryProps> = (
  props
) => {
  const fallback = (err: Error) => {
    if (props.fallback) return props.fallback(err);
    return (
      <div
        data-error-boundary
        style={{
          padding: "var(--wf-space-md)",
          color: "var(--wf-text-muted)",
          "font-size": "var(--wf-font-size-sm)",
          "font-family": "var(--wf-font-sans)",
        }}
      >
        Something went wrong: {err.message}
      </div>
    );
  };

  return (
    <ErrorBoundary fallback={(err) => fallback(err as Error)}>
      {props.children}
    </ErrorBoundary>
  );
};
```

- [ ] **Step 4: Export from index**

```ts
// packages/ui/src/index.ts — add to existing exports
export { ScrollArea, type ScrollAreaProps } from "./scroll-area";
export { WfErrorBoundary, type WfErrorBoundaryProps } from "./error-boundary";
```

- [ ] **Step 5: Run test to verify they pass**

Run: `cd packages/ui && pnpm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add packages/ui/src/scroll-area.tsx packages/ui/src/error-boundary.tsx packages/ui/tests/scroll-area.test.tsx packages/ui/tests/error-boundary.test.tsx packages/ui/src/index.ts
git commit -m "feat(ui): add ScrollArea and ErrorBoundary components"
```

---

## Chunk 4: Auth Module & Build Verification

### Task 8: Auth Context and Provider

**Files:**
- Create: `packages/ui/src/auth/context.ts`
- Create: `packages/ui/src/auth/provider.tsx`
- Create: `packages/ui/src/auth/require-auth.tsx`
- Modify: `packages/ui/src/auth/index.ts`
- Create: `packages/ui/tests/auth.test.tsx`

The auth module wraps better-auth's client. Services use `useAuth()` for the authenticated user. The actual better-auth client is injected by the shell via the provider — `@workfort/ui` doesn't depend on better-auth directly.

- [ ] **Step 1: Write the failing test**

```tsx
// packages/ui/tests/auth.test.tsx
import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { AuthProvider, useAuth, RequireAuth } from "../src/auth";

function AuthConsumer() {
  const auth = useAuth();
  return (
    <div>
      <span data-testid="loading">{String(auth.loading())}</span>
      <span data-testid="user">{auth.user()?.username ?? "none"}</span>
    </div>
  );
}

describe("useAuth", () => {
  it("throws when used outside AuthProvider", () => {
    expect(() => render(() => <AuthConsumer />)).toThrow(
      "useAuth must be used within an AuthProvider"
    );
  });
});

describe("AuthProvider", () => {
  it("provides loading state", () => {
    render(() => (
      <AuthProvider session={null} loading={true}>
        <AuthConsumer />
      </AuthProvider>
    ));
    expect(screen.getByTestId("loading").textContent).toBe("true");
    expect(screen.getByTestId("user").textContent).toBe("none");
  });

  it("provides user when session exists", () => {
    const user = {
      id: "user-1",
      username: "kazw",
      name: "Kaz Walker",
      displayName: "Kaz",
      type: "user" as const,
    };
    render(() => (
      <AuthProvider session={{ user }} loading={false}>
        <AuthConsumer />
      </AuthProvider>
    ));
    expect(screen.getByTestId("user").textContent).toBe("kazw");
    expect(screen.getByTestId("loading").textContent).toBe("false");
  });
});

describe("RequireAuth", () => {
  it("renders children when authenticated", () => {
    const user = {
      id: "user-1",
      username: "kazw",
      name: "Kaz Walker",
      displayName: "Kaz",
      type: "user" as const,
    };
    render(() => (
      <AuthProvider session={{ user }} loading={false}>
        <RequireAuth fallback={<span>Login</span>}>
          <span>Protected</span>
        </RequireAuth>
      </AuthProvider>
    ));
    expect(screen.getByText("Protected")).toBeDefined();
  });

  it("renders fallback when not authenticated", () => {
    render(() => (
      <AuthProvider session={null} loading={false}>
        <RequireAuth fallback={<span>Login</span>}>
          <span>Protected</span>
        </RequireAuth>
      </AuthProvider>
    ));
    expect(screen.getByText("Login")).toBeDefined();
  });

  it("renders neither children nor fallback while loading", () => {
    render(() => (
      <AuthProvider session={null} loading={true}>
        <RequireAuth fallback={<span>Login</span>}>
          <span>Protected</span>
        </RequireAuth>
      </AuthProvider>
    ));
    expect(screen.queryByText("Protected")).toBeNull();
    expect(screen.queryByText("Login")).toBeNull();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd packages/ui && pnpm test`
Expected: FAIL — `AuthProvider`, `useAuth`, `RequireAuth` not exported

- [ ] **Step 3: Write the auth context**

```ts
// packages/ui/src/auth/context.ts
import { createContext, useContext, type Accessor } from "solid-js";

export interface AuthUser {
  id: string;
  username: string;
  name: string;
  displayName: string;
  type: "user" | "agent" | "service";
}

export interface AuthSession {
  user: AuthUser;
}

export interface AuthContextValue {
  user: Accessor<AuthUser | null>;
  session: Accessor<AuthSession | null>;
  loading: Accessor<boolean>;
}

export const AuthContext = createContext<AuthContextValue>();

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
```

- [ ] **Step 4: Write the auth provider**

```tsx
// packages/ui/src/auth/provider.tsx
import { type ParentComponent } from "solid-js";
import { AuthContext, type AuthSession } from "./context";

export interface AuthProviderProps {
  session: AuthSession | null;
  loading: boolean;
}

export const AuthProvider: ParentComponent<AuthProviderProps> = (props) => {
  return (
    <AuthContext.Provider
      value={{
        user: () => props.session?.user ?? null,
        session: () => props.session,
        loading: () => props.loading,
      }}
    >
      {props.children}
    </AuthContext.Provider>
  );
};
```

- [ ] **Step 5: Write the RequireAuth guard**

```tsx
// packages/ui/src/auth/require-auth.tsx
import { Show, type JSX, type ParentComponent } from "solid-js";
import { useAuth } from "./context";

export interface RequireAuthProps {
  fallback: JSX.Element;
}

export const RequireAuth: ParentComponent<RequireAuthProps> = (props) => {
  const auth = useAuth();

  return (
    <Show
      when={!auth.loading()}
      fallback={null}
    >
      <Show when={auth.user()} fallback={props.fallback}>
        {props.children}
      </Show>
    </Show>
  );
};
```

- [ ] **Step 6: Wire up the auth index**

```ts
// packages/ui/src/auth/index.ts
export { AuthProvider, type AuthProviderProps } from "./provider";
export {
  useAuth,
  type AuthUser,
  type AuthSession,
  type AuthContextValue,
} from "./context";
export { RequireAuth, type RequireAuthProps } from "./require-auth";
```

- [ ] **Step 7: Run test to verify it passes**

Run: `cd packages/ui && pnpm test`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add packages/ui/src/auth/ packages/ui/tests/auth.test.tsx
git commit -m "feat(ui): add auth module with AuthProvider, useAuth, RequireAuth"
```

---

### Task 9: Build Verification and Package Exports

**Files:**
- No new files — verify the complete package builds and exports correctly.

- [ ] **Step 1: Run the full test suite**

Run: `cd packages/ui && pnpm test`
Expected: PASS — all tests across all components, theme, and auth

- [ ] **Step 2: Build the library**

Run: `cd packages/ui && pnpm build`
Expected: `dist/` directory created with:
- `dist/index.js` — main exports
- `dist/theme/index.js` — theme sub-module
- `dist/auth/index.js` — auth sub-module
- `dist/index.d.ts` — type declarations
- `dist/theme/index.d.ts`
- `dist/auth/index.d.ts`

- [ ] **Step 3: Verify exports resolve**

Run: `node -e "import('@workfort/ui').then(m => console.log(Object.keys(m)))"`
Expected: Lists all exported component names (Panel, List, TextInput, etc.)

If this fails due to workspace resolution, verify with:
`ls packages/ui/dist/`

- [ ] **Step 4: Update .gitignore for node_modules and dist**

Add to `.gitignore`:
```
node_modules/
packages/ui/dist/
```

- [ ] **Step 5: Commit**

```bash
git add .gitignore
git commit -m "chore: add node_modules and dist to gitignore"
```
