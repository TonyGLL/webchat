import { describe, expect, it } from 'vitest';
import { popoverPosition } from './popover';

describe('popoverPosition', () => {
  it('lines up with the right edge of the trigger when there is room', () => {
    expect(popoverPosition({ right: 600, bottom: 40 }, 1280, 256)).toEqual({ top: 44, left: 344, width: 256 });
  });

  it('stays inside the viewport for a trigger near the left edge (the sidebar bell)', () => {
    expect(popoverPosition({ right: 170, bottom: 40 }, 1280, 256).left).toBe(8);
  });

  it('stays inside the viewport for a trigger at the right edge', () => {
    const p = popoverPosition({ right: 1290, bottom: 40 }, 1280, 256);
    expect(p.left + p.width).toBeLessThanOrEqual(1280 - 8);
  });

  it('shrinks on a very narrow screen', () => {
    expect(popoverPosition({ right: 300, bottom: 40 }, 240, 256)).toEqual({ top: 44, left: 8, width: 224 });
  });
});
