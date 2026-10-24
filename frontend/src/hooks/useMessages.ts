import { useState, useEffect, useCallback } from 'react';
import { IGetRoomMessages, IReaction, messageService } from '@/services/messageService';
import { useSocket } from './useSocket';

// Events the backend broadcasts to a room (see backend/internal/message/application).
type WebSocketMessage =
  | { type: 'new_message'; payload: IGetRoomMessages }
  | { type: 'update_message'; payload: { message_id: string; content: string } }
  | { type: 'reaction_added'; payload: IReaction & { message_id: string } }
  | { type: 'reaction_removed'; payload: IReaction & { message_id: string } };

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
      setMessages(data ?? []);
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

        switch (message.type) {
          case 'new_message':
            if (message.payload.room_id === roomId) {
              setMessages((prev) =>
                prev.some((m) => m.id === message.payload.id) ? prev : [message.payload, ...prev]
              );
            }
            break;

          case 'update_message':
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === message.payload.message_id
                  ? { ...msg, content: message.payload.content, edited_at: new Date().toISOString() }
                  : msg
              )
            );
            break;

          case 'reaction_added':
            setMessages((prev) =>
              prev.map((msg) => {
                if (msg.id !== message.payload.message_id) return msg;
                const exists = msg.reactions?.some(
                  (r) => r.user_id === message.payload.user_id && r.emoji === message.payload.emoji
                );
                return exists ? msg : { ...msg, reactions: [...(msg.reactions ?? []), message.payload] };
              })
            );
            break;

          case 'reaction_removed':
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === message.payload.message_id
                  ? {
                      ...msg,
                      reactions: msg.reactions?.filter(
                        (r) => !(r.emoji === message.payload.emoji && r.user_id === message.payload.user_id)
                      ),
                    }
                  : msg
              )
            );
            break;
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
