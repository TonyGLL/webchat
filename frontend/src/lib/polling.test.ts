import { describe, expect, it } from 'vitest';
import { shouldRefresh } from './polling';

describe('shouldRefresh', () => {
  it('never refreshes while the tab is hidden', () => {
    expect(shouldRefresh({ visible: false, now: 100_000, lastRun: 0, minGapMs: 1000 })).toBe(false);
  });

  it('waits for the minimum gap since the last run', () => {
    expect(shouldRefresh({ visible: true, now: 1500, lastRun: 1000, minGapMs: 5000 })).toBe(false);
    expect(shouldRefresh({ visible: true, now: 6000, lastRun: 1000, minGapMs: 5000 })).toBe(true);
  });
});
