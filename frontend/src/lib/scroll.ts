/** How close to the bottom (px) still counts as "reading the latest messages". */
export const NEAR_BOTTOM_PX = 120;

export const isNearBottom = (el: { scrollTop: number; clientHeight: number; scrollHeight: number }, threshold = NEAR_BOTTOM_PX) =>
  el.scrollHeight - el.scrollTop - el.clientHeight <= threshold;

/**
 * What to do when the newest message changes: `null` = leave the reader where they are.
 * First load always jumps (instantly); the user's own messages always follow; others only when
 * the reader is already at the bottom. Animation is skipped on first load and with reduced motion.
 */
export const followBehavior = (opts: { firstLoad: boolean; own: boolean; nearBottom: boolean; reducedMotion: boolean }): ScrollBehavior | null => {
  if (opts.firstLoad) return 'auto';
  if (!opts.own && !opts.nearBottom) return null;
  return opts.reducedMotion ? 'auto' : 'smooth';
};
