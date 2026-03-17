import { useEffect, useRef } from 'react';
import { IdleDetector, type IdleDetectorOptions } from '@workfort/ui';

export function useIdleDetection(opts: IdleDetectorOptions): void {
  const detectorRef = useRef<IdleDetector | null>(null);

  useEffect(() => {
    const detector = new IdleDetector(opts);
    detector.start();
    detectorRef.current = detector;
    return () => detector.dispose();
  }, []);
}
