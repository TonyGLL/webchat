import { useState, useEffect, useCallback } from 'react';
import { IGetRoomMessages, messageService } from '@/services/messageService';
import { useSocket } from './useSocket';

// Assuming a structure for messages sent from the server
interface WebSocketMessage {
  type: 'new_message' | 'update_message' | 'add_reaction' | 'update_reaction' | 'remove_reaction';
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

    const handleMessage = (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);

        if (message.type === 'new_message' && message.payload.room_id === roomId) {
          setMessages((prev) => [message.payload, ...prev]);
        }

        if (message.type === 'update_message') {
          setMessages((prev) =>
            prev.map((msg) => {
              if (msg.id === message.payload.message_id) {
                return {
                  ...msg,
                  content: message.payload.content,
                  is_edited: true,
                };
              }
              return msg;
            })
          );
        }

        if (message.type === 'add_reaction') {
          setMessages((prev) =>
            prev.map((msg) => {
              if (msg.id === message.payload.message_id) {
                const newReactions = msg.reactions ? [...msg.reactions, message.payload] : [message.payload];
                return {
                  ...msg,
                  reactions: newReactions,
                };
              }
              return msg;
            })
          );
        }

        if (message.type === 'remove_reaction') {
          setMessages((prev) =>
            prev.map((msg) => {
              if (msg.id === message.payload.message_id) {
                const newReactions = msg.reactions?.filter(
                  (r) => !(r.emoji === message.payload.emoji && r.user_id === message.payload.user_id)
                );
                return {
                  ...msg,
                  reactions: newReactions,
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

    socket.addEventListener('message', handleMessage);

    return () => {
      socket.removeEventListener('message', handleMessage);
    };
  }, [socket, roomId]);

  const createMessage = async (data: { room_id: string; content: string }) => {
    try {
      await messageService.createMessage({ room_id: data.room_id, content: data.content });
    } catch (error) {
      console.error('Failed to send message', error);
    }
  };

  const addReaction = async (messageId: string, emoji: string) => {
    try {
      await messageService.addReaction(messageId, { emoji });
    } catch (error) {
      console.error('Failed to add reaction', error);
    }
  };

  const removeReaction = async (messageId: string, emoji: string) => {
    try {
      await messageService.removeReaction(messageId, { emoji });
    } catch (error) {
      console.error('Failed to remove reaction', error);
    }
  };

  return { messages, loading, createMessage, addReaction, removeReaction, setMessages };
};