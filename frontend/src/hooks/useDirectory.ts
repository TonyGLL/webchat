import { useCallback, useRef, useState } from 'react';
import { DirectoryUser, userService } from '@/services/userService';
import { useDebounce } from '@/lib/useDebounce';
import { useSocket } from './useSocket';
import { usePolling } from './usePolling';
import { useSocketEvent } from './useSocketEvents';
import { logError } from '@/lib/logError';

const REFRESH_MS = 60_000;

/** Everyone else with live online status; `query` filters by name or username. */
export const useDirectory = (query = '') => {
  const socket = useSocket();
  const q = useDebounce(query.trim(), 250);
  const [people, setPeople] = useState<DirectoryUser[]>([]);
  const [loading, setLoading] = useState(true);
  const latestRequest = useRef(0);

  // Reload when the filter changes, when the connection (re)opens (presence events sent while it
  // was down are lost) and periodically while the tab is visible, so newly registered people show up.
  // `socket` is a dependency on purpose: a new connection gives `load` a new identity and so a reload.
  const load = useCallback(() => {
    void socket;
    const request = ++latestRequest.current;
    const isLatest = () => request === latestRequest.current;
    userService
      .getDirectory(q)
      .then((list) => isLatest() && setPeople(list))
      .catch((error) => logError('Failed to load people', error))
      .finally(() => isLatest() && setLoading(false));
  }, [q, socket]);
  usePolling(load, REFRESH_MS);

  useSocketEvent('presence', ({ user_id, online }) => {
    setPeople((prev) => prev.map((p) => (p.id === user_id ? { ...p, online } : p)));
  });

  return { people, loading };
};
