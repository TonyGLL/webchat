import { api } from '@/lib/api';
import { RoomRole } from '@/lib/permissions';

export interface Room {
  id: string;
  name: string;
  topic: string | null;
  is_private: boolean;
  /** One-to-one conversation; `name` is then the other person's name. */
  is_direct: boolean;
  owner_id: number;
  created_at: string;
  /** Messages from other people since the user last opened the room. */
  unread_count: number;
  my_role: RoomRole;
}

export interface RoomMember {
  user_id: number;
  username: string;
  name: string;
  last_name: string | null;
  avatar_url: string | null;
  role: RoomRole;
  online: boolean;
}

export interface CreateRoomData {
  name: string;
  is_private: boolean;
  topic: string;
}

export interface Invite {
  id: string;
  room_id: string;
  code: string;
  expires_at: string;
}

export const roomService = {
  createRoom: async (data: CreateRoomData): Promise<Room> => {
    const response = await api.post('/rooms', data);
    return response.data;
  },

  getUserRooms: async (): Promise<Room[] | null> => {
    const response = await api.get('/rooms');
    return response.data;
  },

  openDirect: async (userId: number): Promise<Room> => {
    const response = await api.post(`/rooms/direct/${userId}`);
    return response.data;
  },

  getMembers: async (roomId: string): Promise<RoomMember[]> => {
    const response = await api.get(`/rooms/members/${roomId}`);
    return response.data ?? [];
  },

  leaveRoom: async (roomId: string): Promise<void> => {
    await api.post(`/rooms/leave/${roomId}`);
  },

  markRead: async (roomId: string): Promise<void> => {
    await api.post(`/rooms/read/${roomId}`);
  },

  joinRoom: async (roomId: string) => {
    const response = await api.post(`/rooms/join/${roomId}`);
    return response.data;
  },

  /** `expiresIn` is a Go duration such as "24h". Only the room owner or an admin can invite. */
  createInvite: async (roomId: string, data: { expires_in: string; email?: string }): Promise<Invite> => {
    const response = await api.post(`/rooms/invite/${roomId}`, data);
    return response.data;
  },

  acceptInvite: async (code: string): Promise<Invite> => {
    const response = await api.post(`/rooms/accept-invite/${code}`);
    return response.data;
  },
};
