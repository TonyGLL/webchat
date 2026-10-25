import { useCallback, useRef, useState } from 'react';
import { DirectoryUser, userService } from '@/services/userService';
import { useDebounce } from '@/lib/useDebounce';
import { useOnReconnect } from './useOnReconnect';
import { usePolling } from './usePolling';
import { useSocketEvent } from './useSocketEvents';
import { logError } from '@/lib/logError';

const REFRESH_MS = 60_000;

/** Everyone else with live online status; `query` filters by name or username. */
export const useDirectory = (query = '') => {
  const q = useDebounce(query.trim(), 250);
  const [people, setPeople] = useState<DirectoryUser[]>([]);
  const [loading, setLoading] = useState(true);
  const latestRequest = useRef(0);

  // Reload when the filter changes, when the connection reopens after a drop (presence events sent
  // while it was down are lost) and periodically while the tab is visible, so newly registered people show up.
  const load = useCallback(() => {
    const request = ++latestRequest.current;
    const isLatest = () => request === latestRequest.current;
    userService
      .getDirectory(q)
      .then((list) => isLatest() && setPeople(list))
      .catch((error) => logError('Failed to load people', error))
      .finally(() => isLatest() && setLoading(false));
  }, [q]);
  usePolling(load, REFRESH_MS);
  useOnReconnect(load);

  useSocketEvent('presence', ({ user_id, online }) => {
    // Presence of people who are not listed, or no change at all, keeps the same array (no re-render).
    setPeople((prev) =>
      prev.some((p) => p.id === user_id && p.online !== online) ? prev.map((p) => (p.id === user_id ? { ...p, online } : p)) : prev
    );
  });

  return { people, loading };
};
