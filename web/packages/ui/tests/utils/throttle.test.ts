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
