# Framework Hook Adapters — Plan 7

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `useIdleDetection` and `usePermissions` hooks to all four framework adapters (`ui-solid`, `ui-react`, `ui-svelte`, `ui-vue`), wrapping the core `IdleDetector` and `PermissionSet` classes from `@workfort/ui`.

**Architecture:** Each adapter imports the core class from `@workfort/ui` and wraps it in the framework's lifecycle and reactivity system. The core logic is shared — only the glue code differs.

**Tech Stack:** SolidJS, React, Svelte, Vue, TypeScript

**Repo:** `scope/lead` (packages at `web/packages/{solid,react,svelte,vue}/`)

**Prerequisite:** Plan 5 (core classes) must be complete.

---

### Task 1: `@workfort/ui-solid` — useIdleDetection + usePermissions

**Files:**
- Create: `web/packages/solid/src/use-idle-detection.ts`
- Create: `web/packages/solid/src/use-permissions.ts`
- Modify: `web/packages/solid/src/index.ts` (add exports)

**useIdleDetection:**

```typescript
import { onCleanup } from 'solid-js';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function useIdleDetection(opts: IdleDetectorOptions): IdleDetector {
  const detector = new IdleDetector(opts);
  detector.start();
  onCleanup(() => detector.dispose());
  return detector;
}
```

**usePermissions:**

```typescript
import { createSignal } from 'solid-js';
import { PermissionSet } from '@workfort/ui';

export function usePermissions(initial: string[] = []) {
  const core = new PermissionSet(initial);
  const [permissions, setPermissions] = createSignal<string[]>(initial);

  function can(permission: string): boolean {
    return core.can(permission);
  }

  function update(perms: string[]): void {
    core.update(perms);
    setPermissions(perms);
  }

  return { can, update, permissions };
}
```

Add both to `index.ts` exports.

**Commit:** `feat(@workfort/ui-solid): add useIdleDetection and usePermissions hooks`

---

### Task 2: `@workfort/ui-react` — useIdleDetection + usePermissions

**Files:**
- Create: `web/packages/react/src/use-idle-detection.ts`
- Create: `web/packages/react/src/use-permissions.ts`
- Modify: `web/packages/react/src/index.tsx` (add exports)

**useIdleDetection:**

```typescript
import { useEffect, useRef } from 'react';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function useIdleDetection(opts: IdleDetectorOptions): void {
  const detectorRef = useRef<IdleDetector | null>(null);

  useEffect(() => {
    const detector = new IdleDetector(opts);
    detector.start();
    detectorRef.current = detector;
    return () => detector.dispose();
  }, []);
}
```

**usePermissions:**

```typescript
import { useState, useCallback } from 'react';
import { PermissionSet } from '@workfort/ui';

export function usePermissions(initial: string[] = []) {
  const [core] = useState(() => new PermissionSet(initial));
  const [, setVersion] = useState(0); // trigger re-render on update

  const can = useCallback((permission: string) => core.can(permission), [core]);

  const update = useCallback((perms: string[]) => {
    core.update(perms);
    setVersion((v) => v + 1);
  }, [core]);

  return { can, update };
}
```

**Commit:** `feat(@workfort/ui-react): add useIdleDetection and usePermissions hooks`

---

### Task 3: `@workfort/ui-svelte` — idleDetection + permissions

**Files:**
- Create: `web/packages/svelte/src/idle-detection.ts`
- Create: `web/packages/svelte/src/permissions.ts`
- Modify: `web/packages/svelte/src/index.ts` (add exports)

**idleDetection:**

```typescript
import { onDestroy } from 'svelte';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function idleDetection(opts: IdleDetectorOptions): IdleDetector {
  const detector = new IdleDetector(opts);
  detector.start();
  onDestroy(() => detector.dispose());
  return detector;
}
```

**permissions:**

```typescript
import { writable, derived } from 'svelte/store';
import { PermissionSet } from '@workfort/ui';

export function createPermissions(initial: string[] = []) {
  const core = new PermissionSet(initial);
  const store = writable(initial);

  function can(permission: string): boolean {
    return core.can(permission);
  }

  function update(perms: string[]) {
    core.update(perms);
    store.set(perms);
  }

  return { can, update, subscribe: store.subscribe };
}
```

**Commit:** `feat(@workfort/ui-svelte): add idleDetection and permissions`

---

### Task 4: `@workfort/ui-vue` — useIdleDetection + usePermissions

**Files:**
- Create: `web/packages/vue/src/use-idle-detection.ts`
- Create: `web/packages/vue/src/use-permissions.ts`
- Modify: `web/packages/vue/src/index.ts` (add exports)

**useIdleDetection:**

```typescript
import { onMounted, onUnmounted } from 'vue';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function useIdleDetection(opts: IdleDetectorOptions): IdleDetector {
  const detector = new IdleDetector(opts);
  onMounted(() => detector.start());
  onUnmounted(() => detector.dispose());
  return detector;
}
```

**usePermissions:**

```typescript
import { ref, computed } from 'vue';
import { PermissionSet } from '@workfort/ui';

export function usePermissions(initial: string[] = []) {
  const core = new PermissionSet(initial);
  const version = ref(0);

  function can(permission: string): boolean {
    version.value; // track reactivity
    return core.can(permission);
  }

  function update(perms: string[]) {
    core.update(perms);
    version.value++;
  }

  return { can, update };
}
```

**Commit:** `feat(@workfort/ui-vue): add useIdleDetection and usePermissions composables`
