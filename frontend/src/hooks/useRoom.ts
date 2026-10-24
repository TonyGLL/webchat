import { useState, useEffect, useCallback } from 'react';
import { roomService, Room, CreateRoomData } from '@/services/roomService';
import { countIncomingMessage } from '@/lib/unread';
import { useSocketEvent } from './useSocketEvents';
import { useUser } from './useUser';

/** `activeRoomId` is the room being viewed: its messages are read as they arrive. */
export const useRoom = (activeRoomId: string | null = null) => {
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
  useSocketEvent('role_changed', ({ room_id, user_id, role }) => {
    if (user_id !== profile?.id) return;
    setRooms((prev) => prev.map((r) => (r.id === room_id ? { ...r, my_role: role } : r)));
  });

  useSocketEvent('member_removed', ({ room_id, user_id }) => {
    if (user_id !== profile?.id) return;
    setRooms((prev) => prev.filter((r) => r.id !== room_id));
  });

  useSocketEvent('new_message', (message) => {
    if (message.room_id === activeRoomId && message.author_id !== profile?.id) {
      roomService.markRead(activeRoomId).catch(() => undefined);
    }
    setRooms((prev) => countIncomingMessage(prev, message, activeRoomId, profile?.id));
  });

  const createRoom = async (data: CreateRoomData) => {
    await roomService.createRoom(data);
    fetchRooms();
  };

  /** Opens (creating if needed) the one-to-one conversation with a user and returns its room id. */
  const openDirect = async (userId: number) => {
    const room = await roomService.openDirect(userId);
    await fetchRooms();
    return room.id;
  };

  /** Joins a public room and returns its id once it shows up in the list. */
  const joinRoom = async (roomId: string) => {
    await roomService.joinRoom(roomId);
    await fetchRooms();
    return roomId;
  };

  const leaveRoom = async (roomId: string) => {
    await roomService.leaveRoom(roomId);
    setRooms((prev) => prev.filter((r) => r.id !== roomId));
  };

  return { rooms, loading, createRoom, leaveRoom, openDirect, joinRoom, refreshRooms: fetchRooms };
};
