import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { createRoot } from 'solid-js';
import { useIdleDetection } from '../src/use-idle-detection.js';

describe('useIdleDetection (Solid)', () => {
  beforeEach(() => { vi.useFakeTimers(); });
  afterEach(() => { vi.useRealTimers(); });

  it('starts the detector and calls onActive immediately', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    createRoot((dispose) => {
      useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
      expect(onActive).toHaveBeenCalledOnce();
      dispose();
    });
  });

  it('calls onIdle after timeout', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    createRoot((dispose) => {
      useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
      vi.advanceTimersByTime(1001);
      expect(onIdle).toHaveBeenCalledOnce();
      dispose();
    });
  });

  it('disposes detector on cleanup', () => {
    const spy = vi.spyOn(document, 'removeEventListener');
    const onActive = vi.fn();
    const onIdle = vi.fn();
    createRoot((dispose) => {
      useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
      dispose();
    });
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });
});
