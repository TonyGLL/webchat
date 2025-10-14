import { useState, useEffect, useCallback } from 'react';
import { messageService } from '@/services/messageService';
import { Socket } from 'socket.io-client';

export const useMessages = (roomId: string | null, socket: Socket | null) => {
  const [messages, setMessages] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchMessages = useCallback(async () => {
    if (!roomId) {
      setMessages([]);
      return;
    }
    setLoading(true);
    try {
      const data = await messageService.getMessages(roomId, 1, 50);
      setMessages(data.reverse());
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

    const handleNewMessage = (newMessage: any) => {
      if (newMessage.room_id === roomId) {
        setMessages((prev) => [...prev, newMessage]);
      }
    };

    const handleReaction = (reaction: any) => {
        setMessages((prev) =>
            prev.map((msg) => {
                if (msg.id === reaction.message_id) {
                    return {
                        ...msg,
                        reactions: reaction.reactions,
                    };
                }
                return msg;
            })
        );
    };

    socket.on('new_message', handleNewMessage);
    socket.on('reaction_update', handleReaction);

    return () => {
      socket.off('new_message', handleNewMessage);
      socket.off('reaction_update', handleReaction);
    };
  }, [socket, roomId]);

  const createMessage = async (data: { room_id: string; content: string }) => {
    await messageService.createMessage(data);
  };

  const addReaction = async (messageId: string, emoji: string) => {
    await messageService.addReaction(messageId, { emoji });
  };

  const removeReaction = async (messageId: string, emoji: string) => {
    await messageService.removeReaction(messageId, { emoji });
  };

  return { messages, loading, createMessage, addReaction, removeReaction, setMessages };
};