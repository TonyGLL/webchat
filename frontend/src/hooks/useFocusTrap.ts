import { useEffect, type RefObject } from 'react';

/**
 * Index to focus after Tab (or Shift+Tab) among `count` focusable elements, wrapping at the
 * ends. `current` is the index of the focused element, or -1 when focus is outside the list.
 * Returns null when there is nothing to focus.
 */
export function nextFocusIndex(count: number, current: number, backwards: boolean): number | null {
  if (count <= 0) return null;
  if (current < 0 || current >= count) return backwards ? count - 1 : 0;
  return backwards ? (current - 1 + count) % count : (current + 1) % count;
}

const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

/**
 * Keeps Tab / Shift+Tab inside `container` while `active`, and gives focus back to the element
 * that had it before (e.g. the button that opened a dialog) when the trap ends. It does not
 * move initial focus: the dialog decides what gets it.
 */
export function useFocusTrap(container: RefObject<HTMLElement | null>, active = true) {
  useEffect(() => {
    if (!active) return;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;

    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab' || !container.current) return;
      const items = Array.from(container.current.querySelectorAll<HTMLElement>(FOCUSABLE));
      const next = nextFocusIndex(items.length, items.indexOf(document.activeElement as HTMLElement), e.shiftKey);
      e.preventDefault();
      if (next !== null) items[next].focus();
    };
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('keydown', onKey);
      if (previous?.isConnected) previous.focus();
    };
  }, [container, active]);
}
