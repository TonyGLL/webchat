import { useEffect, useState } from 'react';
import { DirectoryUser, userService } from '@/services/userService';
import { useDebounce } from '@/lib/useDebounce';
import { useSocket } from './useSocket';
import { useSocketEvent } from './useSocketEvents';

const REFRESH_MS = 60_000;

/** Everyone else with live online status; `query` filters by name or username. */
export const useDirectory = (query = '') => {
  const socket = useSocket();
  const q = useDebounce(query.trim(), 250);
  const [people, setPeople] = useState<DirectoryUser[]>([]);
  const [loading, setLoading] = useState(true);

  // Reload when the filter changes, when the connection (re)opens (presence events sent while it
  // was down are lost) and periodically, so newly registered people show up.
  useEffect(() => {
    let cancelled = false;
    const load = () =>
      userService
        .getDirectory(q)
        .then((list) => !cancelled && setPeople(list))
        .catch((error) => console.error('Failed to load people', error))
        .finally(() => !cancelled && setLoading(false));
    load();
    const timer = setInterval(load, REFRESH_MS);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [q, socket]);

  useSocketEvent('presence', ({ user_id, online }) => {
    setPeople((prev) => prev.map((p) => (p.id === user_id ? { ...p, online } : p)));
  });

  return { people, loading };
};
