/**
 * Runs `fn` at most once per `waitMs`, after the window: calls inside a window collapse into one
 * trailing run. `cancel` drops a pending run (call it on unmount or when its target changes).
 */
export function createTrailingThrottle(fn: () => void, waitMs: number) {
  let timer: ReturnType<typeof setTimeout> | undefined;
  return {
    call() {
      if (timer !== undefined) return;
      timer = setTimeout(() => {
        timer = undefined;
        fn();
      }, waitMs);
    },
    cancel() {
      clearTimeout(timer);
      timer = undefined;
    },
  };
}
