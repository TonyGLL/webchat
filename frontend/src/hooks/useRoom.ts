import { useState, useEffect, useCallback, useRef } from 'react';
import { roomService, Room, CreateRoomData } from '@/services/roomService';
import { countIncomingMessage } from '@/lib/unread';
import { useSocketEvent } from './useSocketEvents';
import { useUser } from './useUser';
import { logError } from '@/lib/logError';
import { createTrailingThrottle } from '@/lib/trailingThrottle';

/** Every viewer would otherwise write `last_read_at` once per incoming message: at most once per window. */
const MARK_READ_MS = 4000;

/** `activeRoomId` is the room being viewed: its messages are read as they arrive. */
export const useRoom = (activeRoomId: string | null = null) => {
  const { profile } = useUser();
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  // Bumped whenever the user's own rooms change after the first load (joined, created, left, removed);
  // room discovery refetches on it instead of on `rooms.length`, which also moved on the first load.
  const [membershipVersion, setMembershipVersion] = useState(0);
  const loadedOnce = useRef(false);

  const fetchRooms = useCallback(async () => {
    try {
      const data = await roomService.getUserRooms();
      setRooms(data ?? []);
      if (loadedOnce.current) setMembershipVersion((v) => v + 1);
      loadedOnce.current = true;
    } catch (error) {
      logError('Failed to fetch rooms', error);
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
    setRooms((prev) =>
      prev.some((r) => r.id === activeRoomId && r.unread_count) ? prev.map((r) => (r.id === activeRoomId && r.unread_count ? { ...r, unread_count: 0 } : r)) : prev
    );
    roomService.markRead(activeRoomId).catch((error) => logError('Failed to mark room as read', error));
  }, [activeRoomId]);

  // Messages that arrive while the room is open are marked read in one trailing request per window,
  // and only while the tab is visible; coming back to the tab marks what arrived meanwhile.
  const markReadSoon = useRef<ReturnType<typeof createTrailingThrottle> | null>(null);
  useEffect(() => {
    if (!activeRoomId) return;
    const run = () => {
      if (document.visibilityState !== 'visible') return;
      roomService.markRead(activeRoomId).catch(() => undefined);
    };
    const throttle = createTrailingThrottle(run, MARK_READ_MS);
    markReadSoon.current = throttle;
    const onVisible = () => throttle.call();
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      document.removeEventListener('visibilitychange', onVisible);
      throttle.cancel();
      markReadSoon.current = null;
    };
  }, [activeRoomId]);

  // Live counters: count messages that arrive in rooms that are not open; keep the open one read.
  useSocketEvent('role_changed', ({ room_id, user_id, role }) => {
    if (user_id !== profile?.id) return;
    setRooms((prev) => (prev.some((r) => r.id === room_id && r.my_role !== role) ? prev.map((r) => (r.id === room_id ? { ...r, my_role: role } : r)) : prev));
  });

  useSocketEvent('member_removed', ({ room_id, user_id }) => {
    if (user_id !== profile?.id) return;
    setRooms((prev) => (prev.some((r) => r.id === room_id) ? prev.filter((r) => r.id !== room_id) : prev));
    setMembershipVersion((v) => v + 1);
  });

  useSocketEvent('new_message', (message) => {
    if (message.room_id === activeRoomId && message.author_id !== profile?.id) {
      markReadSoon.current?.call();
    }
    setRooms((prev) => countIncomingMessage(prev, message, activeRoomId, profile?.id));
  });

  const createRoom = useCallback(async (data: CreateRoomData) => {
    await roomService.createRoom(data);
    fetchRooms();
  }, [fetchRooms]);

  /** Opens (creating if needed) the one-to-one conversation with a user and returns its room id. */
  const openDirect = useCallback(async (userId: number) => {
    const room = await roomService.openDirect(userId);
    await fetchRooms();
    return room.id;
  }, [fetchRooms]);

  /** Joins a public room and returns its id once it shows up in the list. */
  const joinRoom = useCallback(async (roomId: string) => {
    await roomService.joinRoom(roomId);
    await fetchRooms();
    return roomId;
  }, [fetchRooms]);

  const leaveRoom = useCallback(async (roomId: string) => {
    await roomService.leaveRoom(roomId);
    setRooms((prev) => prev.filter((r) => r.id !== roomId));
    setMembershipVersion((v) => v + 1);
  }, []);

  return { rooms, loading, membershipVersion, createRoom, leaveRoom, openDirect, joinRoom, refreshRooms: fetchRooms };
};
