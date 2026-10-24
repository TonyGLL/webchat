import { useCallback, useMemo, useSyncExternalStore } from 'react';
import { ALERT_PREFS_KEY, AlertPrefs, parseAlertPrefs } from '@/lib/alerts';

const listeners = new Set<() => void>();
const subscribe = (onChange: () => void) => {
  listeners.add(onChange);
  window.addEventListener('storage', onChange); // another tab changed them
  return () => {
    listeners.delete(onChange);
    window.removeEventListener('storage', onChange);
  };
};

const readRaw = () => {
  try {
    return localStorage.getItem(ALERT_PREFS_KEY) ?? '';
  } catch {
    return '';
  }
};

/** Notification preferences, kept per browser in localStorage and shared across the app. */
export const useAlertPrefs = () => {
  // The raw string is the snapshot: strings compare by value, so React sees no false changes.
  const raw = useSyncExternalStore(subscribe, readRaw, () => '');
  const prefs = useMemo(() => parseAlertPrefs(raw), [raw]);

  const update = useCallback(
    (patch: Partial<AlertPrefs>) => {
      try {
        localStorage.setItem(ALERT_PREFS_KEY, JSON.stringify({ ...parseAlertPrefs(readRaw()), ...patch }));
      } catch {
        // storage blocked: the choice just won't persist
      }
      listeners.forEach((l) => l());
    },
    []
  );

  return { prefs, update };
};
