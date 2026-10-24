'use client';

import { createContext, useEffect, useState } from 'react';
import axios from 'axios';
import { useAuth } from '@/hooks/useAuth';

interface IRawSocketContext {
  socket: WebSocket | null;
}

export const SocketContext = createContext<IRawSocketContext>({
  socket: null,
});

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081/api/v1';
const WS_URL = API_URL.replace(/^http/, 'ws') + '/ws';

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { tokens } = useAuth();
  const [socket, setSocket] = useState<WebSocket | null>(null);

  useEffect(() => {
    if (!tokens?.accessToken) return;

    let cancelled = false;
    let socketInstance: WebSocket | null = null;

    // The JWT never goes in the URL: exchange it for a short-lived, single-use ticket.
    axios
      .post<{ ticket: string }>(`${API_URL}/ws-ticket`, null, {
        headers: { Authorization: `Bearer ${tokens.accessToken}` },
      })
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
  }, [tokens]);

  return (
    <SocketContext.Provider value={{ socket }}>
      {children}
    </SocketContext.Provider>
  );
};
