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

export const userService = {
  getProfile: async () => {
    const response = await api.get('/users/profile');
    return response.data;
  },

  deactivate: async () => {
    const response = await api.post('/users/deactivate');
    return response.data;
  },
};