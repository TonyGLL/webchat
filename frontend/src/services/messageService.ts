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
  edited_at?: string | null;
  reactions?: IReaction[];
}

export const messageService = {
  getMessages: async (roomId: string, page: number, pageSize: number): Promise<IGetRoomMessages[] | null> => {
    const response = await api.get(`/messages/${roomId}/messages`, {
      params: { page, pageSize },
    });
    return response.data;
  },

  createMessage: async (data: { room_id: string; content: string }) => {
    const response = await api.post('/messages', data);
    return response.data;
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