import { afterEach, describe, expect, it, vi } from 'vitest';
import { createTrailingThrottle } from './trailingThrottle';

afterEach(() => vi.useRealTimers());

describe('createTrailingThrottle', () => {
  it('collapses calls inside a window into one trailing run', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const throttle = createTrailingThrottle(fn, 4000);
    for (let i = 0; i < 20; i++) throttle.call();
    expect(fn).not.toHaveBeenCalled();
    vi.advanceTimersByTime(4000);
    expect(fn).toHaveBeenCalledTimes(1);
    throttle.call();
    vi.advanceTimersByTime(4000);
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it('cancel drops the pending run', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const throttle = createTrailingThrottle(fn, 1000);
    throttle.call();
    throttle.cancel();
    vi.advanceTimersByTime(5000);
    expect(fn).not.toHaveBeenCalled();
    throttle.call();
    vi.advanceTimersByTime(1000);
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
