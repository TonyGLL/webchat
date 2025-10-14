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

export const roomService = {
  createRoom: async (data: { name: string; is_public: boolean }) => {
    const response = await api.post('/rooms', data);
    return response.data;
  },

  getUserRooms: async () => {
    const response = await api.get('/rooms');
    return response.data;
  },

  joinRoom: async (roomId: string) => {
    const response = await api.post(`/rooms/${roomId}/join`);
    return response.data;
  },

  createInvite: async (roomId: string, data: { user_id: number; expires_at: string }) => {
    const response = await api.post(`/rooms/${roomId}/invites`, data);
    return response.data;
  },

  acceptInvite: async (token: string) => {
    const response = await api.post(`/rooms/invites/${token}/accept`);
    return response.data;
  },
};