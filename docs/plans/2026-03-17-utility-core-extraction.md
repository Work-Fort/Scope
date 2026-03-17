# Utility + Core Class Extraction — Plan 5

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extract pure utility functions and core classes from Sharkfin into `@workfort/ui` so they're reusable across all WorkFort services and framework adapters.

**Architecture:** Pure functions go into `@workfort/ui/src/utils/`. Framework-agnostic core classes (`IdleDetector`, `PermissionSet`) go into `@workfort/ui/src/core/`. All are exported from the package index. No Lit, no framework dependency — just TypeScript.

**Tech Stack:** TypeScript, vitest, `@workfort/ui`

**Repo:** `scope/lead` (package at `web/packages/ui/`)

---

### Task 1: Extract Utility Functions

**Files:**
- Create: `web/packages/ui/src/utils/initials.ts`
- Create: `web/packages/ui/src/utils/time.ts`
- Create: `web/packages/ui/src/utils/throttle.ts`
- Create: `web/packages/ui/src/utils/index.ts`
- Create: `web/packages/ui/tests/utils/initials.test.ts`
- Create: `web/packages/ui/tests/utils/time.test.ts`
- Create: `web/packages/ui/tests/utils/throttle.test.ts`
- Modify: `web/packages/ui/src/index.ts` (add exports)

**Step 1: Write failing tests**

Create `web/packages/ui/tests/utils/initials.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { initials } from '../../src/utils/initials.js';

describe('initials', () => {
  it('extracts two-part initials from hyphenated name', () => {
    expect(initials('alice-chen')).toBe('AC');
  });

  it('extracts two-part initials from underscored name', () => {
    expect(initials('bob_kim')).toBe('BK');
  });

  it('extracts first two chars for single-word name', () => {
    expect(initials('bob')).toBe('BO');
  });

  it('uppercases result', () => {
    expect(initials('alice-chen')).toBe('AC');
  });

  it('handles dotted names', () => {
    expect(initials('j.doe')).toBe('JD');
  });
});
```

Create `web/packages/ui/tests/utils/time.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { formatTime, formatDateLabel, isSameDay } from '../../src/utils/time.js';

describe('formatTime', () => {
  it('formats ISO to HH:MM', () => {
    const result = formatTime('2026-03-15T09:14:00Z');
    expect(result).toMatch(/\d{2}:\d{2}/);
  });
});

describe('formatDateLabel', () => {
  it('returns date string for old dates', () => {
    const result = formatDateLabel('2020-01-01T12:00:00Z');
    expect(result).toContain('2020');
  });
});

describe('isSameDay', () => {
  it('returns true for same day', () => {
    expect(isSameDay('2026-03-15T09:00:00Z', '2026-03-15T23:00:00Z')).toBe(true);
  });

  it('returns false for different days', () => {
    expect(isSameDay('2026-03-13T12:00:00Z', '2026-03-15T12:00:00Z')).toBe(false);
  });
});
```

Create `web/packages/ui/tests/utils/throttle.test.ts`:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { throttle } from '../../src/utils/throttle.js';

describe('throttle', () => {
  it('calls function immediately on first invocation', () => {
    const fn = vi.fn();
    const throttled = throttle(fn, 1000);
    throttled();
    expect(fn).toHaveBeenCalledOnce();
  });

  it('suppresses calls within the throttle window', () => {
    const fn = vi.fn();
    const throttled = throttle(fn, 1000);
    throttled();
    throttled();
    throttled();
    expect(fn).toHaveBeenCalledOnce();
  });

  it('allows calls after the throttle window', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const throttled = throttle(fn, 100);
    throttled();
    vi.advanceTimersByTime(101);
    throttled();
    expect(fn).toHaveBeenCalledTimes(2);
    vi.useRealTimers();
  });
});
```

**Step 2: Run tests to verify they fail**

Run: `cd web/packages/ui && npx vitest run tests/utils/`
Expected: FAIL — modules not found.

**Step 3: Implement utilities**

Create `web/packages/ui/src/utils/initials.ts`:

```typescript
/** Extract initials from a username like "alice-chen" → "AC" or "bob" → "BO". */
export function initials(username: string): string {
  const parts = username.split(/[-_.\s]+/);
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
  return username.slice(0, 2).toUpperCase();
}
```

Create `web/packages/ui/src/utils/time.ts`:

```typescript
/** Format ISO timestamp to HH:MM. */
export function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
}

/** Format ISO timestamp to a human-readable date label. */
export function formatDateLabel(iso: string): string {
  const d = new Date(iso);
  const today = new Date();
  if (d.toDateString() === today.toDateString()) return 'Today';
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  if (d.toDateString() === yesterday.toDateString()) return 'Yesterday';
  return d.toLocaleDateString(undefined, { month: 'long', day: 'numeric', year: 'numeric' });
}

/** Check if two ISO timestamps fall on the same calendar day (local time). */
export function isSameDay(a: string, b: string): boolean {
  return new Date(a).toDateString() === new Date(b).toDateString();
}
```

Create `web/packages/ui/src/utils/throttle.ts`:

```typescript
/** Throttle a function to execute at most once per `ms` milliseconds. */
export function throttle(fn: () => void, ms: number): () => void {
  let last = 0;
  return () => {
    const now = Date.now();
    if (now - last >= ms) {
      last = now;
      fn();
    }
  };
}
```

Create `web/packages/ui/src/utils/index.ts`:

```typescript
export { initials } from './initials.js';
export { formatTime, formatDateLabel, isSameDay } from './time.js';
export { throttle } from './throttle.js';
```

**Step 4: Add exports to package index**

In `web/packages/ui/src/index.ts`, add before the component imports:

```typescript
// Utilities
export { initials, formatTime, formatDateLabel, isSameDay, throttle } from './utils/index.js';
```

**Step 5: Run tests to verify they pass**

Run: `cd web/packages/ui && npx vitest run tests/utils/`
Expected: PASS.

Run: `cd web/packages/ui && npx vitest run`
Expected: All tests pass (existing + new).

**Step 6: Commit**

```bash
git add web/packages/ui/src/utils/ web/packages/ui/tests/utils/ web/packages/ui/src/index.ts
git commit -m "feat(@workfort/ui): extract utility functions — initials, time formatting, throttle"
```

---

### Task 2: Extract `IdleDetector` Core Class

**Files:**
- Create: `web/packages/ui/src/core/idle-detector.ts`
- Create: `web/packages/ui/src/core/index.ts`
- Create: `web/packages/ui/tests/core/idle-detector.test.ts`
- Modify: `web/packages/ui/src/index.ts` (add exports)

**Step 1: Write failing test**

Create `web/packages/ui/tests/core/idle-detector.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { IdleDetector } from '../../src/core/idle-detector.js';

describe('IdleDetector', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('calls onActive on start', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    const detector = new IdleDetector({ onActive, onIdle, timeout: 1000, throttle: 100 });
    detector.start();
    expect(onActive).toHaveBeenCalledOnce();
    detector.dispose();
  });

  it('calls onIdle after timeout', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    const detector = new IdleDetector({ onActive, onIdle, timeout: 1000, throttle: 100 });
    detector.start();
    vi.advanceTimersByTime(1001);
    expect(onIdle).toHaveBeenCalledOnce();
    detector.dispose();
  });

  it('resets idle timer on activity', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    const detector = new IdleDetector({ onActive, onIdle, timeout: 1000, throttle: 100 });
    detector.start();

    vi.advanceTimersByTime(500);
    // Simulate activity.
    document.dispatchEvent(new Event('click'));
    vi.advanceTimersByTime(101); // past throttle window

    vi.advanceTimersByTime(500);
    expect(onIdle).not.toHaveBeenCalled();

    detector.dispose();
  });

  it('cleans up listeners on dispose', () => {
    const spy = vi.spyOn(document, 'removeEventListener');
    const detector = new IdleDetector({ onActive: vi.fn(), onIdle: vi.fn() });
    detector.start();
    detector.dispose();
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web/packages/ui && npx vitest run tests/core/idle-detector.test.ts`
Expected: FAIL — module not found.

**Step 3: Implement IdleDetector**

Create `web/packages/ui/src/core/idle-detector.ts`:

```typescript
import { throttle } from '../utils/throttle.js';

export interface IdleDetectorOptions {
  onActive: () => void;
  onIdle: () => void;
  /** Milliseconds before idle. Default: 300000 (5 minutes). */
  timeout?: number;
  /** Milliseconds between activity checks. Default: 30000. */
  throttle?: number;
}

/**
 * Framework-agnostic activity tracker.
 * Monitors user input events and calls onActive/onIdle callbacks.
 * Call start() to begin tracking, dispose() to clean up.
 */
export class IdleDetector {
  private opts: Required<IdleDetectorOptions>;
  private timer: ReturnType<typeof setTimeout> | null = null;
  private throttledActivity: (() => void) | null = null;
  private onVisibility: (() => void) | null = null;
  private started = false;

  private static EVENTS = ['mousemove', 'keydown', 'click', 'scroll'] as const;

  constructor(opts: IdleDetectorOptions) {
    this.opts = {
      onActive: opts.onActive,
      onIdle: opts.onIdle,
      timeout: opts.timeout ?? 5 * 60 * 1000,
      throttle: opts.throttle ?? 30_000,
    };
  }

  start(): void {
    if (this.started) return;
    this.started = true;

    this.setActive();

    this.throttledActivity = throttle(() => this.setActive(), this.opts.throttle);
    IdleDetector.EVENTS.forEach((e) =>
      document.addEventListener(e, this.throttledActivity!, { passive: true }),
    );

    this.onVisibility = () => {
      if (document.hidden) {
        this.opts.onIdle();
      } else {
        this.setActive();
      }
    };
    document.addEventListener('visibilitychange', this.onVisibility);
  }

  stop(): void {
    if (!this.started) return;
    if (this.timer) clearTimeout(this.timer);
    this.timer = null;
  }

  dispose(): void {
    this.stop();
    this.started = false;
    if (this.throttledActivity) {
      IdleDetector.EVENTS.forEach((e) =>
        document.removeEventListener(e, this.throttledActivity!),
      );
      this.throttledActivity = null;
    }
    if (this.onVisibility) {
      document.removeEventListener('visibilitychange', this.onVisibility);
      this.onVisibility = null;
    }
  }

  private setActive(): void {
    if (this.timer) clearTimeout(this.timer);
    this.opts.onActive();
    this.timer = setTimeout(() => this.opts.onIdle(), this.opts.timeout);
  }
}
```

Create `web/packages/ui/src/core/index.ts`:

```typescript
export { IdleDetector } from './idle-detector.js';
export type { IdleDetectorOptions } from './idle-detector.js';
```

**Step 4: Add exports to package index**

In `web/packages/ui/src/index.ts`, add:

```typescript
// Core classes
export { IdleDetector } from './core/index.js';
export type { IdleDetectorOptions } from './core/index.js';
```

**Step 5: Run tests**

Run: `cd web/packages/ui && npx vitest run tests/core/`
Expected: PASS.

**Step 6: Commit**

```bash
git add web/packages/ui/src/core/ web/packages/ui/tests/core/ web/packages/ui/src/index.ts
git commit -m "feat(@workfort/ui): extract IdleDetector core class"
```

---

### Task 3: Extract `PermissionSet` Core Class

**Files:**
- Create: `web/packages/ui/src/core/permission-set.ts`
- Create: `web/packages/ui/tests/core/permission-set.test.ts`
- Modify: `web/packages/ui/src/core/index.ts` (add export)
- Modify: `web/packages/ui/src/index.ts` (add export)

**Step 1: Write failing test**

Create `web/packages/ui/tests/core/permission-set.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { PermissionSet } from '../../src/core/permission-set.js';

describe('PermissionSet', () => {
  it('reports permissions correctly', () => {
    const perms = new PermissionSet(['send_message', 'channel_list']);
    expect(perms.can('send_message')).toBe(true);
    expect(perms.can('channel_list')).toBe(true);
    expect(perms.can('manage_roles')).toBe(false);
  });

  it('starts empty', () => {
    const perms = new PermissionSet([]);
    expect(perms.can('anything')).toBe(false);
  });

  it('updates permissions', () => {
    const perms = new PermissionSet([]);
    perms.update(['send_message']);
    expect(perms.can('send_message')).toBe(true);
  });

  it('returns all permissions', () => {
    const perms = new PermissionSet(['a', 'b']);
    expect(perms.all()).toEqual(['a', 'b']);
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web/packages/ui && npx vitest run tests/core/permission-set.test.ts`
Expected: FAIL.

**Step 3: Implement PermissionSet**

Create `web/packages/ui/src/core/permission-set.ts`:

```typescript
/**
 * Framework-agnostic permission checker.
 * Holds a set of permission strings and provides a can() method.
 * Framework adapters wrap this in reactive state.
 */
export class PermissionSet {
  private perms: Set<string>;

  constructor(permissions: string[]) {
    this.perms = new Set(permissions);
  }

  /** Check if a permission is granted. */
  can(permission: string): boolean {
    return this.perms.has(permission);
  }

  /** Replace all permissions. */
  update(permissions: string[]): void {
    this.perms = new Set(permissions);
  }

  /** Return all permissions as an array. */
  all(): string[] {
    return [...this.perms];
  }
}
```

**Step 4: Update exports**

In `web/packages/ui/src/core/index.ts`, add:

```typescript
export { PermissionSet } from './permission-set.js';
```

In `web/packages/ui/src/index.ts`, update the core export:

```typescript
export { IdleDetector, PermissionSet } from './core/index.js';
export type { IdleDetectorOptions } from './core/index.js';
```

**Step 5: Run tests**

Run: `cd web/packages/ui && npx vitest run`
Expected: All tests pass.

**Step 6: Build**

Run: `cd web/packages/ui && pnpm build`
Expected: Build succeeds with new exports.

**Step 7: Commit**

```bash
git add web/packages/ui/src/core/ web/packages/ui/tests/core/ web/packages/ui/src/index.ts
git commit -m "feat(@workfort/ui): extract PermissionSet core class"
```
