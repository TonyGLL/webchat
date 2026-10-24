/**
 * Index of the tab to activate after `key` on a tablist (WAI-ARIA tabs pattern, automatic
 * activation): arrows wrap around, Home/End jump to the ends. Returns null for other keys
 * or an empty list.
 */
export function nextTabIndex(count: number, current: number, key: string): number | null {
  if (count <= 0) return null;
  switch (key) {
    case 'ArrowRight':
      return (current + 1) % count;
    case 'ArrowLeft':
      return (current - 1 + count) % count;
    case 'Home':
      return 0;
    case 'End':
      return count - 1;
    default:
      return null;
  }
}
