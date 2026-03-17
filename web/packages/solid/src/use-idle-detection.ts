import { onCleanup } from 'solid-js';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function useIdleDetection(opts: IdleDetectorOptions): IdleDetector {
  const detector = new IdleDetector(opts);
  detector.start();
  onCleanup(() => detector.dispose());
  return detector;
}
