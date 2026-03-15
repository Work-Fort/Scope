# Shell Boot Provider — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the shell's boot sequence configurable per environment via a Vite virtual module that resolves the right boot provider at build time.

**Architecture:** A Vite plugin generates `virtual:boot-provider` based on environment detection. Each environment (BFF, Tauri) provides a `BootProvider` implementation with `configure()`, `authenticate()`, and optional `beforeMount()` stages. The shell runs the pipeline before rendering.

**Tech Stack:** Vite (virtual module plugin), SolidJS, TypeScript, @workfort/auth

**Spec:** `docs/shell-boot-provider-design.md`

---

## Chunk 1: Types and Vite Plugin

### Task 1: Create boot provider type definitions

**Why:** All boot providers implement this interface. Types go in first so both `bff.ts` and `tauri.ts` can import them.

**Files:**
- Create: `web/shell/src/boot/types.ts`

- [ ] **Step 1: Create boot directory**

```bash
mkdir -p web/shell/src/boot
```

- [ ] **Step 2: Create types.ts**

```typescript
// web/shell/src/boot/types.ts

export interface BootProvider {
  configure(): Promise<BootConfig>;
  authenticate(): Promise<AuthState>;
  beforeMount?(): Promise<void>;
}

export interface BootConfig {
  apiBase: string;
  environment: 'bff' | 'tauri' | string;
  needsSetup?: boolean;
}

export interface AuthState {
  authenticated: boolean;
  user?: { id: string; email: string; name: string };
}
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd web/shell && npx tsc --noEmit src/boot/types.ts
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add web/shell/src/boot/types.ts
git commit -m "feat(shell): add BootProvider, BootConfig, AuthState type definitions"
```

### Task 2: Create Vite boot-provider plugin

**Why:** The virtual module `virtual:boot-provider` resolves to the correct boot implementation at build time. Environment detection uses `process.env.TAURI_ENV` which Tauri sets during its build process.

**Files:**
- Create: `web/shell/plugins/boot-provider.ts`

- [ ] **Step 1: Create plugins directory**

```bash
mkdir -p web/shell/plugins
```

- [ ] **Step 2: Create the plugin**

```typescript
// web/shell/plugins/boot-provider.ts

import type { Plugin } from 'vite';

export function bootProviderPlugin(): Plugin {
  return {
    name: 'workfort-boot-provider',
    resolveId(id) {
      if (id === 'virtual:boot-provider') return '\0boot-provider';
    },
    load(id) {
      if (id !== '\0boot-provider') return;
      if (process.env.TAURI_ENV) {
        return `export { default } from '../src/boot/tauri';`;
      }
      return `export { default } from '../src/boot/bff';`;
    },
  };
}
```

- [ ] **Step 3: Create TypeScript type declaration for the virtual module**

The `tsconfig.json` only includes `src/`, so we need a `.d.ts` inside `src/` to declare the virtual module type. This lets TypeScript understand `import bootProvider from 'virtual:boot-provider'`.

```typescript
// web/shell/src/boot/virtual.d.ts

declare module 'virtual:boot-provider' {
  import type { BootProvider } from './types';
  const provider: BootProvider;
  export default provider;
}
```

- [ ] **Step 4: Commit**

```bash
git add web/shell/plugins/boot-provider.ts web/shell/src/boot/virtual.d.ts
git commit -m "feat(shell): add Vite boot-provider virtual module plugin"
```

### Task 3: Register plugin in vite.config.ts

**Why:** The plugin must be registered for Vite to resolve `virtual:boot-provider` imports.

**Files:**
- Modify: `web/shell/vite.config.ts`

- [ ] **Step 1: Add import and register plugin**

Add the import at the top of `vite.config.ts`:

```typescript
import { bootProviderPlugin } from './plugins/boot-provider';
```

Add `bootProviderPlugin()` to the `plugins` array, before the other plugins:

```typescript
plugins: [
    bootProviderPlugin(),
    UnoCSS(),
    solid(),
    federation({
```

The boot-provider plugin should come first so the virtual module is resolved before other plugins process the import graph.

- [ ] **Step 2: Verify Vite config loads without errors**

```bash
cd web/shell && npx vite build --mode development 2>&1 | head -20
```

Expected: Build starts (may fail later because `bff.ts` doesn't exist yet — that's fine). The important thing is no "plugin not found" or config parse errors. Look for `vite` output, not an immediate crash.

- [ ] **Step 3: Commit**

```bash
git add web/shell/vite.config.ts
git commit -m "feat(shell): register boot-provider plugin in Vite config"
```

---

## Chunk 2: BFF Boot Provider

### Task 4: Create BFF boot provider

**Why:** The BFF (browser-for-frontend) provider is the default path for web builds. It returns an empty `apiBase` (relative URLs work via the Go BFF proxy) and uses `getAuthClient()` from `@workfort/auth` which auto-detects `WebAuthClient`.

**Files:**
- Create: `web/shell/src/boot/bff.ts`

- [ ] **Step 1: Create the provider**

```typescript
// web/shell/src/boot/bff.ts

import { getAuthClient } from '@workfort/auth';
import type { BootProvider } from './types';

const provider: BootProvider = {
  async configure() {
    return { apiBase: '', environment: 'bff' as const };
  },
  async authenticate() {
    const client = await getAuthClient();
    const user = await client.getUser();
    return { authenticated: !!user, user: user ?? undefined };
  },
};

export default provider;
```

Note: `getAuthClient()` returns `Promise<AuthClient>` so we must `await` it. The design spec showed it without `await` — this matches the actual `@workfort/auth` API from `web/packages/auth/src/index.ts`.

Also note `user` from `client.getUser()` is `UserInfo | null`, but `AuthState.user` expects `UserInfo | undefined`, so we convert with `?? undefined`.

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd web/shell && npx tsc --noEmit
```

Expected: May have errors in `index.tsx` (it doesn't use boot provider yet) but `bff.ts` itself should have no type errors. Check that no errors reference `boot/bff.ts`.

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/boot/bff.ts
git commit -m "feat(shell): add BFF boot provider implementation"
```

### Task 5: Modify index.tsx to use boot pipeline

**Why:** Replace the direct render with the boot pipeline: configure -> authenticate -> optional beforeMount -> render. The boot provider is imported from the virtual module.

**Files:**
- Modify: `web/shell/src/index.tsx`

- [ ] **Step 1: Replace index.tsx contents**

Replace the entire file:

```typescript
// web/shell/src/index.tsx

import 'virtual:uno.css';
import '@workfort/ui';
import '@workfort/ui/style.css';
import './global.css';
import { render } from 'solid-js/web';
import bootProvider from 'virtual:boot-provider';
import App from './app';

async function boot() {
  const config = await bootProvider.configure();
  const auth = await bootProvider.authenticate();
  await bootProvider.beforeMount?.();

  render(() => <App config={config} auth={auth} />, document.getElementById('app')!);
}

boot();
```

- [ ] **Step 2: Commit**

```bash
git add web/shell/src/index.tsx
git commit -m "feat(shell): use boot pipeline in index.tsx instead of direct render"
```

### Task 6: Modify app.tsx to accept config and auth props

**Why:** The `App` component needs to receive `config` and `auth` from the boot pipeline. When `config.needsSetup` is true (Tauri with no server URL configured), the app should render a setup screen instead of the normal router.

**Files:**
- Modify: `web/shell/src/app.tsx`

- [ ] **Step 1: Add props interface and update App component**

Add the import for types and update the `App` component signature to accept props:

```typescript
import type { BootConfig, AuthState } from './boot/types';
```

Change the `App` component to accept and use the props:

```typescript
interface AppProps {
  config: BootConfig;
  auth: AuthState;
}

const App: Component<AppProps> = (props) => {
  return (
    <Show when={!props.config.needsSetup} fallback={<SetupScreen />}>
      <Router>
        <Route path="/" component={FortPicker} />
        <Route path="/forts/:fort" component={FortShell}>
          <Route path="/:service/*rest" component={ServicePage} />
          <Route path="/" component={FortIndex} />
        </Route>
      </Router>
    </Show>
  );
};
```

Add a minimal `SetupScreen` placeholder component in the same file (before `App`):

```typescript
const SetupScreen: Component = () => {
  return (
    <div class="shell-unavailable">
      <div style={{ "text-align": "center" }}>
        <h2>WorkFort Setup</h2>
        <p style={{ color: "var(--wf-color-text-muted)", "margin-top": "var(--wf-space-sm)" }}>
          Configure a server URL to get started.
        </p>
      </div>
    </div>
  );
};
```

The full modified `app.tsx` should look like:

```typescript
import {
  type Component,
  Show,
  createContext,
  createEffect,
  createMemo,
  createSignal,
  onCleanup,
  useContext,
} from 'solid-js';
import { Navigate, Route, Router, useParams } from '@solidjs/router';
import { services, startPolling, stopPolling } from './stores/services';
import './stores/theme';
import ShellLayout from './components/shell-layout';
import ServiceMount from './components/service-mount';
import Unavailable from './components/unavailable';
import FortPicker from './components/fort-picker';
import type { ServiceModule } from './lib/remotes';
import type { BootConfig, AuthState } from './boot/types';

// Context to pass sidebar setter from FortShell to ServicePage.
const FortShellContext = createContext<{
  setSidebarComponent: (v: (() => any) | undefined) => void;
}>({ setSidebarComponent: () => {} });

interface AppProps {
  config: BootConfig;
  auth: AuthState;
}

const SetupScreen: Component = () => {
  return (
    <div class="shell-unavailable">
      <div style={{ "text-align": "center" }}>
        <h2>WorkFort Setup</h2>
        <p style={{ color: "var(--wf-color-text-muted)", "margin-top": "var(--wf-space-sm)" }}>
          Configure a server URL to get started.
        </p>
      </div>
    </div>
  );
};

const App: Component<AppProps> = (props) => {
  return (
    <Show when={!props.config.needsSetup} fallback={<SetupScreen />}>
      <Router>
        <Route path="/" component={FortPicker} />
        <Route path="/forts/:fort" component={FortShell}>
          <Route path="/:service/*rest" component={ServicePage} />
          <Route path="/" component={FortIndex} />
        </Route>
      </Router>
    </Show>
  );
};

const FortShell: Component = (props: { children?: any }) => {
  const params = useParams<{ fort: string }>();
  const [sidebarComponent, setSidebarComponent] = createSignal<(() => any) | undefined>();

  createEffect(() => {
    const fort = params.fort;
    startPolling(fort);
  });
  onCleanup(() => stopPolling());

  return (
    <FortShellContext.Provider value={{ setSidebarComponent }}>
      <ShellLayout sidebar={sidebarComponent()}>{props.children}</ShellLayout>
    </FortShellContext.Provider>
  );
};

const ServicePage: Component = () => {
  const params = useParams<{ fort: string; service: string }>();
  const ctx = useContext(FortShellContext);

  const handleModule = (mod: ServiceModule | null) => {
    ctx.setSidebarComponent(
      mod?.SidebarContent ? () => mod.SidebarContent! : undefined,
    );
  };

  const svc = createMemo(() =>
    services().find((s) => s.enabled && s.route === `/${params.service}`),
  );

  return (
    <>
      {svc() ? (
        svc()!.ui ? (
          <ServiceMount name={svc()!.name} label={svc()!.label} connected={svc()!.connected} onModule={handleModule} />
        ) : (
          <Unavailable label={svc()!.label} />
        )
      ) : (
        <Navigate href={`/forts/${params.fort}`} />
      )}
    </>
  );
};

const FortIndex: Component = () => {
  const params = useParams<{ fort: string }>();
  const firstRoute = createMemo(() => {
    const enabled = services().find((s) => s.enabled);
    return enabled ? `/forts/${params.fort}${enabled.route}` : null;
  });

  return (
    <Show when={firstRoute()} fallback={<div class="shell-unavailable">No services available.</div>}>
      <Navigate href={firstRoute()!} />
    </Show>
  );
};

export default App;
```

- [ ] **Step 2: Verify BFF build succeeds**

```bash
cd web/shell && pnpm build
```

Expected: Build succeeds. The virtual module resolves to `bff.ts`, the `App` component accepts `config` and `auth` props, and tree-shaking eliminates the Tauri code path.

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/app.tsx
git commit -m "feat(shell): App component accepts config/auth props, shows SetupScreen when needsSetup"
```

---

## Chunk 3: Tauri Boot Provider

### Task 7: Create Tauri boot provider

**Why:** The Tauri provider calls `invoke('get_server_url')` to check for a stored server URL. If none exists, it returns `needsSetup: true` so the shell renders the setup screen. Authentication uses the same `getAuthClient()` factory which auto-detects `TauriAuthClient` via `window.__TAURI_INTERNALS__`.

**Files:**
- Create: `web/shell/src/boot/tauri.ts`

- [ ] **Step 1: Create the provider**

```typescript
// web/shell/src/boot/tauri.ts

import { invoke } from '@tauri-apps/api/core';
import { getAuthClient } from '@workfort/auth';
import type { BootProvider } from './types';

const provider: BootProvider = {
  async configure() {
    const url = await invoke<string | null>('get_server_url');
    if (!url) {
      return { apiBase: '', environment: 'tauri' as const, needsSetup: true };
    }
    return { apiBase: url, environment: 'tauri' as const };
  },
  async authenticate() {
    const client = await getAuthClient();
    const user = await client.getUser();
    return { authenticated: !!user, user: user ?? undefined };
  },
};

export default provider;
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd web/shell && npx tsc --noEmit
```

Expected: No errors. `@tauri-apps/api` is in `devDependencies` so types are available.

- [ ] **Step 3: Commit**

```bash
git add web/shell/src/boot/tauri.ts
git commit -m "feat(shell): add Tauri boot provider implementation"
```

### Task 8: Add Tauri commands for server URL storage

**Why:** The Tauri boot provider calls `invoke('get_server_url')` and the setup screen will call `invoke('set_server_url')`. These commands need Rust implementations. The URL is stored in `AppState` using an `Option<String>` wrapped in a `Mutex`, alongside the existing token store.

**Files:**
- Modify: `src-tauri/src/proxy.rs` — add `server_url` field to `AppState`
- Modify: `src-tauri/src/lib.rs` — add `get_server_url` and `set_server_url` commands, register them

- [ ] **Step 1: Add server_url field to AppState in proxy.rs**

Add a new field to the `AppState` struct:

```rust
pub server_url: Arc<Mutex<Option<String>>>,
```

Update `AppState::new()` to initialize it from the environment variable (if set):

```rust
impl AppState {
    pub fn new(api_base_url: &str) -> Self {
        let client = Client::builder()
            .timeout(std::time::Duration::from_secs(10))
            .build()
            .expect("Failed to build HTTP client");

        Self {
            client,
            tokens: TokenStore::new(),
            api_base: Url::parse(api_base_url).expect("Invalid API base URL"),
            server_url: Arc::new(Mutex::new(Some(api_base_url.to_string()))),
        }
    }
}
```

- [ ] **Step 2: Add get_server_url and set_server_url commands in lib.rs**

Add the command functions before the `run()` function:

```rust
/// Tauri command: get the stored server URL.
/// Returns None if no URL has been configured.
#[tauri::command]
fn get_server_url(state: tauri::State<'_, AppState>) -> Option<String> {
    state.server_url.lock().unwrap().clone()
}

/// Tauri command: set the server URL.
/// Called from the setup screen when the user configures a server.
#[tauri::command]
fn set_server_url(state: tauri::State<'_, AppState>, url: String) {
    *state.server_url.lock().unwrap() = Some(url);
}
```

Register the new commands in `tauri::generate_handler!`:

```rust
.invoke_handler(tauri::generate_handler![
    auth::login,
    auth::logout,
    auth::get_user,
    get_server_url,
    set_server_url,
])
```

- [ ] **Step 3: Verify Rust compiles**

```bash
cd src-tauri && cargo check
```

Expected: Compiles without errors.

- [ ] **Step 4: Commit**

```bash
git add src-tauri/src/proxy.rs src-tauri/src/lib.rs
git commit -m "feat(tauri): add get_server_url and set_server_url IPC commands"
```

---

## Chunk 4: Verification

### Task 9: Verify BFF build works

**Why:** Confirm the full BFF build pipeline works end-to-end with the boot provider.

- [ ] **Step 1: Run BFF build**

```bash
cd web/shell && pnpm build
```

Expected: Build succeeds with no errors. The virtual module resolves to `bff.ts`.

- [ ] **Step 2: Verify the built output includes boot code**

```bash
grep -l 'environment' web/shell/dist/assets/*.js | head -1
```

Expected: At least one JS file contains `environment` from the boot config.

```bash
grep -c 'tauri' web/shell/dist/assets/*.js
```

Expected: All counts are `0` — Tauri code is tree-shaken out in BFF builds.

### Task 10: Verify Tauri build works

**Why:** Confirm the Tauri build resolves the virtual module to `tauri.ts` and compiles both the Rust and frontend sides.

- [ ] **Step 1: Run Rust check**

```bash
cd src-tauri && cargo check
```

Expected: Compiles without errors.

- [ ] **Step 2: Run Tauri build (debug mode)**

```bash
cd web/shell && npx @tauri-apps/cli build --debug
```

Expected: Both frontend and Rust compile. A debug binary is produced. The `TAURI_ENV` variable is set by the Tauri CLI during builds, so the virtual module resolves to `tauri.ts`.

Note: This step requires Tauri build toolchain to be installed. If not available, verify the Rust side compiles (`cargo check`) and the frontend compiles with `TAURI_ENV=1 pnpm build` to simulate the environment detection:

```bash
cd web/shell && TAURI_ENV=1 pnpm build
```

Expected: Build succeeds. Verify Tauri code is included:

```bash
grep -c 'get_server_url' web/shell/dist/assets/*.js
```

Expected: At least one file contains `get_server_url` (the Tauri invoke call).

- [ ] **Step 3: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "fix(shell): boot provider build fixes"
```

Only create this commit if changes were needed. Skip if everything passed cleanly.
