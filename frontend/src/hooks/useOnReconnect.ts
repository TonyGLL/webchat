import { useEffect, useRef } from 'react';
import { useSocket } from './useSocket';

/**
 * Runs `callback` after the websocket reopens following a drop (not on its first open: callers load
 * their data on mount). Events sent while it was down are lost, so callers refetch; a random delay of
 * up to `maxDelayMs` keeps a herd of reconnecting clients from refetching in the same instant.
 */
export const useOnReconnect = (callback: () => void, maxDelayMs = 2000) => {
  const socket = useSocket();
  const latest = useRef(callback);
  useEffect(() => {
    latest.current = callback;
  });
  const wasConnected = useRef(false);
  const hadDrop = useRef(false);

  useEffect(() => {
    if (!socket) {
      if (wasConnected.current) hadDrop.current = true;
      return;
    }
    wasConnected.current = true;
    if (!hadDrop.current) return;
    hadDrop.current = false;
    const timer = setTimeout(() => latest.current(), Math.random() * maxDelayMs);
    return () => clearTimeout(timer);
  }, [socket, maxDelayMs]);
};
