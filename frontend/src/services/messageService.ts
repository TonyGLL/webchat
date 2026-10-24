import { api } from '@/lib/api';

export interface IReaction {
  message_id?: string;
  user_id: number;
  emoji: string;
}

export interface IGetRoomMessages {
  id: string;
  content: string;
  author_id: number;
  room_id: string;
  created_at: string;
  reply_to_message_id?: string | null;
  pinned_at?: string | null;
  edited_at?: string | null;
  reactions?: IReaction[];
}

export const messageService = {
  /** Newest first. Pass `before` (id of the oldest loaded message) to get the previous page. */
  getMessages: async (roomId: string, limit: number, before?: string): Promise<IGetRoomMessages[] | null> => {
    const response = await api.get(`/messages/${roomId}/messages`, {
      params: { limit, before },
    });
    return response.data;
  },

  createMessage: async (data: { room_id: string; content: string; reply_to_message_id?: string }) => {
    const response = await api.post('/messages', data);
    return response.data;
  },

  /** Searches the messages of the user's rooms (or one room). Newest first. */
  search: async (query: string, roomId?: string, limit = 20): Promise<IGetRoomMessages[]> => {
    const response = await api.get('/messages/search', { params: { q: query, room_id: roomId, limit } });
    return response.data ?? [];
  },

  getPinned: async (roomId: string): Promise<IGetRoomMessages[]> => {
    const response = await api.get(`/messages/${roomId}/pinned`);
    return response.data ?? [];
  },

  setPinned: async (messageId: string, pinned: boolean): Promise<void> => {
    if (pinned) await api.put(`/messages/${messageId}/pin`);
    else await api.delete(`/messages/${messageId}/pin`);
  },

  deleteMessage: async (messageId: string): Promise<void> => {
    await api.delete(`/messages/${messageId}`);
  },

  addReaction: async (messageId: string, data: { emoji: string }) => {
    const response = await api.post(`/messages/${messageId}/reactions`, data);
    return response.data;
  },

  removeReaction: async (messageId: string, data: { emoji: string }) => {
    const response = await api.delete(`/messages/${messageId}/reactions`, { data });
    return response.data;
  },
};