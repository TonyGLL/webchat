import { describe, expect, it } from 'vitest';
import { nextTabIndex } from './tabs';

describe('nextTabIndex', () => {
  it('moves with the arrows and wraps at the ends', () => {
    expect(nextTabIndex(3, 0, 'ArrowRight')).toBe(1);
    expect(nextTabIndex(3, 2, 'ArrowRight')).toBe(0);
    expect(nextTabIndex(3, 1, 'ArrowLeft')).toBe(0);
    expect(nextTabIndex(3, 0, 'ArrowLeft')).toBe(2);
  });

  it('jumps to the ends with Home and End', () => {
    expect(nextTabIndex(3, 1, 'Home')).toBe(0);
    expect(nextTabIndex(3, 0, 'End')).toBe(2);
  });

  it('ignores other keys and empty lists', () => {
    expect(nextTabIndex(3, 0, 'a')).toBeNull();
    expect(nextTabIndex(3, 0, 'ArrowDown')).toBeNull();
    expect(nextTabIndex(0, 0, 'ArrowRight')).toBeNull();
  });
});
