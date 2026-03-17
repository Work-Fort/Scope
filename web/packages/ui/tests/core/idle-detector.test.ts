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
