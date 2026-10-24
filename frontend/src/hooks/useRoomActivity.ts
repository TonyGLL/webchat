import { useCallback, useEffect, useRef, useState } from 'react';
import { roomService, RoomMember } from '@/services/roomService';
import { moderationService } from '@/services/moderationService';
import { RoomRole } from '@/lib/permissions';
import { useSocket } from './useSocket';
import { useUser } from './useUser';

const TYPING_VISIBLE_MS = 4000;
const TYPING_SEND_INTERVAL_MS = 2000;

type ActivityEvent =
  | { type: 'presence'; payload: { user_id: number; online: boolean } }
  | { type: 'role_changed'; payload: { room_id: string; user_id: number; role: RoomRole } }
  | { type: 'member_removed'; payload: { room_id: string; user_id: number } }
  | { type: 'typing'; payload: { room_id: string; user_id: number } };

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
      console.error('Failed to fetch members', error);
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

  useEffect(() => {
    if (!socket) return;
    const pending = timers.current;

    const onMessage = (event: MessageEvent) => {
      let data: ActivityEvent;
      try {
        data = JSON.parse(event.data);
      } catch {
        return;
      }

      if (data.type === 'presence') {
        const { user_id, online } = data.payload;
        setMembers((prev) => prev.map((m) => (m.user_id === user_id ? { ...m, online } : m)));
      } else if (data.type === 'role_changed' && data.payload.room_id === roomId) {
        const { user_id, role } = data.payload;
        setMembers((prev) => prev.map((m) => (m.user_id === user_id ? { ...m, role } : m)));
      } else if (data.type === 'member_removed' && data.payload.room_id === roomId) {
        setMembers((prev) => prev.filter((m) => m.user_id !== data.payload.user_id));
      } else if (data.type === 'typing' && data.payload.room_id === roomId && data.payload.user_id !== profile?.id) {
        const userId = data.payload.user_id;
        clearTimeout(pending.get(userId));
        setTypingIds((prev) => (prev.includes(userId) ? prev : [...prev, userId]));
        pending.set(
          userId,
          setTimeout(() => {
            pending.delete(userId);
            setTypingIds((prev) => prev.filter((id) => id !== userId));
          }, TYPING_VISIBLE_MS)
        );
      }
    };

    socket.addEventListener('message', onMessage);
    return () => socket.removeEventListener('message', onMessage);
  }, [socket, roomId, profile?.id]);

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
