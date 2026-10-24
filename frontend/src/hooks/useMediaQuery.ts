import { useSyncExternalStore } from 'react';

/** Live result of a CSS media query (false on the server and during hydration). */
export const useMediaQuery = (query: string) =>
  useSyncExternalStore(
    (notify) => {
      const list = window.matchMedia(query);
      list.addEventListener('change', notify);
      return () => list.removeEventListener('change', notify);
    },
    () => window.matchMedia(query).matches,
    () => false
  );
