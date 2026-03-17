import { onMounted, onUnmounted } from 'vue';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function useIdleDetection(opts: IdleDetectorOptions): IdleDetector {
  const detector = new IdleDetector(opts);
  onMounted(() => detector.start());
  onUnmounted(() => detector.dispose());
  return detector;
}
