import { api } from '@/lib/api';
import type { User } from './authService';

export const userService = {
  getProfile: async (): Promise<User> => {
    const response = await api.get('/users/profile');
    return response.data;
  },

  deactivate: async () => {
    const response = await api.delete('/users/deactivate');
    return response.data;
  },
};
