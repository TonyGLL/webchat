'use client';

import { createContext, useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { api, API_URL } from '@/lib/api';

interface IRawSocketContext {
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
    let socketInstance: WebSocket | null = null;

    // The JWT never goes in the URL: exchange it for a short-lived, single-use ticket.
    api
      .post<{ ticket: string }>('/ws-ticket')
      .then(({ data }) => {
        if (cancelled) return;
        socketInstance = new WebSocket(`${WS_URL}?ticket=${encodeURIComponent(data.ticket)}`);

        socketInstance.onopen = () => {
          console.log('WebSocket connection established.');
          setSocket(socketInstance);
        };

        socketInstance.onclose = (event) => {
          console.log('WebSocket connection closed:', event.code, event.reason);
          setSocket(null);
        };

        socketInstance.onerror = (error) => {
          console.error('WebSocket error:', error);
        };
      })
      .catch((error) => console.error('Failed to get websocket ticket:', error));

    return () => {
      cancelled = true;
      socketInstance?.close();
    };
  }, [isAuthenticated]);

  return (
    <SocketContext.Provider value={{ socket }}>
      {children}
    </SocketContext.Provider>
  );
};
