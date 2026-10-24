import { useState, useEffect, useCallback, useRef } from 'react';
import { IGetRoomMessages, IReaction, messageService } from '@/services/messageService';
import { useSocket } from './useSocket';

// Events the backend broadcasts to a room (see backend/internal/message/application).
type WebSocketMessage =
  | { type: 'new_message'; payload: IGetRoomMessages }
  | { type: 'update_message'; payload: { message_id: string; content: string } }
  | { type: 'reaction_added'; payload: IReaction & { message_id: string } }
  | { type: 'reaction_removed'; payload: IReaction & { message_id: string } };

const PAGE_SIZE = 50;

/** A message the user sent that the server has not confirmed yet. */
export interface PendingMessage {
  tempId: string;
  room_id: string;
  content: string;
  status: 'sending' | 'failed';
}

export const useMessages = (roomId: string | null) => {
  const socket = useSocket();
  const [messages, setMessages] = useState<IGetRoomMessages[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [pending, setPending] = useState<PendingMessage[]>([]);

  const fetchMessages = useCallback(async () => {
    if (!roomId) {
      setMessages([]);
      return;
    }
    setLoading(true);
    try {
      const data = await messageService.getMessages(roomId, PAGE_SIZE);
      setMessages(data ?? []);
      setHasMore((data?.length ?? 0) === PAGE_SIZE);
    } catch (error) {
      console.error('Failed to fetch messages', error);
    } finally {
      setLoading(false);
    }
  }, [roomId]);

  useEffect(() => {
    fetchMessages();
    setPending([]); // unconfirmed messages belong to the room they were typed in
  }, [fetchMessages]);

  // Events sent while the connection was down are lost: reload the latest page after a reconnect.
  const wasConnected = useRef(false);
  const hadDrop = useRef(false);
  useEffect(() => {
    if (socket) {
      if (hadDrop.current) {
        hadDrop.current = false;
        fetchMessages();
      }
      wasConnected.current = true;
    } else if (wasConnected.current) {
      hadDrop.current = true;
    }
  }, [socket, fetchMessages]);

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

  // Messages are kept newest first, so the last one is the oldest loaded.
  const loadMore = async () => {
    const oldest = messages[messages.length - 1];
    if (!roomId || !oldest || loadingMore) return;
    setLoadingMore(true);
    try {
      const older = (await messageService.getMessages(roomId, PAGE_SIZE, oldest.id)) ?? [];
      setMessages((prev) => [...prev, ...older.filter((o) => !prev.some((m) => m.id === o.id))]);
      setHasMore(older.length === PAGE_SIZE);
    } catch (error) {
      console.error('Failed to load older messages', error);
    } finally {
      setLoadingMore(false);
    }
  };

  const deliver = async (item: PendingMessage) => {
    setPending((prev) => prev.map((p) => (p.tempId === item.tempId ? { ...p, status: 'sending' } : p)));
    try {
      const created = await messageService.createMessage({ room_id: item.room_id, content: item.content });
      setPending((prev) => prev.filter((p) => p.tempId !== item.tempId));
      // The websocket event usually arrives first; only add the message if it did not.
      if (created?.id && item.room_id === roomId) {
        setMessages((prev) => (prev.some((m) => m.id === created.id) ? prev : [created, ...prev]));
      }
    } catch (error) {
      console.error('Failed to send message', error);
      setPending((prev) => prev.map((p) => (p.tempId === item.tempId ? { ...p, status: 'failed' } : p)));
    }
  };

  /** Shows the message immediately as pending, then confirms or marks it failed. */
  const createMessage = async (data: { room_id: string; content: string }) => {
    const item: PendingMessage = { tempId: crypto.randomUUID(), ...data, status: 'sending' };
    setPending((prev) => [...prev, item]);
    await deliver(item);
  };

  const retryMessage = (tempId: string) => {
    const item = pending.find((p) => p.tempId === tempId);
    if (item && item.status === 'failed') void deliver(item);
  };

  const discardMessage = (tempId: string) => setPending((prev) => prev.filter((p) => p.tempId !== tempId));

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

  return { messages, pending, retryMessage, discardMessage, loading, loadingMore, hasMore, loadMore, createMessage, addReaction, removeReaction, setMessages };
};
