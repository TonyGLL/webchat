'use client';

import { createContext, useEffect, useState } from 'react';
import { useAuth } from '@/hooks/useAuth';

interface IRawSocketContext {
  socket: WebSocket | null;
}

export const SocketContext = createContext<IRawSocketContext>({
  socket: null,
});

const WS_URL = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1')
  .replace(/^http/, 'ws') + '/ws/';

export const SocketProvider = ({ children }: { children: React.ReactNode }) => {
  const { tokens } = useAuth();
  const [socket, setSocket] = useState<WebSocket | null>(null);

  useEffect(() => {
    if (tokens?.accessToken) {
      const finalUrl = `${WS_URL}?token=${tokens.accessToken}`;
      const socketInstance = new WebSocket(finalUrl);

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

      return () => {
        socketInstance.close();
      };
    }
  }, [tokens]);

  return (
    <SocketContext.Provider value={{ socket }}>
      {children}
    </SocketContext.Provider>
  );
};
