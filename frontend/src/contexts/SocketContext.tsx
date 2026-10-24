'use client';

import { createContext, useEffect, useMemo, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { api, API_URL } from '@/lib/api';
import { backoffDelay } from '@/lib/backoff';
import { createSocketEventBus, SocketEventBus } from '@/lib/socketEvents';

interface IRawSocketContext {
  /** The open connection, or null while connecting / reconnecting. */
  socket: WebSocket | null;
  /** Frames parsed once and fanned out by event type; stable for the provider's lifetime. */
  events: SocketEventBus;
}

export const SocketContext = createContext<IRawSocketContext>({
  socket: null,
  events: createSocketEventBus(),
});

const WS_URL = API_URL.replace(/^http/, 'ws') + '/ws';

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { tokens } = useAuth();
  const isAuthenticated = Boolean(tokens);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [events] = useState(createSocketEventBus);

  useEffect(() => {
    if (!isAuthenticated) return;

    let cancelled = false;
    let current: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;
    let attempt = 0;

    const scheduleReconnect = () => {
      if (cancelled) return;
      retryTimer = setTimeout(connect, backoffDelay(attempt++));
    };

    const connect = () => {
      // The JWT never goes in the URL: exchange it for a short-lived, single-use ticket.
      api
        .post<{ ticket: string }>('/ws-ticket')
        .then(({ data }) => {
          if (cancelled) return;
          const ws = new WebSocket(`${WS_URL}?ticket=${encodeURIComponent(data.ticket)}`);
          current = ws;

          ws.onopen = () => {
            attempt = 0;
            setSocket(ws);
          };
          ws.onmessage = (event) => events.emit(event.data);
          ws.onclose = () => {
            if (current === ws) setSocket(null);
            scheduleReconnect();
          };
          ws.onerror = () => ws.close();
        })
        .catch(scheduleReconnect);
    };

    connect();

    return () => {
      cancelled = true;
      clearTimeout(retryTimer);
      current?.close();
      setSocket(null);
    };
  }, [isAuthenticated, events]);

  const value = useMemo(() => ({ socket, events }), [socket, events]);

  return <SocketContext.Provider value={value}>{children}</SocketContext.Provider>;
};
