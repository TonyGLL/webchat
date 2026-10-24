'use client';

import { createContext, useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { api, API_URL } from '@/lib/api';
import { backoffDelay } from '@/lib/backoff';

interface IRawSocketContext {
  /** The open connection, or null while connecting / reconnecting. */
  socket: WebSocket | null;
}

export const SocketContext = createContext<IRawSocketContext>({
  socket: null,
});

const WS_URL = API_URL.replace(/^http/, 'ws') + '/ws';

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { tokens } = useAuth();
  const isAuthenticated = Boolean(tokens);
  const [socket, setSocket] = useState<WebSocket | null>(null);

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
  }, [isAuthenticated]);

  return <SocketContext.Provider value={{ socket }}>{children}</SocketContext.Provider>;
};
