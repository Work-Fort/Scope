# React Getting Started

Build a React service frontend for WorkFort.

## 1. Scaffold

Create a new React + TypeScript Vite project:

```bash
pnpm create vite my-service --template react-ts
cd my-service
pnpm install
```

## 2. Install Dependencies

```bash
pnpm add @workfort/ui @workfort/ui-react @workfort/auth @module-federation/vite
```

- `@workfort/ui` — Web Components and shared design tokens
- `@workfort/ui-react` — React wrappers for Web Components (required for React 18 event handling)
- `@workfort/auth` — Authentication client (shared singleton)
- `@module-federation/vite` — Module Federation for remote loading

## 3. Vite Config

React itself is bundled by the remote (not shared). Only `@workfort/ui` and `@workfort/auth` are shared singletons. `@workfort/ui-react` is a local dependency.

Create `vite.config.ts`:

```ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { federation } from '@module-federation/vite';

export default defineConfig({
  plugins: [
    react(),
    federation({
      name: 'my-service',
      filename: 'remoteEntry.js',
      exposes: {
        './index': './src/index.tsx',
      },
      shared: {
        '@workfort/ui': { singleton: true, import: false },
        '@workfort/auth': { singleton: true, import: false },
      },
    }),
  ],
  build: {
    target: 'esnext',
    outDir: 'dist',
  },
});
```

Both shared packages use `import: false` so the remote consumes the shell's singletons instead of bundling its own copies.

## 4. Entry Module

Create `src/index.tsx`:

```tsx
import { Panel, Button } from '@workfort/ui-react';
import { useAuth } from '@workfort/ui-react';

export const manifest = {
  name: 'my-service',
  label: 'My Service',
  route: '/my-service',
  minWidth: 320,
};

export default function MyService(props: { connected: boolean }) {
  const { user, isAuthenticated } = useAuth();

  return (
    <Panel label={manifest.label}>
      <div style={{ padding: '1rem' }}>
        {isAuthenticated ? (
          <p>Welcome, {user?.displayName || 'User'}!</p>
        ) : (
          <p>Not authenticated.</p>
        )}
        <p>Service is {props.connected ? 'online' : 'offline'}</p>
        <Button onWfClick={() => console.log('Clicked!')}>
          Action
        </Button>
      </div>
    </Panel>
  );
}
```

**Key differences from SolidJS:**

- Use React wrappers from `@workfort/ui-react` (not raw `wf-*` elements). React 18's synthetic event system doesn't forward custom element events.
- Event props use camelCase: `onWfClick` (not `onwf-click`).
- `useAuth()` returns plain values (`user: User | null`, `isAuthenticated: boolean`), not accessors.

See [`@workfort/ui-react`](../shared-packages.md#workfortuiReact) for all available components and hooks.

## 5. Go Wiring

See [SolidJS Guide — Go Wiring](./solidjs.md#go-wiring). The process is identical regardless of frontend framework.

## 6. Fort Config

See [SolidJS Guide — Fort Config](./solidjs.md#fort-config). Same for React.

## 7. Run

Development:

```bash
pnpm run dev
```

The remote runs on `http://localhost:5173/remoteEntry.js` by default. Configure the shell's `vite.config.ts` to load it:

```ts
remotes: {
  'my-service': 'http://localhost:5173/remoteEntry.js',
}
```

Production build:

```bash
pnpm run build
```

Output is in `dist/`. Deploy per your infrastructure. The shell's build must reference the production URL in `remotes`.

---

**Next:** See [Service Frontend Contract](../service-contract.md) for the complete spec. See [Shared Packages](../shared-packages.md) for the full component and hook reference.
