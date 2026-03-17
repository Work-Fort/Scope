import { onDestroy } from 'svelte';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function idleDetection(opts: IdleDetectorOptions): IdleDetector {
  const detector = new IdleDetector(opts);
  detector.start();
  onDestroy(() => detector.dispose());
  return detector;
}
