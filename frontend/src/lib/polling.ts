/** Whether a periodic or resume-triggered refresh should run now: never while hidden, never twice within `minGapMs`. */
export const shouldRefresh = (opts: { visible: boolean; now: number; lastRun: number; minGapMs: number }) =>
  opts.visible && opts.now - opts.lastRun >= opts.minGapMs;
