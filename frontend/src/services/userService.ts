import { api } from '@/lib/api';
import type { User } from './authService';

export interface DirectoryUser {
  id: number;
  username: string;
  name: string;
  last_name: string | null;
  avatar_url: string | null;
  online: boolean;
}

export const userService = {
  /** Everyone else, online people first; `q` filters by name or username. */
  getDirectory: async (q = ''): Promise<DirectoryUser[]> => {
    const response = await api.get('/users/directory', { params: { q: q || undefined } });
    return response.data ?? [];
  },

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

  deactivate: async (password: string) => {
    const response = await api.delete('/users/deactivate', { data: { password } });
    return response.data;
  },
};
