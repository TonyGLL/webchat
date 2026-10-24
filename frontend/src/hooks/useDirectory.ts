import { useEffect, useState } from 'react';
import { DirectoryUser, userService } from '@/services/userService';
import { useDebounce } from '@/lib/useDebounce';
import { useSocket } from './useSocket';

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

  useEffect(() => {
    if (!socket) return;
    const onMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type !== 'presence') return;
        const { user_id, online } = data.payload as { user_id: number; online: boolean };
        setPeople((prev) => prev.map((p) => (p.id === user_id ? { ...p, online } : p)));
      } catch {
        // not a presence event
      }
    };
    socket.addEventListener('message', onMessage);
    return () => socket.removeEventListener('message', onMessage);
  }, [socket]);

  return { people, loading };
};
