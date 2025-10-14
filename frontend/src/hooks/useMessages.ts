import { useState, useEffect, useCallback } from 'react';
import { IGetRoomMessages, messageService } from '@/services/messageService';
import { useSocket } from './useSocket';

// Assuming a structure for messages sent from the server
interface WebSocketMessage {
  type: 'new_message' | 'reaction_update';
  payload: any;
}

export const useMessages = (roomId: string | null) => {
  const socket = useSocket();
  const [messages, setMessages] = useState<IGetRoomMessages[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchMessages = useCallback(async () => {
    if (!roomId) {
      setMessages([]);
      return;
    }
    setLoading(true);
    try {
      const data = await messageService.getMessages(roomId, 1, 50);
      setMessages(data);
    } catch (error) {
      console.error('Failed to fetch messages', error);
    } finally {
      setLoading(false);
    }
  }, [roomId]);

  useEffect(() => {
    fetchMessages();
  }, [fetchMessages]);

  useEffect(() => {
    if (!socket) return;

    socket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);

        if (message.type === 'new_message' && message.payload.room_id === roomId) {
          setMessages((prev) => [...prev, message.payload]);
        }

        if (message.type === 'reaction_update') {
          setMessages((prev) =>
            prev.map((msg) => {
              if (msg.id === message.payload.message_id) {
                return {
                  ...msg,
                  reactions: message.payload.reactions,
                };
              }
              return msg;
            })
          );
        }
      } catch (error) {
        console.error('Error parsing incoming socket message:', error);
      }
    };

  }, [socket, roomId]);

  const sendMessage = (type: string, payload: any) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type, payload }));
    } else {
      console.error('Socket is not open. Cannot send message.');
    }
  };

  const createMessage = (data: { room_id: string; content: string }) => {
    sendMessage('create_message', data);
  };

  const addReaction = (messageId: string, emoji: string) => {
    sendMessage('add_reaction', { message_id: messageId, emoji });
  };

  const removeReaction = (messageId: string, emoji: string) => {
    sendMessage('remove_reaction', { message_id: messageId, emoji });
  };

  return { messages, loading, createMessage, addReaction, removeReaction, setMessages };
};