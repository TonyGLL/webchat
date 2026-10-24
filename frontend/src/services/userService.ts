import { api } from '@/lib/api';
import type { User } from './authService';

export const userService = {
  getProfile: async (): Promise<User> => {
    const response = await api.get('/users/profile');
    return response.data;
  },

  exportData: async (): Promise<Blob> => {
    const response = await api.get('/users/export', { responseType: 'blob' });
    return response.data;
  },

  deleteAccount: async (password: string): Promise<void> => {
    await api.delete('/users/me', { data: { password } });
  },

  deactivate: async () => {
    const response = await api.delete('/users/deactivate');
    return response.data;
  },
};
