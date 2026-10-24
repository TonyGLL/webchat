import { describe, expect, it } from 'vitest';
import { followBehavior, isNearBottom } from './scroll';

describe('isNearBottom', () => {
  it('is true within the threshold and false further up', () => {
    expect(isNearBottom({ scrollTop: 880, clientHeight: 500, scrollHeight: 1500 })).toBe(true); // 120 px away
    expect(isNearBottom({ scrollTop: 800, clientHeight: 500, scrollHeight: 1500 })).toBe(false); // 200 px away
    expect(isNearBottom({ scrollTop: 0, clientHeight: 500, scrollHeight: 400 })).toBe(true); // nothing to scroll
  });
});

describe('followBehavior', () => {
  const base = { firstLoad: false, own: false, nearBottom: true, reducedMotion: false };

  it('jumps instantly on first load, even far from the bottom', () => {
    expect(followBehavior({ ...base, firstLoad: true, nearBottom: false })).toBe('auto');
  });

  it('follows with animation when the reader is at the bottom', () => {
    expect(followBehavior(base)).toBe('smooth');
  });

  it('does not move a reader who is looking at older messages', () => {
    expect(followBehavior({ ...base, nearBottom: false })).toBeNull();
  });

  it('always follows the user\'s own messages', () => {
    expect(followBehavior({ ...base, own: true, nearBottom: false })).toBe('smooth');
  });

  it('never animates with reduced motion', () => {
    expect(followBehavior({ ...base, reducedMotion: true })).toBe('auto');
  });
});
