import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

const api = axios.create({
  baseURL: API_URL,
});

api.interceptors.request.use((config) => {
  const tokens = localStorage.getItem('authTokens');
  if (tokens) {
    const { accessToken } = JSON.parse(tokens);
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

export const messageService = {
  getMessages: async (roomId: string, page: number, pageSize: number) => {
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