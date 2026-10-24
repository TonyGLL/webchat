import { useState, useEffect, useCallback, useRef } from 'react';
import { IGetRoomMessages, messageService } from '@/services/messageService';
import { useSocket } from './useSocket';
import { useSocketEvent } from './useSocketEvents';
import { logError } from '@/lib/logError';

const PAGE_SIZE = 50;

/** A message the user sent that the server has not confirmed yet. */
export interface PendingMessage {
  tempId: string;
  room_id: string;
  content: string;
  reply_to_message_id?: string;
  status: 'sending' | 'failed';
}

export const useMessages = (roomId: string | null) => {
  const socket = useSocket();
  const [messages, setMessages] = useState<IGetRoomMessages[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [pinned, setPinned] = useState<IGetRoomMessages[]>([]);
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
      logError('Failed to fetch messages', error);
    } finally {
      setLoading(false);
    }
  }, [roomId]);

  // Keep the pin state shown on messages already loaded in sync.
  const markPinned = useCallback((messageId: string, isPinned: boolean) => {
    setMessages((prev) =>
      prev.map((m) => (m.id === messageId ? { ...m, pinned_at: isPinned ? new Date().toISOString() : null } : m))
    );
  }, []);

  const fetchPinned = useCallback(async () => {
    if (!roomId) {
      setPinned([]);
      return;
    }
    try {
      setPinned(await messageService.getPinned(roomId));
    } catch (error) {
      logError('Failed to fetch pinned messages', error);
    }
  }, [roomId]);

  useEffect(() => {
    fetchPinned();
  }, [fetchPinned]);

  useEffect(() => {
    fetchMessages();
  }, [fetchMessages]);

  // Events sent while the connection was down are lost: reload the latest page after a reconnect.
  const wasConnected = useRef(false);
  const hadDrop = useRef(false);
  useEffect(() => {
    if (socket) {
      if (hadDrop.current) {
        hadDrop.current = false;
        fetchMessages();
        fetchPinned();
      }
      wasConnected.current = true;
    } else if (wasConnected.current) {
      hadDrop.current = true;
    }
  }, [socket, fetchMessages, fetchPinned]);

  useSocketEvent('new_message', (message) => {
    if (message.room_id === roomId) {
      setMessages((prev) => (prev.some((m) => m.id === message.id) ? prev : [message, ...prev]));
    }
  });

  useSocketEvent('pin_message', ({ message_id, pinned: isPinned }) => {
    markPinned(message_id, isPinned);
    fetchPinned();
  });

  useSocketEvent('delete_message', ({ message_id }) => {
    setMessages((prev) => prev.filter((msg) => msg.id !== message_id));
    setPinned((prev) => prev.filter((msg) => msg.id !== message_id));
  });

  useSocketEvent('update_message', ({ message_id, content }) => {
    setMessages((prev) =>
      prev.map((msg) => (msg.id === message_id ? { ...msg, content, edited_at: new Date().toISOString() } : msg))
    );
  });

  useSocketEvent('reaction_added', (reaction) => {
    setMessages((prev) =>
      prev.map((msg) => {
        if (msg.id !== reaction.message_id) return msg;
        const exists = msg.reactions?.some((r) => r.user_id === reaction.user_id && r.emoji === reaction.emoji);
        return exists ? msg : { ...msg, reactions: [...(msg.reactions ?? []), reaction] };
      })
    );
  });

  useSocketEvent('reaction_removed', (reaction) => {
    setMessages((prev) =>
      prev.map((msg) =>
        msg.id === reaction.message_id
          ? { ...msg, reactions: msg.reactions?.filter((r) => !(r.emoji === reaction.emoji && r.user_id === reaction.user_id)) }
          : msg
      )
    );
  });

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
      logError('Failed to load older messages', error);
    } finally {
      setLoadingMore(false);
    }
  };

  const deliver = async (item: PendingMessage) => {
    setPending((prev) => prev.map((p) => (p.tempId === item.tempId ? { ...p, status: 'sending' } : p)));
    try {
      const created = await messageService.createMessage({
        room_id: item.room_id,
        content: item.content,
        reply_to_message_id: item.reply_to_message_id,
      });
      setPending((prev) => prev.filter((p) => p.tempId !== item.tempId));
      // The websocket event usually arrives first; only add the message if it did not.
      if (created?.id && item.room_id === roomId) {
        setMessages((prev) => (prev.some((m) => m.id === created.id) ? prev : [created, ...prev]));
      }
    } catch (error) {
      logError('Failed to send message', error);
      setPending((prev) => prev.map((p) => (p.tempId === item.tempId ? { ...p, status: 'failed' } : p)));
    }
  };

  /** Shows the message immediately as pending, then confirms or marks it failed. */
  const createMessage = async (data: { room_id: string; content: string; reply_to_message_id?: string }) => {
    const item: PendingMessage = { tempId: crypto.randomUUID(), ...data, status: 'sending' };
    setPending((prev) => [...prev, item]);
    await deliver(item);
  };

  const retryMessage = (tempId: string) => {
    const item = pending.find((p) => p.tempId === tempId);
    if (item && item.status === 'failed') void deliver(item);
  };

  const discardMessage = (tempId: string) => setPending((prev) => prev.filter((p) => p.tempId !== tempId));

  const togglePin = async (messageId: string, pin: boolean) => {
    try {
      await messageService.setPinned(messageId, pin);
      markPinned(messageId, pin);
      await fetchPinned();
    } catch (error) {
      logError('Failed to update pin', error);
    }
  };

  const deleteMessage = async (messageId: string) => {
    try {
      await messageService.deleteMessage(messageId);
      setMessages((prev) => prev.filter((msg) => msg.id !== messageId));
    } catch (error) {
      logError('Failed to delete message', error);
    }
  };

  const addReaction = async (messageId: string, emoji: string) => {
    try {
      await messageService.addReaction(messageId, { emoji });
    } catch (error) {
      logError('Failed to add reaction', error);
      // The caller tells the user (the server refuses past its reaction limits).
      throw error;
    }
  };

  const removeReaction = async (messageId: string, emoji: string) => {
    try {
      await messageService.removeReaction(messageId, { emoji });
    } catch (error) {
      logError('Failed to remove reaction', error);
    }
  };

  return { messages, pinned, togglePin, pending, retryMessage, discardMessage, loading, loadingMore, hasMore, loadMore, createMessage, deleteMessage, addReaction, removeReaction, setMessages };
};
