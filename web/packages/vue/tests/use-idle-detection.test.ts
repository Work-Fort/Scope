import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useIdleDetection } from '../src/use-idle-detection.js';

// Mock Vue lifecycle hooks since we're not inside a component setup
const cleanups: (() => void)[] = [];
vi.mock('vue', async () => {
  const actual = await vi.importActual<typeof import('vue')>('vue');
  return {
    ...actual,
    onMounted: (fn: () => void) => { fn(); },
    onUnmounted: (fn: () => void) => { cleanups.push(fn); },
  };
});

describe('useIdleDetection (Vue)', () => {
  beforeEach(() => { vi.useFakeTimers(); cleanups.length = 0; });
  afterEach(() => { vi.useRealTimers(); });

  it('starts detector and calls onActive on mount', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
    expect(onActive).toHaveBeenCalledOnce();
  });

  it('calls onIdle after timeout', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
    vi.advanceTimersByTime(1001);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  it('disposes detector on unmount', () => {
    const spy = vi.spyOn(document, 'removeEventListener');
    const onActive = vi.fn();
    const onIdle = vi.fn();
    useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
    expect(cleanups.length).toBe(1);
    cleanups[0]();
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });
});
