import { describe, expect, it } from 'vitest';
import { backoffDelay, reconnectDelay } from './backoff';

describe('backoffDelay', () => {
  it('doubles the upper bound with each attempt', () => {
    const max = () => 1; // the top of the jitter range
    expect([0, 1, 2, 3].map((n) => backoffDelay(n, max))).toEqual([1000, 2000, 4000, 8000]);
  });

  it('is capped at 30 seconds', () => {
    expect(backoffDelay(20, () => 1)).toBe(30000);
  });

  it('uses full jitter: anywhere between 0 and the delay', () => {
    expect(backoffDelay(2, () => 0)).toBe(0);
    expect(backoffDelay(2, () => 0.5)).toBe(2000);
    expect(backoffDelay(2, () => 1)).toBe(4000);
  });
});

describe('reconnectDelay', () => {
  it('spreads clients over 0-5 s after a server shutdown (1012), whatever the attempt', () => {
    expect(reconnectDelay(0, 1012, () => 0)).toBe(0);
    expect(reconnectDelay(0, 1012, () => 1)).toBe(5000);
    expect(reconnectDelay(7, 1012, () => 0.5)).toBe(2500);
  });

  it('uses the exponential backoff for any other close', () => {
    expect(reconnectDelay(2, 1006, () => 1)).toBe(4000);
    expect(reconnectDelay(2, undefined, () => 1)).toBe(4000);
  });
});
