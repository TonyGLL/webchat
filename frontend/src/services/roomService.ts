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

/** A room in the discovery list: visible to everyone, readable only by members. */
export interface RoomSummary {
  id: string;
  name: string;
  topic: string | null;
  is_private: boolean;
  member_count: number;
  is_member: boolean;
}

/** An invitation addressed to the current user. */
export interface Invitation {
  id: string;
  room_id: string;
  room_name: string;
  inviter_id: number;
  inviter_name: string;
  invitee_id: number;
  created_at: string;
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

const MEMBERS_PAGE = 200;
const DISCOVER_PAGE = 30;

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

  /** Every member, fetched page by page (the API returns at most 200 per request). */
  getMembers: async (roomId: string): Promise<RoomMember[]> => {
    const all: RoomMember[] = [];
    for (;;) {
      const response = await api.get(`/rooms/members/${roomId}`, { params: { limit: MEMBERS_PAGE, offset: all.length } });
      const page: RoomMember[] = response.data ?? [];
      all.push(...page);
      if (page.length < MEMBERS_PAGE) return all;
    }
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

  /** One page of rooms after the room `after` (the last one received); `hasMore` says to ask again. */
  discoverRooms: async (after?: string): Promise<{ rooms: RoomSummary[]; hasMore: boolean }> => {
    const response = await api.get('/rooms/discover', { params: { limit: DISCOVER_PAGE, after } });
    const rooms: RoomSummary[] = response.data ?? [];
    return { rooms, hasMore: rooms.length === DISCOVER_PAGE };
  },

  /** Sends a personal invitation; only the owner or an admin of the room can. */
  inviteUser: async (roomId: string, userId: number): Promise<Invitation> => {
    const response = await api.post(`/rooms/invite-user/${roomId}`, { user_id: userId });
    return response.data;
  },

  getInvitations: async (): Promise<Invitation[]> => {
    const response = await api.get('/rooms/invitations');
    return response.data ?? [];
  },

  acceptInvitation: async (id: string): Promise<void> => {
    await api.post(`/rooms/invitations/${id}/accept`);
  },

  declineInvitation: async (id: string): Promise<void> => {
    await api.post(`/rooms/invitations/${id}/decline`);
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
