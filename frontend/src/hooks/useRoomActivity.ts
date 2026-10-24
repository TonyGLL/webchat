import { useCallback, useEffect, useRef, useState } from 'react';
import { roomService, RoomMember } from '@/services/roomService';
import { moderationService } from '@/services/moderationService';
import { RoomRole } from '@/lib/permissions';
import { useSocket } from './useSocket';
import { useSocketEvent } from './useSocketEvents';
import { useUser } from './useUser';
import { logError } from '@/lib/logError';

const TYPING_VISIBLE_MS = 4000;
const TYPING_SEND_INTERVAL_MS = 2000;

/** Members of a room with live online status, and who is typing right now. */
export const useRoomActivity = (roomId: string | null) => {
  const socket = useSocket();
  const { profile } = useUser();
  const [members, setMembers] = useState<RoomMember[]>([]);
  const [typingIds, setTypingIds] = useState<number[]>([]);
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>());
  const lastSent = useRef(0);

  const fetchMembers = useCallback(async () => {
    if (!roomId) {
      setMembers([]);
      return;
    }
    try {
      setMembers(await roomService.getMembers(roomId));
    } catch (error) {
      logError('Failed to fetch members', error);
    }
  }, [roomId]);

  // Reload when the room changes and whenever the connection (re)opens: presence events
  // sent while it was down are lost.
  useEffect(() => {
    fetchMembers();
  }, [fetchMembers, socket]);

  useEffect(() => {
    const pending = timers.current;
    return () => {
      pending.forEach(clearTimeout);
      pending.clear();
      setTypingIds([]);
    };
  }, [roomId]);

  useSocketEvent('presence', ({ user_id, online }) => {
    setMembers((prev) => prev.map((m) => (m.user_id === user_id ? { ...m, online } : m)));
  });

  useSocketEvent('role_changed', (payload) => {
    if (payload.room_id !== roomId) return;
    setMembers((prev) => prev.map((m) => (m.user_id === payload.user_id ? { ...m, role: payload.role } : m)));
  });

  useSocketEvent('member_removed', (payload) => {
    if (payload.room_id !== roomId) return;
    setMembers((prev) => prev.filter((m) => m.user_id !== payload.user_id));
  });

  useSocketEvent('typing', (payload) => {
    if (payload.room_id !== roomId || payload.user_id === profile?.id) return;
    const pending = timers.current;
    const userId = payload.user_id;
    clearTimeout(pending.get(userId));
    setTypingIds((prev) => (prev.includes(userId) ? prev : [...prev, userId]));
    pending.set(
      userId,
      setTimeout(() => {
        pending.delete(userId);
        setTypingIds((prev) => prev.filter((id) => id !== userId));
      }, TYPING_VISIBLE_MS)
    );
  });

  /** Tell the room we are typing (at most once every couple of seconds). */
  const notifyTyping = useCallback(() => {
    if (!socket || !roomId || socket.readyState !== WebSocket.OPEN) return;
    const now = Date.now();
    if (now - lastSent.current < TYPING_SEND_INTERVAL_MS) return;
    lastSent.current = now;
    socket.send(JSON.stringify({ type: 'typing', room_id: roomId }));
  }, [socket, roomId]);

  const changeRole = async (userId: number, role: RoomRole) => {
    if (!roomId) return;
    await moderationService.changeRole(roomId, userId, role);
    setMembers((prev) => prev.map((m) => (m.user_id === userId ? { ...m, role } : m)));
  };

  const kickMember = async (userId: number) => {
    if (!roomId) return;
    await moderationService.kick(roomId, userId);
    setMembers((prev) => prev.filter((m) => m.user_id !== userId));
  };

  const banMember = async (userId: number, reason: string, duration: string) => {
    if (!roomId) return;
    await moderationService.ban(roomId, userId, reason, duration.trim() || undefined);
    setMembers((prev) => prev.filter((m) => m.user_id !== userId));
  };

  const typingNames = typingIds
    .map((id) => members.find((m) => m.user_id === id))
    .filter((m): m is RoomMember => Boolean(m))
    .map((m) => m.name);

  return { members, typingNames, notifyTyping, changeRole, kickMember, banMember };
};
