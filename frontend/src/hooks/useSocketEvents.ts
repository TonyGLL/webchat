import { useContext, useEffect, useRef } from 'react';
import { SocketContext } from '@/contexts/SocketContext';
import type { SocketEventHandler, SocketEventType } from '@/lib/socketEvents';

/**
 * Runs `handler` for every incoming event of `type` (each frame is parsed once, in SocketContext).
 * The latest handler is always used, so callers need not memoize it; the subscription survives reconnects.
 */
export const useSocketEvent = <T extends SocketEventType>(type: T, handler: SocketEventHandler<T>) => {
  const { events } = useContext(SocketContext);
  const latest = useRef(handler);
  useEffect(() => {
    latest.current = handler;
  });
  useEffect(() => events.subscribe(type, (payload) => latest.current(payload)), [events, type]);
};
