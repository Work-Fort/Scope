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
