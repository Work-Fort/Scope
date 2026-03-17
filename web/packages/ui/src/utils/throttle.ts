/** Throttle a function to execute at most once per `ms` milliseconds. */
export function throttle(fn: () => void, ms: number): () => void {
  let last = 0;
  return () => {
    const now = Date.now();
    if (now - last >= ms) {
      last = now;
      fn();
    }
  };
}
