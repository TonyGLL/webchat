import { describe, expect, it } from 'vitest';
import { backoffDelay } from './backoff';

describe('backoffDelay', () => {
  it('doubles with each attempt', () => {
    const max = () => 1; // no jitter reduction
    expect([0, 1, 2, 3].map((n) => backoffDelay(n, max))).toEqual([1000, 2000, 4000, 8000]);
  });

  it('is capped at 30 seconds', () => {
    expect(backoffDelay(20, () => 1)).toBe(30000);
  });

  it('applies jitter between 50 % and 100 %', () => {
    expect(backoffDelay(2, () => 0)).toBe(2000);
    expect(backoffDelay(2, () => 1)).toBe(4000);
  });
});
