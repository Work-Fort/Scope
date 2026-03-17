import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import React from 'react';
import { render, cleanup } from '@testing-library/react';
import { useIdleDetection } from '../src/use-idle-detection.js';

function TestComponent({ onActive, onIdle }: { onActive: () => void; onIdle: () => void }) {
  useIdleDetection({ onActive, onIdle, timeout: 1000, throttle: 100 });
  return <div data-testid="idle">mounted</div>;
}

describe('useIdleDetection (React)', () => {
  beforeEach(() => { vi.useFakeTimers(); });
  afterEach(() => { cleanup(); vi.useRealTimers(); });

  it('starts detector and calls onActive on mount', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    render(<TestComponent onActive={onActive} onIdle={onIdle} />);
    expect(onActive).toHaveBeenCalledOnce();
  });

  it('calls onIdle after timeout', () => {
    const onActive = vi.fn();
    const onIdle = vi.fn();
    render(<TestComponent onActive={onActive} onIdle={onIdle} />);
    vi.advanceTimersByTime(1001);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  it('disposes detector on unmount', () => {
    const spy = vi.spyOn(document, 'removeEventListener');
    const onActive = vi.fn();
    const onIdle = vi.fn();
    const { unmount } = render(<TestComponent onActive={onActive} onIdle={onIdle} />);
    unmount();
    expect(spy).toHaveBeenCalled();
    spy.mockRestore();
  });
});
