# Pylon Frontend Integration — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Update the shell frontend to support Pylon forts: adjust polling intervals, add HTTP URL warnings in the fort picker, and add a "Service Status" view accessible from the hamburger menu.

**Architecture:** The frontend is environment-agnostic (same code runs in browser and Tauri). Fort picker shows for multi-fort (Tauri) and auto-navigates for single-fort (scope-server). A new `/forts/:fort/status` route renders a Scope-provided status page showing each service's connectivity, UI availability, and protocol (HTTP vs HTTPS). The data comes from the existing `/forts/{fort}/api/services` endpoint — no new APIs needed.

**Tech Stack:** SolidJS, TypeScript, @workfort/ui web components

**Depends on:** `docs/plans/2026-03-18-pylon-backend.md` (backend must serve Pylon data first)

---

### Task 1: Adjust poll interval for Pylon forts

**Files:**
- Modify: `web/shell/src/stores/services.ts:8`
- Modify: `web/shell/src/lib/api.ts`

**Step 1: Add a way to know if the current fort uses Pylon**

In `web/shell/src/lib/api.ts`, the `FortInfo` type already has `pylon?: string`. The services store needs access to this.

In `web/shell/src/stores/services.ts`, add a signal for the current fort's info and a getter for whether it's a Pylon fort:

```typescript
const [fortInfo, setFortInfo] = createSignal<FortInfo | null>(null);

export const isPylonFort = () => !!fortInfo()?.pylon;
```

**Step 2: Update `startPolling` to use fort info for interval**

Modify `startPolling` in `web/shell/src/stores/services.ts` to accept the fort object and use the right interval:

```typescript
const LOCAL_POLL_INTERVAL = 30_000;
const PYLON_POLL_INTERVAL = 120_000;

export function startPolling(fort: string, info?: FortInfo): void {
  if (activeFort !== fort) {
    stopPolling();
    prevConnected = new Map();
    setServiceList([]);
    setConflictList([]);
    setNeedsAuth(true);
    setSessionChecked(false);
  }
  activeFort = fort;
  if (info) setFortInfo(info);

  const pollInterval = info?.pylon ? PYLON_POLL_INTERVAL : LOCAL_POLL_INTERVAL;

  // ... rest of existing startPolling logic ...

  intervalId = setInterval(() => {
    fetchServices(fort).then(handlePollResult).catch(console.error);
  }, pollInterval);
}
```

**Step 3: Update FortShell in app.tsx to pass fort info**

In `web/shell/src/app.tsx`, `FortShell` calls `startPolling(fort)`. Update it to fetch fort info and pass it:

```typescript
const [forts] = createResource(fetchForts);

onMount(() => {
  const info = forts()?.find((f) => f.name === params.fort);
  startPolling(params.fort, info);
});
```

**Step 4: Verify it builds**

Run: `cd web/shell && pnpm build`
Expected: Builds without errors

**Step 5: Commit**

```bash
git add web/shell/src/stores/services.ts web/shell/src/lib/api.ts web/shell/src/app.tsx
git commit -m "feat(shell): use 2-minute poll interval for Pylon forts"
```

---

### Task 2: HTTP URL warning in fort picker

When any fort's services include `http://` base URLs, show a warning.

**Files:**
- Modify: `web/shell/src/components/fort-picker.tsx`
- Modify: `web/shell/src/lib/api.ts`

**Step 1: Add a pre-check API call**

The fort picker needs to fetch services for each fort to check for HTTP URLs. Add a helper to `web/shell/src/lib/api.ts`:

```typescript
export async function checkFortServices(fort: string): Promise<ServiceInfo[]> {
  try {
    const res = await fetchServices(fort);
    return res.services;
  } catch {
    return [];
  }
}
```

**Step 2: Update fort picker to show HTTP warnings**

In `web/shell/src/components/fort-picker.tsx`:

```typescript
import { createResource, createSignal, Show, For, onMount, type Component } from 'solid-js';
import { Navigate, useNavigate } from '@solidjs/router';
import { fetchForts, checkFortServices, type FortInfo } from '../lib/api';

const FortPicker: Component = () => {
  const [forts] = createResource(fetchForts);
  const [httpWarnings, setHttpWarnings] = createSignal<Record<string, boolean>>({});
  const [checking, setChecking] = createSignal(false);
  const navigate = useNavigate();

  // Pre-check Pylon forts for HTTP URLs
  createEffect(() => {
    const fortList = forts();
    if (!fortList) return;

    const pylonForts = fortList.filter((f) => f.pylon);
    if (pylonForts.length === 0) return;

    setChecking(true);
    Promise.all(
      pylonForts.map(async (f) => {
        const services = await checkFortServices(f.name);
        const hasHttp = services.some((s) =>
          s.base_url?.startsWith('http://'),
        );
        return [f.name, hasHttp] as const;
      }),
    ).then((results) => {
      const warnings: Record<string, boolean> = {};
      for (const [name, hasHttp] of results) {
        if (hasHttp) warnings[name] = true;
      }
      setHttpWarnings(warnings);
      setChecking(false);
    });
  });

  return (
    <Show when={!forts.loading && !checking()} fallback={<wf-skeleton width="100%" height="200px" />}>
      <Show
        when={forts() && forts()!.length !== 1}
        fallback={
          forts() && forts()!.length === 1
            ? <Navigate href={`/forts/${forts()![0].name}`} />
            : <div class="shell-unavailable">No forts configured.</div>
        }
      >
        <div class="fort-picker">
          <h2 class="fort-picker__title">Select a Fort</h2>
          <wf-list>
            <For each={forts()}>
              {(fort) => (
                <wf-list-item on:wf-select={() => navigate(`/forts/${fort.name}`)}>
                  {fort.name}
                  <Show when={httpWarnings()[fort.name]}>
                    <wf-badge color="yellow" slot="suffix">HTTP</wf-badge>
                  </Show>
                </wf-list-item>
              )}
            </For>
          </wf-list>
        </div>
      </Show>
    </Show>
  );
};

export default FortPicker;
```

Note: `base_url` isn't currently in `ServiceInfo`. It needs to be added.

**Step 3: Add `base_url` to `ServiceInfo`**

In `web/shell/src/lib/api.ts`, add to the `ServiceInfo` interface:

```typescript
export interface ServiceInfo {
  name: string;
  label: string;
  route: string;
  enabled: boolean;
  ui: boolean;
  connected: boolean;
  setup_mode?: boolean;
  display?: 'nav' | 'menu';
  base_url?: string;
}
```

**Step 4: Verify it builds**

Run: `cd web/shell && pnpm build`
Expected: Builds without errors

**Step 5: Commit**

```bash
git add web/shell/src/components/fort-picker.tsx web/shell/src/lib/api.ts
git commit -m "feat(shell): HTTP URL warning in fort picker for Pylon forts"
```

---

### Task 3: Create the Service Status page

A Scope-provided view showing each service's status: connected, UI available, protocol, and all manifest fields.

**Files:**
- Create: `web/shell/src/components/service-status.tsx`

**Step 1: Create the status page component**

```typescript
import { type Component, For } from 'solid-js';
import { useParams } from '@solidjs/router';
import { services } from '../stores/services';

const ServiceStatus: Component = () => {
  const params = useParams<{ fort: string }>();

  return (
    <div class="service-status">
      <h2 class="service-status__title">Service Status</h2>
      <p class="service-status__subtitle">Fort: {params.fort}</p>
      <table class="service-status__table">
        <thead>
          <tr>
            <th>Service</th>
            <th>Status</th>
            <th>UI</th>
            <th>Protocol</th>
            <th>Route</th>
            <th>Display</th>
          </tr>
        </thead>
        <tbody>
          <For each={services()}>
            {(svc) => {
              const isHttps = svc.base_url?.startsWith('https://');
              return (
                <tr>
                  <td>
                    <strong>{svc.label}</strong>
                    <br />
                    <small class="service-status__name">{svc.name}</small>
                  </td>
                  <td>
                    <wf-badge color={svc.connected ? 'green' : 'red'}>
                      {svc.connected ? 'Connected' : 'Disconnected'}
                    </wf-badge>
                  </td>
                  <td>
                    <wf-badge color={svc.ui ? 'blue' : 'yellow'}>
                      {svc.ui ? 'Available' : 'No UI'}
                    </wf-badge>
                  </td>
                  <td>
                    <wf-badge color={isHttps ? 'green' : 'yellow'}>
                      {isHttps ? 'HTTPS' : 'HTTP'}
                    </wf-badge>
                  </td>
                  <td><code>{svc.route}</code></td>
                  <td>{svc.display ?? 'nav'}</td>
                </tr>
              );
            }}
          </For>
        </tbody>
      </table>
    </div>
  );
};

export default ServiceStatus;
```

**Step 2: Add styles**

In `web/shell/src/global.css`, add at the end:

```css
/* Service Status */
.service-status {
  padding: var(--wf-space-lg);
  max-width: 800px;
}
.service-status__title {
  font-size: var(--wf-font-size-xl);
  margin-bottom: var(--wf-space-xs);
}
.service-status__subtitle {
  color: var(--wf-color-text-muted);
  margin-bottom: var(--wf-space-lg);
}
.service-status__table {
  width: 100%;
  border-collapse: collapse;
}
.service-status__table th,
.service-status__table td {
  text-align: left;
  padding: var(--wf-space-sm) var(--wf-space-md);
  border-bottom: 1px solid var(--wf-color-border);
}
.service-status__table th {
  color: var(--wf-color-text-muted);
  font-weight: 500;
  font-size: var(--wf-font-size-sm);
}
.service-status__name {
  color: var(--wf-color-text-muted);
  font-family: monospace;
}
```

**Step 3: Verify it builds**

Run: `cd web/shell && pnpm build`
Expected: Builds (component isn't routed yet, but should compile)

**Step 4: Commit**

```bash
git add web/shell/src/components/service-status.tsx web/shell/src/global.css
git commit -m "feat(shell): add Service Status page component"
```

---

### Task 4: Add route for Service Status

**Files:**
- Modify: `web/shell/src/app.tsx`

**Step 1: Import and add the route**

In `web/shell/src/app.tsx`, add the import:

```typescript
import ServiceStatus from './components/service-status';
```

Add the route as a child of the FortShell route, before the catch-all service route:

```tsx
<Route path="/forts/:fort" component={FortShell}>
  <Route path="/status" component={ServiceStatus} />
  <Route path="/:service/*rest" component={ServicePage} />
  <Route path="/" component={FortIndex} />
</Route>
```

**Step 2: Verify it builds**

Run: `cd web/shell && pnpm build`
Expected: Builds

**Step 3: Commit**

```bash
git add web/shell/src/app.tsx
git commit -m "feat(shell): add /forts/:fort/status route"
```

---

### Task 5: Add "Service Status" link to hamburger menu

**Files:**
- Modify: `web/shell/src/components/nav-bar.tsx:85-88`

**Step 1: Add the link above "Sign out"**

In `web/shell/src/components/nav-bar.tsx`, insert the "Service Status" item just above the divider before "Sign out":

```tsx
          <wf-list-item on:wf-select={() => toggleHandedness()}>
            {handedness() === 'right' ? '← Left-handed' : '→ Right-handed'}
          </wf-list-item>
          <wf-divider />
          <wf-list-item on:wf-select={() => { closeHamburger(); navigate(`/forts/${params.fort}/status`); }}>
            Service status
          </wf-list-item>
          <wf-list-item on:wf-select={handleLogout}>
            Sign out
          </wf-list-item>
```

**Step 2: Verify it builds**

Run: `cd web/shell && pnpm build`
Expected: Builds

**Step 3: Commit**

```bash
git add web/shell/src/components/nav-bar.tsx
git commit -m "feat(shell): add Service Status link in hamburger menu"
```

---

### Task 6: Full build + verify

**Step 1: Build everything**

Run: `cd web/shell && pnpm build`
Expected: Successful build

**Step 2: Lint check**

Run: `cd web/shell && pnpm exec tsc --noEmit`
Expected: No type errors

**Step 3: Commit any fixes**
