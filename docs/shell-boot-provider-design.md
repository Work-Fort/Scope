# Shell Boot Provider — Design Spec

**Goal:** Make the shell's boot sequence configurable per environment (BFF, Tauri desktop, Tauri mobile, future) via a Vite virtual module that resolves the right boot provider at build time.

**Key Principle:** Each environment provides a `BootProvider` implementation. The shell imports `virtual:boot-provider` and runs a standard pipeline. Dead code for other environments is tree-shaken out. Adding a new environment is one file + one condition.

---

## Boot Provider Interface

```ts
interface BootProvider {
  configure(): Promise<BootConfig>;
  authenticate(): Promise<AuthState>;
  beforeMount?(): Promise<void>;
}

interface BootConfig {
  apiBase: string;
  environment: 'bff' | 'tauri' | string;
  needsSetup?: boolean;
}

interface AuthState {
  authenticated: boolean;
  user?: { id: string; email: string; name: string };
}
```

### Stage responsibilities

| Stage | BFF | Tauri |
|-------|-----|-------|
| `configure()` | Returns `{ apiBase: '', environment: 'bff' }` immediately | Checks stored URL via `invoke('get_server_url')`. If none, returns `{ needsSetup: true }` |
| `authenticate()` | Uses `WebAuthClient` from `@workfort/auth` | Uses `TauriAuthClient` from `@workfort/auth` |
| `beforeMount()` | Not used | Optional: register config service as MF remote, inject platform capabilities |

---

## Boot Pipeline

`web/shell/src/index.tsx` runs the pipeline before rendering:

```ts
import bootProvider from 'virtual:boot-provider';

const config = await bootProvider.configure();
const auth = await bootProvider.authenticate();
await bootProvider.beforeMount?.();

render(() => <App config={config} auth={auth} />, document.getElementById('app')!);
```

If `config.needsSetup` is true, the `App` component renders the setup/config service instead of the fort picker. Once the user configures a server URL, the app re-boots with the new config.

---

## Vite Plugin

`web/shell/plugins/boot-provider.ts`:

```ts
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
    }
  };
}
```

Registered in `web/shell/vite.config.ts` alongside existing plugins.

---

## Boot Implementations

### BFF (`web/shell/src/boot/bff.ts`)

```ts
import { getAuthClient } from '@workfort/auth';
import type { BootProvider } from './types';

const provider: BootProvider = {
  async configure() {
    return { apiBase: '', environment: 'bff' };
  },
  async authenticate() {
    const client = getAuthClient();
    const user = await client.getUser();
    return { authenticated: !!user, user };
  },
};

export default provider;
```

### Tauri (`web/shell/src/boot/tauri.ts`)

```ts
import { invoke } from '@tauri-apps/api/core';
import { getAuthClient } from '@workfort/auth';
import type { BootProvider } from './types';

const provider: BootProvider = {
  async configure() {
    const url = await invoke<string | null>('get_server_url');
    if (!url) {
      return { apiBase: '', environment: 'tauri', needsSetup: true };
    }
    return { apiBase: url, environment: 'tauri' };
  },
  async authenticate() {
    const client = getAuthClient();
    const user = await client.getUser();
    return { authenticated: !!user, user };
  },
};

export default provider;
```

---

## File Structure

```
web/shell/
├── plugins/
│   └── boot-provider.ts      # Vite plugin
├── src/
│   ├── boot/
│   │   ├── types.ts           # BootProvider, BootConfig, AuthState interfaces
│   │   ├── bff.ts             # BFF boot provider
│   │   └── tauri.ts           # Tauri boot provider
│   ├── index.tsx              # Modified to use boot pipeline
│   └── app.tsx                # Modified to handle needsSetup
```

---

## Adding a New Environment

1. Create `web/shell/src/boot/newenv.ts` implementing `BootProvider`
2. Add a condition in `plugins/boot-provider.ts`:
   ```ts
   if (process.env.NEW_ENV) {
     return `export { default } from '../src/boot/newenv';`;
   }
   ```
3. No changes to the shell, app, or any other file
