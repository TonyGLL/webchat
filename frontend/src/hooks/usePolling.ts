import { useEffect, useRef } from 'react';
import { shouldRefresh } from '@/lib/polling';

/**
 * Runs `refresh` once now, then every `intervalMs` while the tab is visible, and again as soon as the
 * tab becomes visible or focused (at most once per `minGapMs`). Pass a stable (useCallback) function:
 * a new identity restarts the schedule, which is how callers refetch when their inputs change.
 */
export const usePolling = (refresh: () => void, intervalMs: number, minGapMs = 5_000) => {
  const lastRun = useRef(0);

  useEffect(() => {
    const run = () => {
      lastRun.current = Date.now();
      refresh();
    };
    const runIfDue = () => {
      if (shouldRefresh({ visible: document.visibilityState === 'visible', now: Date.now(), lastRun: lastRun.current, minGapMs })) run();
    };

    run();
    const timer = setInterval(runIfDue, intervalMs);
    document.addEventListener('visibilitychange', runIfDue);
    window.addEventListener('focus', runIfDue);
    return () => {
      clearInterval(timer);
      document.removeEventListener('visibilitychange', runIfDue);
      window.removeEventListener('focus', runIfDue);
    };
  }, [refresh, intervalMs, minGapMs]);
};
