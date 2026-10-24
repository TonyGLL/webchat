import { useState, useEffect, useCallback } from 'react';
import { roomService, Room, CreateRoomData } from '@/services/roomService';
import { countIncomingMessage } from '@/lib/unread';
import { useSocket } from './useSocket';
import { useUser } from './useUser';

/** `activeRoomId` is the room being viewed: its messages are read as they arrive. */
export const useRoom = (activeRoomId: string | null = null) => {
  const socket = useSocket();
  const { profile } = useUser();
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchRooms = useCallback(async () => {
    try {
      const data = await roomService.getUserRooms();
      setRooms(data ?? []);
    } catch (error) {
      console.error('Failed to fetch rooms', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRooms();
  }, [fetchRooms]);

  // Opening a room clears its counter (optimistically) and tells the server.
  useEffect(() => {
    if (!activeRoomId) return;
    setRooms((prev) => prev.map((r) => (r.id === activeRoomId && r.unread_count ? { ...r, unread_count: 0 } : r)));
    roomService.markRead(activeRoomId).catch((error) => console.error('Failed to mark room as read', error));
  }, [activeRoomId]);

  // Live counters: count messages that arrive in rooms that are not open; keep the open one read.
  useEffect(() => {
    if (!socket) return;
    const onMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type !== 'new_message') return;
        const message = data.payload as { room_id: string; author_id: number };
        if (message.room_id === activeRoomId && message.author_id !== profile?.id) {
          roomService.markRead(activeRoomId).catch(() => undefined);
        }
        setRooms((prev) => countIncomingMessage(prev, message, activeRoomId, profile?.id));
      } catch {
        // not a message event we understand
      }
    };
    socket.addEventListener('message', onMessage);
    return () => socket.removeEventListener('message', onMessage);
  }, [socket, activeRoomId, profile?.id]);

  const createRoom = async (data: CreateRoomData) => {
    await roomService.createRoom(data);
    fetchRooms();
  };

  const leaveRoom = async (roomId: string) => {
    await roomService.leaveRoom(roomId);
    setRooms((prev) => prev.filter((r) => r.id !== roomId));
  };

  return { rooms, loading, createRoom, leaveRoom };
};
