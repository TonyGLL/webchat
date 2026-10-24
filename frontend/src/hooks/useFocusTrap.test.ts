import { describe, expect, it } from 'vitest';
import { nextFocusIndex } from './useFocusTrap';

describe('nextFocusIndex', () => {
  it('moves forward and wraps to the first element', () => {
    expect(nextFocusIndex(3, 0, false)).toBe(1);
    expect(nextFocusIndex(3, 2, false)).toBe(0);
  });

  it('moves backwards and wraps to the last element', () => {
    expect(nextFocusIndex(3, 2, true)).toBe(1);
    expect(nextFocusIndex(3, 0, true)).toBe(2);
  });

  it('pulls focus in from outside the list', () => {
    expect(nextFocusIndex(3, -1, false)).toBe(0);
    expect(nextFocusIndex(3, -1, true)).toBe(2);
  });

  it('stays on a single element and handles an empty list', () => {
    expect(nextFocusIndex(1, 0, false)).toBe(0);
    expect(nextFocusIndex(1, 0, true)).toBe(0);
    expect(nextFocusIndex(0, -1, false)).toBeNull();
  });
});
