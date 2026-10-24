export interface AnchorRect {
  right: number;
  bottom: number;
}

const MARGIN = 8;

/**
 * Where to place a popover under its trigger. It tries to line up its right edge with the
 * trigger's, then slides to stay inside the viewport (and shrinks on very narrow screens), so it
 * is never cut off whatever side of the screen the trigger is on.
 */
export function popoverPosition(anchor: AnchorRect, viewportWidth: number, width: number, gap = 4) {
  const w = Math.min(width, Math.max(0, viewportWidth - MARGIN * 2));
  const left = Math.min(Math.max(anchor.right - w, MARGIN), viewportWidth - w - MARGIN);
  return { top: anchor.bottom + gap, left, width: w };
}
