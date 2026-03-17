import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { idleDetection } from '../src/idle-detection.js';

// Mock onDestroy since we're not inside a Svelte component lifecycle
vi.mock('svelte', () => ({
  onDestroy: (fn: () => void) => { (globalThis as any).__svelteCleanup = fn; },
}));

describe('idleDetection (Svelte)', () => {
  beforeEach(() => { vi.useFakeTimers(); });
  afterEach(() => { vi.useRealTimers(); delete (globalThis as any).__svelteCleanup; });

  it('starts the detector and calls onActive immediately', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    idleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
    expect(onActive).toHaveBeenCalledOnce();
  });

  it('calls onIdle after timeout', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    idleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
    vi.advanceTimersByTime(1001);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  it('registers cleanup via onDestroy', () => {
    const spy = vi.spyOn(document, 'removeEventListener');
    const onActive = vi.fn();
    const onIdle = vi.fn();
    idleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
    const cleanup = (globalThis as any).__svelteCleanup;
    expect(cleanup).toBeDefined();
    cleanup();
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });
});
